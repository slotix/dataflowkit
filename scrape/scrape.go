package scrape

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/segmentio/ksuid"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/storage"
	"github.com/slotix/dataflowkit/utils"
	"github.com/spf13/viper"
	"github.com/temoto/robotstxt"
	"go.uber.org/zap"
)

// NewTask creates new task to parse fetched page following the rules from Payload.
func NewTask(p Payload) *Task {
	//init other fields
	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	p.PayloadMD5 = string(utils.GenerateCRC32(utils.GenerateMD5(data)))

	delay := time.Duration(viper.GetInt("FETCH_DELAY")) * time.Millisecond
	p.FetchDelay = &delay
	rand := viper.GetBool("RANDOMIZE_FETCH_DELAY")
	p.RandomizeFetchDelay = &rand

	if p.Paginator != nil {
		if p.Paginator.MaxPages == 0 {
			p.Paginator.MaxPages = viper.GetInt("MAX_PAGES")
		}
		if p.Paginator.Type != "next" {
			p.Request.Actions = fmt.Sprintf(`{"paginate":{"maxpage":%d, "element":"%s"}}`, p.Paginator.MaxPages, p.Paginator.Selector)
			p.Request.Type = "chrome"
		}
	}
	if p.PaginateResults == nil {
		pag := viper.GetBool("PAGINATE_RESULTS")
		p.PaginateResults = &pag
	}
	//https://blog.kowalczyk.info/article/JyRZ/generating-good-random-and-unique-ids-in-go.html
	id := ksuid.New()
	//tQueue := make(chan *Scraper, 100)
	storageType := viper.GetString("STORAGE_TYPE")
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{
		ID:      id.String(),
		Payload: p,
		//Errors:       []error{},
		Robots:       make(map[string]*robotstxt.RobotsData),
		Parsed:       false,
		requestCount: make(map[string]uint32),
		storage:      storage.NewStore(storageType),
		mx:           &sync.Mutex{},
		jobDone:      sync.WaitGroup{},
		ctx:          ctx,
		Cancel:       cancel,
		statePool:    make(map[string]scrapeState),
	}

}

// Parse specified payload.
func (task *Task) Parse() (io.ReadCloser, error) {
	begin := time.Now()
	scraper, err := task.Payload.newScraper("initial")
	if err != nil {
		return nil, err
	}
	//scrape request and return results.

	task.fetchChannel = make(chan *fetchInfo, viper.GetInt("FETCH_CHANNEL_SIZE"))
	for i := 0; i < viper.GetInt("FETCH_WORKER_NUM"); i++ {
		go task.fetchWorker()
	}
	task.blockChannel = make(chan *blockStruct, viper.GetInt("BLOCK_CHANNEL_SIZE"))
	for i := 0; i < viper.GetInt("BLOCK_WORKER_NUM"); i++ {
		go task.blockWorker(task.blockChannel)
	}

	defer task.closeTask()

	uid := string(utils.GenerateCRC32([]byte(task.Payload.PayloadMD5)))
	tw := taskWorker{
		currentPageNum:  0,
		scraper:         scraper,
		UID:             uid,
		useBlockCounter: false,
		keys:            make(map[int][]int),
	}
	task.jobDone.Add(1)
	_, err = task.scrape(&tw)
	task.jobDone.Wait()
	switch e := err.(type) {
	//don't try to fetch a page with chrome fetcher if forbiddenByRobots error returned
	case errs.Error:
		if e.Status() == http.StatusForbidden {
			return nil, e
		}
	case errs.Cancel:
		return nil, e
	}
	if !task.Parsed {
		logger.Info("Failed to scrape with base fetcher. Reinitializing to scrape with Chrome fetcher.")
		if task.Payload.Request.Type == "chrome" {
			return nil, err
		}
		task.Payload.Request.Type = "chrome"
		scraper.Request.Type = "chrome"
		delete(task.statePool, tw.UID)
		task.jobDone.Add(1)
		_, err = task.scrape(&tw)
		task.jobDone.Wait()
		if err != nil {
			return nil, err
		}

		if !task.Parsed {
			return nil, err
		}
	}

	for k := range tw.keys {
		sort.Slice(tw.keys[k], func(i, j int) bool { return tw.keys[k][i] < tw.keys[k][j] })
	}

	j, err := json.Marshal(tw.keys)
	if err != nil {
		return nil, err
	}
	err = task.storage.Write(storage.Record{
		Type:    storage.INTERMEDIATE,
		Key:     string(uid),
		Value:   j,
		ExpTime: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot write parse results key map. %s", err.Error())
	}

	task.storage.Close()

	w := bufio.NewWriter(os.Stdout)
	task.printLog(w)
	w.Flush()
	var e encoder
	switch strings.ToLower(task.Payload.Format) {
	case "csv":
		e = CSVEncoder{
			comma:     ",",
			partNames: scraper.partNames(),
		}
	case "json":
		e = JSONEncoder{
			//		paginateResults: *task.Payload.PaginateResults,
		}
	// TODO: implemetation ndJSON payload
	case "jsonl":
		e = JSONEncoder{
			JSONL: true,
		}
	case "xml":
		e = XMLEncoder{}
	case "xls":
		e = XLSXEncoder{
			partNames: scraper.partNames(),
		}
	default:
		return nil, errors.New("invalid output format specified")
	}
	r, err := EncodeToFile(task.ctx, &e, encodeInfo{
		payloadMD5: string(uid),
		extension:  task.Payload.Format,
		compressor: strings.ToLower(task.Payload.Compressor),
		// compressLevel: 0,
	})
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{
		"Task ID":     task.ID,
		"Requests":    task.requestCount,
		"Responses":   task.responseCount,
		"Output file": string(r),
		"Took":        time.Since(begin).String(),
	}
	parseResults, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(parseResults)), err
}

// Create a new scraper with the provided configuration.
func (p Payload) newScraper(reqType string) (*Scraper, error) {
	parts, err := p.fields2parts()
	if err != nil {
		return nil, err
	}
	var paginator paginate.Paginator
	paginatorType := ""
	if p.Paginator == nil {
		paginator = &dummyPaginator{}

	} else {
		paginator = paginate.BySelector(p.Paginator.Selector, p.Paginator.Attribute)
		paginatorType = p.Paginator.Type
	}

	selectors, err := p.selectors()
	if err != nil {
		return nil, err
	}

	var dividePageFunc DividePageFunc

	dividePageFunc = DividePageByIntersection(selectors)

	scraper := &Scraper{
		Request:       p.Request,
		reqType:       reqType,
		DividePage:    dividePageFunc,
		Parts:         parts,
		Paginator:     paginator,
		IsPath:        p.IsPath,
		paginatorType: paginatorType,
	}

	// All set!
	return scraper, nil
}

//fields2parts converts payload []field to []scrape.Part
func (p Payload) fields2parts() ([]Part, error) {
	parts := []Part{}
	//Payload fields
	for _, f := range p.Fields {
		if p.IsPath && !utils.ArrayContains(f.Extractor.Types, "path") {
			continue
		}
		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params
		}

		for _, t := range f.Extractor.Types {
			part := Part{
				Name:     f.Name + "_" + t,
				Selector: f.Selector,
			}
			e, err := p.newExtractor(t, &f, &part, &params)
			if err != nil {
				return nil, err
			}
			if e == nil {
				continue
			}
			part.Extractor = *e
			parts = append(parts, part)
		}
	}
	// Validate payload fields
	if len(parts) == 0 {
		return nil, errs.BadPayload{errs.ErrNoParts}
	}

	for _, part := range parts {
		if len(part.Name) == 0 || len(part.Selector) == 0 {
			e := fmt.Sprintf(errs.ErrNoPartOrSelectorProvided, part.Name+part.Selector)
			return nil, errs.BadPayload{e}
		}

	}
	return parts, nil
}

func (p Payload) newExtractor(t string, f *Field, part *Part, params *map[string]interface{}) (*extract.Extractor, error) {
	var e extract.Extractor
	switch strings.ToLower(t) {
	case "text":
		e = &extract.Text{
			Filters: f.Extractor.Filters,
		}
	case "href", "src", "path":
		extrAttr := t
		if t == "path" {
			extrAttr = "href"
		}
		e = &extract.Attr{
			Attr: extrAttr,
			//BaseURL: p.Request.URL,
		}
		//******* details
		if f.Details != nil {
			detailsPayload := p
			detailsPayload.Name = f.Name + "Details"
			detailsPayload.Fields = f.Details.Fields
			detailsPayload.Paginator = f.Details.Paginator
			detailsPayload.IsPath = f.Details.IsPath
			//Request refers to  srarting URL here. Requests will be changed in Scrape function to Details pages afterwards
			scraper, err := detailsPayload.newScraper("details")
			if err != nil {
				return nil, err
			}
			part.Details = *scraper
		}

	case "alt":
		e = &extract.Attr{
			Attr:    t,
			Filters: f.Extractor.Filters,
		}
	case "width", "height":
		e = &extract.Attr{Attr: t}
	case "regex":
		r := &extract.Regex{}
		regExp := (*params)["regexp"]
		r.Regex = regexp.MustCompile(regExp.(string))
		//it is obligatory parameter and we don't need to add it again in further fillStruct() func. So we can delete it here
		delete((*params), "regexp")
		e = r
	case "const":
		e = &extract.Const{Val: (*params)["value"]}
	case "count":
		e = &extract.Count{}
	// case "html":
	// 	e = &extract.Html{}
	case "outerhtml":
		e = &extract.OuterHtml{}

	default:
		logger.Error(t + ": Unknown selector type")
		return nil, nil
	}
	return &e, nil
}

func (task *Task) allowedByRobots(req fetch.Request) error {
	//get Robotstxt Data
	host, err := req.Host()
	if err != nil {
		return err
	}
	if _, ok := task.Robots[host]; !ok {
		robots, err := fetch.RobotstxtData(req.URL)
		if err != nil {
			robotsURL, err1 := fetch.AssembleRobotstxtURL(req.URL)
			if err1 != nil {
				return err1
			}
			logger.Warn(err.Error(),
				zap.String("Robots.txt URL", robotsURL))
		}
		task.Robots[host] = robots
	}

	//check if scraping of current url is not forbidden
	if !fetch.AllowedByRobots(req.URL, task.Robots[host]) {
		return errs.StatusError{403, errors.New(http.StatusText(http.StatusForbidden))}
	}
	return nil
}

// scrape is a core function which follows the rules listed in task payload, processes all pages/ details pages. It stores parsed results to Task.Results
func (task *Task) scrape(tw *taskWorker) (*Results, error) {
	defer task.jobDone.Done()
	/* task.mx.Lock()
	if err, scraped := task.statePool[tw.UID]; scraped {
		return nil, err.state
	}
	task.mx.Unlock() */
	req := tw.scraper.Request
	url := req.URL

	err := task.allowedByRobots(req)
	if err != nil {
		task.Cancel()
		return nil, err
	}

	//call remote fetcher to download web page
	//content, err := fetchContent(req)
	errorChan := make(chan error)
	resultChan := make(chan io.ReadCloser)
	fi := fetchInfo{
		request: req,
		reqType: tw.scraper.reqType,
		result:  resultChan,
		err:     errorChan,
	}
	task.fetchChannel <- &fi
	var content io.ReadCloser
	select {
	case err := <-errorChan:
		task.mx.Lock()
		task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: err}
		task.mx.Unlock()
		return nil, err
	case content = <-resultChan:
		//increment Task response count
		task.mx.Lock()
		count := task.responseCount
		task.responseCount = atomic.AddUint32(&count, 1)
		task.mx.Unlock()
	case <-task.ctx.Done():
		task.mx.Lock()
		task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: err}
		task.mx.Unlock()
		return nil, &errs.Cancel{}
	}

	// Create a goquery document.
	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		task.mx.Lock()
		task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: err}
		task.mx.Unlock()
		return nil, err
	}

	if tw.scraper.paginatorType == "next" {
		url, err = tw.scraper.Paginator.NextPage(url, doc.Selection)
		if err != nil {
			task.mx.Lock()
			task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: err}
			task.mx.Unlock()
			return nil, err
		}
		// Repeat until we don't have any more URLs, or until we hit our page limit.
		if len(url) != 0 &&
			viper.GetInt("MAX_PAGES") > 0 && tw.currentPageNum < viper.GetInt("MAX_PAGES")-1 {
			paginatorScraper := Scraper{
				DividePage:    tw.scraper.DividePage,
				IsPath:        tw.scraper.IsPath,
				Paginator:     tw.scraper.Paginator,
				paginatorType: tw.scraper.paginatorType,
				Parts:         tw.scraper.Parts,
				reqType:       tw.scraper.reqType,
				Request:       tw.scraper.Request,
			}
			paginatorScraper.Request.URL = url
			paginatorScraper.reqType = "paginator"
			curPageNum := tw.currentPageNum + 1
			if tw.useBlockCounter {
				curPageNum = 0
			}
			paginatorTW := taskWorker{
				currentPageNum:  curPageNum,
				scraper:         &paginatorScraper,
				UID:             tw.UID,
				keys:            tw.keys,
				useBlockCounter: tw.useBlockCounter,
			}
			task.jobDone.Add(1)
			go task.scrape(&paginatorTW)
		}
	}

	blockSelections := tw.scraper.DividePage(doc.Selection)
	if len(blockSelections) == 0 {
		task.mx.Lock()
		task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: err}
		task.mx.Unlock()
		return nil, &errs.NoBlocksToParse{URL: req.URL}
	}

	// Divide this page into blocks
	var miscChannel chan *blockStruct
	var miscWG sync.WaitGroup
	if tw.scraper.reqType != "initial" {
		miscChannel = make(chan *blockStruct)
		miscWG = sync.WaitGroup{}
		go task.blockWorker(miscChannel)
	}
	for i, blockSel := range blockSelections {
		// if scraper contain path then page ref should always be 0
		curPageNum := tw.currentPageNum
		if tw.scraper.IsPath {
			curPageNum = 0
		}
		ref := fmt.Sprintf("%s-%d-%d", tw.UID, curPageNum, i)
		block := blockStruct{
			blockSelection:  blockSel,
			key:             ref,
			hash:            tw.UID,
			useBlockCounter: tw.useBlockCounter,
			keys:            &tw.keys,
			scraper:         tw.scraper,
		}
		if tw.scraper.reqType == "initial" {
			task.blockChannel <- &block
			task.jobDone.Add(1)
		} else {
			block.wg = &miscWG
			miscWG.Add(1)
			miscChannel <- &block
		}
	}
	if tw.scraper.reqType != "initial" {
		miscWG.Wait()
		close(miscChannel)
		if tw.scraper.reqType == "paginator" {
			task.updateKeys(tw.UID, tw.keys)
		}
	}
	task.mx.Lock()
	task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: &errs.OK{}}
	task.mx.Unlock()
	return nil, err

}

//selectors returns selectors from payload
func (p Payload) selectors() ([]string, error) {
	selectors := []string{}
	for _, f := range p.Fields {
		if f.Selector != "" {
			selectors = append(selectors, f.Selector)
		}
	}
	if len(selectors) == 0 {
		return nil, errs.BadPayload{errs.ErrNoSelectors}
	}
	return selectors, nil
}

//response sends request to fetch service and returns fetch.FetchResponser
func fetchContent(req fetch.Request) (io.ReadCloser, error) {
	svc, err := fetch.NewHTTPClient(viper.GetString("DFK_FETCH"))
	if err != nil {
		logger.Error(err.Error())
	}
	return svc.Fetch(req)
}

//partNames returns Part Names which are used as a header of output CSV
func (s Scraper) partNames() []string {
	scr := s
	for scr.IsPath {
		for _, part := range scr.Parts {
			if len(part.Details.Parts) > 0 {
				scr = part.Details
				break
			}
		}
	}
	names := []string{}
	for _, part := range scr.Parts {
		names = append(names, part.Name)
	}
	return names
}

// First returns the first set of results - i.e. the results from the first
// block on the first page.
// This function can return nil if there were no blocks found on the first page
// of the scrape.
func (r *Results) First() map[string]interface{} {
	if len(r.Output[0]) == 0 {
		return nil
	}

	return r.Output[0][0]
}

// AllBlocks returns a single list of results from every block on all pages.
// This function will always return a list, even if no blocks were found.
func (r *Results) AllBlocks() []map[string]interface{} {
	ret := []map[string]interface{}{}

	for _, page := range r.Output {
		for _, block := range page {
			ret = append(ret, block)
		}
	}

	return ret
}

//KSUID stores the timestamp portion in ID. So we can retrieve it from Task object as a Time object
func (task Task) startTime() (*time.Time, error) {
	id, err := ksuid.Parse(task.ID)
	if err != nil {
		return nil, err
	}
	idTime := id.Time()
	return &idTime, nil
}

func (task *Task) blockWorker(blocks chan *blockStruct) {
	//defer wrk.wg.Done()
	for block := range blocks {
		select {
		default:
			blockResults := map[string]interface{}{}
			// Process each part of this block
			for _, part := range block.scraper.Parts {
				sel := block.blockSelection
				if part.Selector != "." {
					sel = sel.Find(part.Selector)
				}
				//update base URL to reflect attr relative URL change
				/* switch part.Extractor.(type) {
				case *extract.Attr: */
				retryImageIfFail := false
				attr, ok := part.Extractor.(*extract.Attr)
				if ok && (attr.Attr == "href" || attr.Attr == "src" || attr.Attr == "style") {
					task.mx.Lock()
					attr.BaseURL = block.scraper.Request.URL
					task.mx.Unlock()
					if attr.Attr == "src" || attr.Attr == "style" {
						retryImageIfFail = true
					}
				}
				/* } */
				task.mx.Lock()
				extractedPartResults, err := part.Extractor.Extract(sel)
				task.mx.Unlock()
				if err != nil {
					logger.Error(err.Error())
					return
				}
				// A nil response from an extractor means that we don't even include it in
				// the results.
				if extractedPartResults == nil {
					if retryImageIfFail {
						task.mx.Lock()
						attr, _ := part.Extractor.(*extract.Attr)
						attr.Attr = "style"
						extractedPartResults, err = part.Extractor.Extract(sel)
						if err != nil {
							logger.Error(err.Error())
							continue
						}
						task.mx.Unlock()
						if extractedPartResults == nil {
							continue
						}
					} else {
						continue
					}
				}
				if retryImageIfFail && attr.Attr == "style" {
					imgSrc, ok := extractedPartResults.(string)
					if ok {
						nStart := strings.Index(imgSrc, "url(") + len("url(")
						nEnd := strings.Index(imgSrc, ")")
						extractedPartResults = imgSrc[nStart:nEnd]
					}
				}
				if !block.scraper.IsPath {
					blockResults[part.Name] = extractedPartResults
				}
				//********* details
				if len(part.Details.Parts) > 0 {
					if !task.scrapeDetails(extractedPartResults, &part, block, &blockResults) {
						continue
					}
				}
				//********* end details
			}
			if len(blockResults) > 0 {
				task.saveToStorage(&blockResults, block)
			}
			if block.wg != nil {
				block.wg.Done()
			} else {
				task.jobDone.Done()
			}
		case <-task.ctx.Done():
			if block.wg != nil {
				block.wg.Done()
			} else {
				task.jobDone.Done()
			}
			//return
		}
	}
}

func (task *Task) scrapeDetails(extractedPartResults interface{}, part *Part, block *blockStruct, blockResults *map[string]interface{}) bool {
	var requests []fetch.Request

	switch extractedPartResults.(type) {
	case string:
		rq := fetch.Request{URL: extractedPartResults.(string)}
		requests = append(requests, rq)
	case []string:
		for _, r := range extractedPartResults.([]string) {
			rq := fetch.Request{URL: r}
			requests = append(requests, rq)
		}
	}
	for _, r := range requests {
		part.Details.Request = r
		//check if domain is the same for initial URL and details' URLs
		//If original host is the same as details' host sleep for some time before  fetching of details page  to avoid ban and other sanctions

		var uid string
		var blockKeys map[int][]int
		ubc := false
		if block.scraper.IsPath {
			uid = block.hash
			ubc = true
			blockKeys = *block.keys
		} else {
			uid = string(utils.GenerateCRC32([]byte(r.URL)))
			blockKeys = make(map[int][]int)
		}
		tw := taskWorker{
			currentPageNum:  0,
			scraper:         &part.Details,
			UID:             uid,
			useBlockCounter: ubc,
			keys:            blockKeys,
		}
		tw.scraper.Request.Type = task.Payload.Request.Type
		task.jobDone.Add(1)
		_, err := task.scrape(&tw)
		if err != nil {
			return false
		}
		if block.scraper.IsPath {
			return false
		}
		(*blockResults)[part.Name+"_details"] = uid //generate uid resDetails.AllBlocks()
		// Sort keys to keep an order before write them into storage.
		err = task.updateKeys(uid, tw.keys)
		if err != nil {
			// logger.Warning(fmt.Errorf("Failed to write %s. %s", string(uid), err.Error()))
			task.statePool[tw.UID] = scrapeState{url: tw.scraper.Request.URL, state: err}
			return false
		}
	}
	return true
}

func (task *Task) saveToStorage(blockResults *map[string]interface{}, block *blockStruct) {
	task.mx.Lock()
	if !task.Parsed {
		task.Parsed = true
	}
	task.mx.Unlock()

	output, err := json.Marshal(blockResults)
	if err != nil {
		logger.Error(err.Error())
	}
	if !block.scraper.IsPath {
		task.mx.Lock()
		key := block.key
		/* if block.useBlockCounter {
			blockNum := len(task.BlockCounter)
			key = fmt.Sprintf("%s-0-%d", block.hash, blockNum)
			task.BlockCounter = append(task.BlockCounter, blockNum)
		} else { */
		keys := strings.Split(block.key, "-")
		pageNum, err := strconv.Atoi(keys[1])
		if err != nil {
			logger.Error(fmt.Errorf("Failed to convert string to int %s. %s", string(key[1]), err.Error()).Error())
		}
		var blockNum int
		if block.useBlockCounter {
			blockNum = len((*block.keys)[pageNum])
			key = fmt.Sprintf("%s-0-%d", block.hash, blockNum)
		} else {
			blockNum, err = strconv.Atoi(keys[2])
			if err != nil {
				logger.Error(fmt.Errorf("Failed to convert string to int %s. %s", string(key[1]), err.Error()).Error())
			}
		}

		(*block.keys)[pageNum] = append((*block.keys)[pageNum], blockNum)
		//}
		err = task.storage.Write(storage.Record{
			Type:    storage.INTERMEDIATE,
			Key:     key,
			Value:   output,
			ExpTime: 0,
		})
		if err != nil {
			logger.Error(fmt.Errorf("Failed to write %s. %s", key, err.Error()).Error())
		}
		task.mx.Unlock()
	}
}

func (task *Task) fetchWorker() {
	for fetch := range task.fetchChannel {
		if !viper.GetBool("IGNORE_FETCH_DELAY") {
			if *task.Payload.RandomizeFetchDelay {
				//Sleep for time equal to FetchDelay * random value between 500 and 1500 msec
				rand := utils.Random(500, 1500)
				delay := *task.Payload.FetchDelay * time.Duration(rand) / 1000
				time.Sleep(delay)
			} else {
				time.Sleep(*task.Payload.FetchDelay)
			}
		}
		//increment Task request count
		task.mx.Lock()
		count := task.requestCount[fetch.reqType]
		task.requestCount[fetch.reqType] = atomic.AddUint32(&count, 1)
		task.mx.Unlock()

		content, err := fetchContent(fetch.request)
		if err != nil {
			fetch.err <- err
		} else {
			fetch.result <- content
		}
	}
}

func (task *Task) updateKeys(uid string, keys map[int][]int) error {
	task.mx.Lock()
	for k := range keys {
		sort.Slice(keys[k], func(i, j int) bool { return keys[k][i] < keys[k][j] })
	}
	task.mx.Unlock()
	j, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	err = task.storage.Write(storage.Record{
		Type:    storage.INTERMEDIATE,
		Key:     string(uid),
		Value:   j,
		ExpTime: 0,
	})
	if err != nil {
		return err
	}
	return nil
}

func (task *Task) closeTask() {
	close(task.fetchChannel)
	close(task.blockChannel)
	task.storage.Close()
}

func (task *Task) printLog(w io.Writer) {
	for _, state := range task.statePool {
		w.Write([]byte(fmt.Sprintf("URL: %s State: %s", state.url, state.state.Error())))
	}
}
