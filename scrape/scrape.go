package scrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/slotix/dataflowkit/storage"

	"github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
	"github.com/segmentio/ksuid"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/logger"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/utils"
	"github.com/spf13/viper"
	"github.com/temoto/robotstxt"
)

var logger *logrus.Logger

var fetchCannel chan *fetchInfo

func init() {
	logger = log.NewLogger(true)
}

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
		if p.Paginator.InfiniteScroll {
			p.Request.InfiniteScroll = true
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
	return &Task{
		ID:           id.String(),
		Payload:      p,
		Errors:       []error{},
		Robots:       make(map[string]*robotstxt.RobotsData),
		Parsed:       false,
		BlockCounter: []int{},
		storage:      storage.NewStore(storageType),
		mx:           &sync.Mutex{},
	}

}

// Parse processes specified task which parses fetched page.
func (task *Task) Parse() (io.ReadCloser, error) {

	scraper, err := task.Payload.newScraper()
	if err != nil {
		return nil, err
	}
	//scrape request and return results.

	fetchCannel = make(chan *fetchInfo, 100)
	for i := 0; i < 50; i++ {
		go task.fetchWorker(fetchCannel)
	}
	// Array of page keys
	wg := sync.WaitGroup{}
	uid := string(utils.GenerateCRC32([]byte(task.Payload.PayloadMD5)))
	mx := sync.Mutex{}
	tw := taskWorker{
		wg:              &wg,
		currentPageNum:  0,
		scraper:         scraper,
		UID:             uid,
		mx:              &mx,
		useBlockCounter: false,
		keys:            make(map[int][]int),
	}
	wg.Add(1)
	_, err = task.scrape(&tw)
	wg.Wait()
	if !task.Parsed {
		logger.Info("Failed to scrape with base fetcher. Reinitializing to scrape with Chrome fetcher.")
		if task.Payload.Request.Type == "chrome" {
			close(fetchCannel)
			return nil, err
		}
		task.Payload.Request.Type = "chrome"
		scraper.Request.Type = "chrome"
		//request := task.Payload.initRequest("")
		//task.Payload.Request = request
		//scraper.Request = request
		wg.Add(1)
		_, err = task.scrape(&tw)
		wg.Wait()
		if !task.Parsed {
			close(fetchCannel)
			return nil, err
		}
	}
	close(fetchCannel)

	if len(task.BlockCounter) > 0 {
		tw.keys[0] = task.BlockCounter
	} else {
		// We have to sort a keys to keep an order
		for k := range tw.keys {
			sort.Slice(tw.keys[k], func(i, j int) bool { return tw.keys[k][i] < tw.keys[k][j] })
		}
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
	case "xml":
		e = XMLEncoder{}
	default:
		return nil, errors.New("invalid output format specified")
	}
	r, err := EncodeToFile(&e, task.Payload.Format, string(uid))
	if err != nil {
		return nil, err
	}
	fName := ioutil.NopCloser(bytes.NewReader(r))
	return fName, err
}

// Create a new scraper with the provided configuration.
func (p Payload) newScraper() (*Scraper, error) {
	parts, err := p.fields2parts()
	if err != nil {
		return nil, err
	}
	var paginator paginate.Paginator
	if p.Paginator == nil {
		paginator = &dummyPaginator{}

	} else {
		paginator = paginate.BySelector(p.Paginator.Selector, p.Paginator.Attribute)
	}

	selectors, err := p.selectors()
	if err != nil {
		return nil, err
	}

	var dividePageFunc DividePageFunc

	dividePageFunc = DividePageByIntersection(selectors)

	scraper := &Scraper{
		Request:    p.Request,
		DividePage: dividePageFunc,
		Parts:      parts,
		Paginator:  paginator,
		IsPath:     p.IsPath,
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
		return nil, &errs.BadPayload{errs.ErrNoParts}
	}

	for _, part := range parts {
		if len(part.Name) == 0 || len(part.Selector) == 0 {
			e := fmt.Sprintf(errs.ErrNoPartOrSelectorProvided, part.Name+part.Selector)
			return nil, &errs.BadPayload{e}
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
			scraper, err := detailsPayload.newScraper()
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
		logger.Error(errors.New(t + ": Unknown selector type"))
		return nil, nil
	}
	return &e, nil
}

func (task *Task) allowedByRobots(req fetch.Request) error {
	//get Robotstxt Data
	host, err := req.Host()
	if err != nil {
		task.Errors = append(task.Errors, err)
		logger.Error(err)
		//return err
	}
	if _, ok := task.Robots[host]; !ok {
		robots, err := fetch.RobotstxtData(req.URL)
		if err != nil {
			robotsURL, err1 := fetch.AssembleRobotstxtURL(req.URL)
			if err1 != nil {
				return err1
			}
			task.Errors = append(task.Errors, err)
			logger.WithFields(
				logrus.Fields{
					"err": err,
				}).Warn("Robots.txt URL: ", robotsURL)
			//logger.Warning(err)
			//return err
		}
		task.Robots[host] = robots
	}

	//check if scraping of current url is not forbidden
	if !fetch.AllowedByRobots(req.URL, task.Robots[host]) {
		task.Errors = append(task.Errors, &errs.ForbiddenByRobots{req.URL})
	}
	return nil
}

// scrape is a core function which follows the rules listed in task payload, processes all pages/ details pages. It stores parsed results to Task.Results
func (task *Task) scrape(tw *taskWorker) (*Results, error) {

	req := tw.scraper.Request
	url := req.URL

	err := task.allowedByRobots(req)
	if err != nil {
		//tw.wg.Done()
		return nil, err
	}

	//call remote fetcher to download web page
	//content, err := fetchContent(req)
	errorChan := make(chan error)
	resultChan := make(chan io.ReadCloser)
	fi := fetchInfo{
		request: req,
		result:  resultChan,
		err:     errorChan,
	}
	fetchCannel <- &fi
	var content io.ReadCloser
	select {
	case err := <-errorChan:
		tw.wg.Done()
		return nil, err
	case content = <-resultChan:
	}

	// Create a goquery document.
	doc, err := goquery.NewDocumentFromReader(content)
	if err != nil {
		tw.wg.Done()
		return nil, err
	}

	if task.Payload.Paginator != nil {
		if !task.Payload.Paginator.InfiniteScroll {
			url, err = tw.scraper.Paginator.NextPage(url, doc.Selection)
			if err != nil {
				tw.wg.Done()
				return nil, err
			}
			// Repeat until we don't have any more URLs, or until we hit our page limit.
			if len(url) != 0 &&
				task.Payload.Paginator.MaxPages > 0 && tw.currentPageNum < task.Payload.Paginator.MaxPages {
				paginatorPayload := task.Payload
				paginatorPayload.Request.URL = url
				//paginatorPayload.Request = paginatorPayload.initRequest(url)
				paginatorScraper, err := paginatorPayload.newScraper()
				if err != nil {
					tw.wg.Done()
					return nil, err
				}
				curPageNum := tw.currentPageNum + 1
				if tw.scraper.IsPath {
					curPageNum = 0
				}
				paginatorTW := taskWorker{
					wg:             tw.wg,
					currentPageNum: curPageNum,
					scraper:        paginatorScraper,
					UID:            tw.UID,
					mx:             tw.mx,
					keys:           tw.keys,
				}
				tw.wg.Add(1)
				go task.scrape(&paginatorTW)
			}
		}
		//todo: test this case
		// } else {
		// 	url = ""
		// }
	}
	blocks := make(chan *blockStruct)
	wg := sync.WaitGroup{}
	wrk := &worker{
		wg:      &wg,
		scraper: tw.scraper,
	}

	blockSelections := tw.scraper.DividePage(doc.Selection)

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go task.blockWorker(blocks, wrk)
	}

	// Divide this page into blocks
	for i, blockSel := range blockSelections {
		ref := fmt.Sprintf("%s-%d-%d", tw.UID, tw.currentPageNum, i)
		block := blockStruct{
			blockSelection:  blockSel,
			key:             ref,
			hash:            tw.UID,
			useBlockCounter: tw.useBlockCounter,
			keys:            &tw.keys,
		}
		blocks <- &block
	}
	close(blocks)
	wg.Wait()
	tw.wg.Done()
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
		return nil, &errs.BadPayload{errs.ErrNoSelectors}
	}
	return selectors, nil
}

//response sends request to fetch service and returns fetch.FetchResponser
func fetchContent(req fetch.Request) (io.ReadCloser, error) {
	svc, err := fetch.NewHTTPClient(viper.GetString("DFK_FETCH"))
	if err != nil {
		logger.Error(err)
	}
	return svc.Fetch(req)
}

//partNames returns Part Names which are used as a header of output CSV
func (s Scraper) partNames() []string {
	names := []string{}
	for _, part := range s.Parts {
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

func (task *Task) blockWorker(blocks chan *blockStruct, wrk *worker) {
	defer wrk.wg.Done()
	url := wrk.scraper.Request.URL
	for block := range blocks {
		blockResults := map[string]interface{}{}

		// Process each part of this block
		for _, part := range wrk.scraper.Parts {
			sel := block.blockSelection
			if part.Selector != "." {
				sel = sel.Find(part.Selector)
			}
			//update base URL to reflect attr relative URL change
			/* switch part.Extractor.(type) {
			case *extract.Attr: */
			attr, ok := part.Extractor.(*extract.Attr)
			if ok && (attr.Attr == "href" || attr.Attr == "src") {
				task.mx.Lock()
				attr.BaseURL = url
				task.mx.Unlock()
			}
			/* } */
			task.mx.Lock()
			extractedPartResults, err := part.Extractor.Extract(sel)
			task.mx.Unlock()
			if err != nil {
				logger.Error(err)
				return
			}

			// A nil response from an extractor means that we don't even include it in
			// the results.
			if extractedPartResults == nil {
				continue
			}
			if !wrk.scraper.IsPath {
				blockResults[part.Name] = extractedPartResults
			}
			//********* details
			if len(part.Details.Parts) > 0 {
				if !task.scrapeDetails(extractedPartResults, &part, wrk, block, &blockResults) {
					continue
				}
			}
			//********* end details
		}
		if len(blockResults) > 0 {
			task.saveToStorage(&blockResults, wrk, block)
		}
	}
}

func (task *Task) scrapeDetails(extractedPartResults interface{}, part *Part, wrk *worker, block *blockStruct, blockResults *map[string]interface{}) bool {
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

		wg := sync.WaitGroup{}
		var uid string
		ubc := false
		if wrk.scraper.IsPath {
			uid = block.hash
			ubc = true
		} else {
			uid = string(utils.GenerateCRC32([]byte(r.URL)))
		}
		tw := taskWorker{
			wg:              &wg,
			currentPageNum:  0,
			scraper:         &part.Details,
			UID:             uid,
			useBlockCounter: ubc,
			keys:            make(map[int][]int),
		}
		wg.Add(1)
		tw.scraper.Request.Type = task.Payload.Request.Type
		_, err := task.scrape(&tw)
		if err != nil {
			logger.Error(err)
			return false
		}
		if wrk.scraper.IsPath {
			return false
		}
		(*blockResults)[part.Name+"_details"] = uid //generate uid resDetails.AllBlocks()
		// Sort keys to keep an order before write them into storage.
		for k := range tw.keys {
			sort.Slice(tw.keys[k], func(i, j int) bool { return tw.keys[k][i] < tw.keys[k][j] })
		}
		j, err := json.Marshal(tw.keys)
		if err != nil {
			//return nil, err
			logger.Warning(fmt.Errorf("Failed to marshal details key. %s", err.Error()))
			return false
		}
		err = task.storage.Write(storage.Record{
			Type:    storage.INTERMEDIATE,
			Key:     string(uid),
			Value:   j,
			ExpTime: 0,
		})
		if err != nil {
			logger.Warning(fmt.Errorf("Failed to write %s. %s", string(uid), err.Error()))
			return false
		}
	}
	return true
}

func (task *Task) saveToStorage(blockResults *map[string]interface{}, wrk *worker, block *blockStruct) {
	task.mx.Lock()
	if !task.Parsed {
		task.Parsed = true
	}
	task.mx.Unlock()

	output, err := json.Marshal(blockResults)
	if err != nil {
		logger.Error(err)
	}
	if !wrk.scraper.IsPath {
		task.mx.Lock()
		key := block.key
		if block.useBlockCounter {
			blockNum := len(task.BlockCounter)
			key = fmt.Sprintf("%s-0-%d", block.hash, blockNum)
			task.BlockCounter = append(task.BlockCounter, blockNum)
		} else {
			keys := strings.Split(block.key, "-")
			pageNum, err := strconv.Atoi(keys[1])
			if err != nil {
				logger.Error(fmt.Errorf("Failed to convert string to int %s. %s", string(key[1]), err.Error()))
			}
			blockNum, err := strconv.Atoi(keys[2])
			if err != nil {
				logger.Error(fmt.Errorf("Failed to convert string to int %s. %s", string(key[1]), err.Error()))
			}
			(*block.keys)[pageNum] = append((*block.keys)[pageNum], blockNum)
		}
		err = task.storage.Write(storage.Record{
			Type:    storage.INTERMEDIATE,
			Key:     key,
			Value:   output,
			ExpTime: 0,
		})
		if err != nil {
			logger.Error(fmt.Errorf("Failed to write %s. %s", key, err.Error()))
		}
		task.mx.Unlock()
	}
}

func (task *Task) fetchWorker(fc chan *fetchInfo) {
	for fetch := range fc {
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
		content, err := fetchContent(fetch.request)
		if err != nil {
			fetch.err <- err
		} else {
			fetch.result <- content
		}
	}
}
