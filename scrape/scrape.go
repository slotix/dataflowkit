package scrape

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/storage"
	"github.com/slotix/dataflowkit/utils"
	"github.com/spf13/viper"
	"github.com/temoto/robotstxt"
	"go.uber.org/zap"
)

func (f *Field) extract(content *goquery.Selection, results *map[string]interface{}, baseURL string) error {
	for _, attr := range f.Attrs {
		values := []interface{}{}
		var err error
		content.Find(f.CSSSelector).Each(func(index int, s *goquery.Selection) {
			switch strings.ToLower(attr) {
			case "text":
				value := s.Text()
				for _, filter := range f.Filters {
					filteredValue, err := filter.Apply(value)
					if err == nil {
						value = filteredValue
					} else {
						logger.Sugar().Error(err)
					}
				}
				values = append(values, value)
			case "outerhtml":
				if value, err := goquery.OuterHtml(s); err == nil {
					values = append(values, value)
				}
			case "path":
				if value, ok := s.Attr("href"); ok {
					value, err = utils.RelUrl(baseURL, value)
					if err != nil {
						return
					}
					values = append(values, value)
				}
			default:
				if value, ok := s.Attr(attr); ok {
					for _, filter := range f.Filters {
						value, err = filter.Apply(value)
						if err != nil {
							logger.Sugar().Error(err)
						}
					}
					if attr == "href" || attr == "src" {
						value, err = utils.RelUrl(baseURL, value)
						if err != nil {
							return
						}
					}
					values = append(values, value)
				}
			}
		})
		switch len(values) {
		case 0:
			return fmt.Errorf("%s", errs.ErrEmptyResults)
		case 1:
			(*results)[f.Name+"_"+attr] = values[0]
		default:
			(*results)[f.Name+"_"+attr] = values
		}
	}
	return nil
}

func (f Filter) Apply(data string) (string, error) {
	if data == "" {
		return "", errors.New("Data source is empty")
	}
	switch strings.ToLower(f.Name) {
	case "trim":
		data = strings.TrimSpace(data)
	case "lowercase":
		data = strings.ToLower(data)
	case "uppercase":
		data = strings.ToUpper(data)
	case "capitalize":
		data = strings.Title(data)
	case "regex":
		if f.Param == "" {
			return "", errors.New("No regex given")
		}
		regex := regexp.MustCompile(f.Param)
		if regex.NumSubexp() == 0 {
			regex = regexp.MustCompile("(" + f.Param + ")")
		}
		if regex.NumSubexp() > 1 {
			return "", errors.New("Regex filter doesn't support subexpressions")
		}
		ret := regex.FindAllStringSubmatch(data, -1)

		// A return value of nil == no match
		if ret == nil {
			return "", nil
		}
		results := ""
		// For each regex match...
		for _, submatches := range ret {
			// The 0th entry will be the match of the entire string.  The 1st
			// entry will be the first capturing group, which is what we want to
			// extract.
			if len(submatches) > 1 {
				results += submatches[1] + ";"
			}
		}
		return results, nil
	default:
		return "", fmt.Errorf("Unknown filter name %s", f.Name)
	}
	return data, nil
}

func (p *Payload) InitUID() {
	clone := Payload{
		Compressor:          p.Compressor,
		FetchDelay:          p.FetchDelay,
		Fields:              p.Fields,
		Format:              "",
		IsPath:              p.IsPath,
		Name:                p.Name,
		PaginateResults:     p.PaginateResults,
		Paginator:           p.Paginator,
		PayloadMD5:          "",
		RandomizeFetchDelay: p.RandomizeFetchDelay,
		Request:             p.Request,
		RetryTimes:          p.RetryTimes,
	}
	clone.Request.Type = ""
	for _, field := range clone.Fields {
		field.Attrs = []string{}
	}
	data, _ := json.Marshal(clone)
	p.PayloadMD5 = string(utils.GenerateCRC32(utils.GenerateMD5(data)))
}

func (p *Payload) fieldNames() []string {
	names := []string{}
	for _, field := range p.Fields {
		for _, attr := range field.Attrs {
			names = append(names, fmt.Sprintf("%s_%s", field.Name, attr))
		}
	}
	return names
}

// NewTask creates new task to parse fetched page following the rules from Payload.
func NewTask() *Task {
	storageType := viper.GetString("STORAGE_TYPE")
	return &Task{
		//Errors:       []error{},
		Robots:       make(map[string]*robotstxt.RobotsData),
		requestCount: 0,
		storage:      storage.NewStore(storageType),
		jobDone:      sync.WaitGroup{},
	}

}

func (task *Task) checkPayload(p *Payload) error {
	if len(p.Fields) == 0 {
		return errors.New("Bad payload: No fields to scrape")
	}
	for i, field := range p.Fields {
		if field.Name == "" {
			return fmt.Errorf("Bad payload: Field %d has no name", i)
		}
		if field.CSSSelector == "" {
			return fmt.Errorf("Bad payload: Field %d has no css selector", i)
		}
		if len(field.Attrs) == 0 {
			return fmt.Errorf("Bad payload: Field %d has no attributes to extract", i)
		}
	}
	supportedOutputFormats := map[string]interface{}{"json": nil, "jsonl": nil, "xml": nil, "csv": nil}
	if _, ok := supportedOutputFormats[strings.ToLower(p.Format)]; !ok {
		return fmt.Errorf("Bad payload: Unsupported output format %s", p.Format)
	}
	return nil
}

// Parse specified payload.
func (task *Task) Parse(payload Payload) (io.ReadCloser, error) {
	task.payloads = make(chan Payload, viper.GetInt("PAYLOAD_POOL_SIZE"))
	defer close(task.payloads)

	begin := time.Now()
	err := task.checkPayload(&payload)
	if err != nil {
		return nil, err
	}
	for i := 0; i < viper.GetInt("PAYLOAD_WORKERS_NUM"); i++ {
		go task.scrapeContent(context.Background())
	}

	payload.InitUID()
	task.rootUID = payload.PayloadMD5
	task.templateRequest = payload.Request

	task.jobDone.Add(1)
	task.payloads <- payload
	task.jobDone.Wait()

	if !task.isParsed && payload.Request.Type != "chrome" {
		payload.Request.Type = "chrome"
		payload.InitUID()
		task.rootUID = payload.PayloadMD5
		task.templateRequest = payload.Request
		task.jobDone.Add(1)
		task.payloads <- payload
		task.jobDone.Wait()
	}
	if !task.isParsed {
		return nil, errs.ParseError{URL: payload.Request.URL, Err: errors.New(errs.ErrEmptyResults)}
	}
	var e encoder
	switch strings.ToLower(payload.Format) {
	case "csv":
		e = CSVEncoder{
			comma:     ",",
			partNames: payload.fieldNames(),
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
	case "xlsx":
		e = XLSXEncoder{
			partNames: payload.fieldNames(),
		}
	default:
		return nil, errors.New("invalid output format specified")
	}
	r, err := EncodeToFile(context.Background(), &e, encodeInfo{
		payloadMD5: string(payload.PayloadMD5),
		extension:  payload.Format,
		compressor: strings.ToLower(payload.Compressor),
		// compressLevel: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to encode to %s", payload.Format)
	}

	m := map[string]interface{}{
		"Task ID":     payload.PayloadMD5,
		"Requests":    task.requestCount,
		"Responses":   task.responseCount,
		"Output file": string(r),
		"Took":        time.Since(begin).String(),
	}
	parseResults, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(parseResults)), nil
}

func (task *Task) allowedByRobots(req fetch.Request, initFetchWorkers bool) error {
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

//response sends request to fetch service and returns fetch.FetchResponser
func fetchContent(req fetch.Request) (io.ReadCloser, error) {
	svc, err := fetch.NewHTTPClient(viper.GetString("DFK_FETCH"))
	if err != nil {
		logger.Error(err.Error())
	}
	return svc.Fetch(req)
}

func (task *Task) scrapeContent(ctx context.Context) error {
	for payload := range task.payloads {
		select {
		case <-ctx.Done():
			return nil
		default:
			var errs []<-chan error
			fetchChannel := make(chan flow)
			content, errc := task.fetch(ctx, fetchChannel, payload.PayloadMD5)
			errs = append(errs, errc)
			paginateContent, errc := task.paginate(ctx, content, payload.Paginator, fetchChannel)
			errs = append(errs, errc)
			blockChannel, errc, err := task.divide(ctx, paginateContent, payload.Fields)
			if err != nil {
				return err
			}
			errs = append(errs, errc)
			scrapedChannel, errc := task.parse(ctx, blockChannel, payload.Fields, payload.IsPath, payload.blockCounter)
			errs = append(errs, errc)
			fetchChannel <- flow{"", "", payload.Request}
			errc = task.saveIntermediate(ctx, scrapedChannel)
			errs = append(errs, errc)
			WaitForPipeline(errs...)
			task.jobDone.Done()
		}
	}
	return nil
}

func (task *Task) fetch(ctx context.Context, in <-chan flow, uid string) (<-chan flow, <-chan error) {
	contentChannel := make(chan flow)
	errc := make(chan error)

	go func() {
		defer close(contentChannel)
		defer close(errc)
		for data := range in {
			var fetchDelay time.Duration
			if !viper.GetBool("IGNORE_FETCH_DELAY") {
				rand := utils.Random(500, 1500)
				fetchDelay = (time.Duration(viper.GetInt("FETCH_DELAY")) + time.Duration(rand)*time.Millisecond)
			}
			select {
			case <-time.After(fetchDelay):
				if request, ok := data.data.(fetch.Request); ok {
					task.mx.Lock()
					task.requestCount++
					task.mx.Unlock()
					request.Type = task.templateRequest.Type
					content, err := fetchContent(request)
					if err != nil {
						errc <- errs.ParseError{request.URL, err}
						return
					} else {
						task.mx.Lock()
						task.responseCount++
						task.mx.Unlock()
						contentChannel <- flow{fmt.Sprintf("%s", uid), request.URL, content}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return contentChannel, errc
}

func (task *Task) paginate(ctx context.Context, in <-chan flow, nextPageSelector string, fetcherChannel chan flow) (<-chan flow, <-chan error) {
	contentChannel := make(chan flow)
	errc := make(chan error)
	go func() {
		defer close(contentChannel)
		defer close(errc)
		currentPageNum := 0
		for data := range in {
			select {
			case <-ctx.Done():
				return
			default:
				if nextPageSelector == "" {
					data.key = fmt.Sprintf("%s-%d", data.key, currentPageNum)
					contentChannel <- data
					close(fetcherChannel)
					return
				}
				content := data.data.(io.ReadCloser)
				doc, err := goquery.NewDocumentFromReader(content)
				if err != nil {
					errc <- errs.ParseError{data.url, err}
					continue
				}
				// feed parser with data
				selectionContent, _ := goquery.OuterHtml(doc.Selection)
				contentChannel <- flow{fmt.Sprintf("%s-%d", data.key, currentPageNum), data.url, ioutil.NopCloser(strings.NewReader(selectionContent))}
				f := Field{CSSSelector: nextPageSelector, Attrs: []string{"href"}, Name: "paginator"}
				paginator := make(map[string]interface{})
				err = f.extract(doc.Selection, &paginator, task.templateRequest.URL) /* tw.scraper.Paginator.NextPage(url, doc.Selection) */
				if err != nil {
					errc <- errs.ParseError{data.url, err}
					close(fetcherChannel)
					return
				}
				// Repeat until we don't have any more URLs, or until we hit our page limit.
				if len(paginator) != 0 &&
					viper.GetInt("MAX_PAGES") > 0 && currentPageNum < viper.GetInt("MAX_PAGES")-1 {
					// TODO clone request to use same settings
					currentPageNum++
					fetcherChannel <- flow{fmt.Sprintf("%s-%d", data.key, currentPageNum), data.url,
						fetch.Request{
							Actions:   "",
							FormData:  "",
							Method:    "",
							Type:      "",
							URL:       paginator["paginator_href"].(string),
							UserToken: "",
						},
					}
				} else {
					close(fetcherChannel)
					return
				}
			}
		}
	}()
	return contentChannel, errc
}

func (task *Task) divide(ctx context.Context, in <-chan flow, fields []Field) (<-chan flow, <-chan error, error) {
	if len(fields) == 0 {
		return nil, nil, errors.New("No fields to parse")
	}
	blockChannel := make(chan flow)
	errc := make(chan error)
	go func() {
		defer close(blockChannel)
		defer close(errc)
		for data := range in {
			select {
			case <-ctx.Done():
				return
			default:
				content := data.data.(io.ReadCloser)
				doc, err := goquery.NewDocumentFromReader(content)
				if err != nil {
					errc <- errs.ParseError{data.url, err}
					continue
				}
				var selectorAncestor *goquery.Selection
				index := -1
				for i, field := range fields {
					selectorAncestor = doc.Find(field.CSSSelector).First().Parent()
					if selectorAncestor.Length() > 0 {
						index = i
						break
					}
				}
				if index < 0 {
					errc <- errs.ParseError{URL: data.url, Err: errors.New(errs.ErrNoSelectors)}
					continue
				}
				bFound := false
				selectorsSlice := fields[index+1:]
				if len(selectorsSlice) > 0 {
					for !bFound {
						for _, f := range selectorsSlice {
							sel := doc.Find(f.CSSSelector).First()
							sel = sel.ParentsUntilSelection(selectorAncestor).Last()
							//check last node.. if it equal html its mean that first selector's parent
							//not found
							if goquery.NodeName(sel) == "html" {
								selectorAncestor = doc.FindSelection(selectorAncestor.Parent().First())
								bFound = false
								break
							}
							bFound = true
						}
					}
				}
				fullPath := goquery.NodeName(selectorAncestor)
				parents := selectorAncestor.ParentsUntilSelection(doc.Find("body"))
				parents.Each(func(i int, s *goquery.Selection) {
					//avoid antiscrapin' tech like twitter
					selector := attrOrDataValue(s)
					fullPath = selector + " > " + fullPath
				})
				items := doc.Find(fullPath)
				if items.Length() == 0 {
					errc <- errs.ParseError{data.url, errors.New("No blocks found")}
					continue
					//return nil, &errs.BadPayload{ErrText: "No blocks found"}
				}
				items.Each(func(i int, s *goquery.Selection) {
					blockChannel <- flow{data.key, data.url, s}
				})
			}
		}
	}()
	return blockChannel, errc, nil
}

func (task *Task) parse(ctx context.Context, in <-chan flow, fields []Field, isPath bool, blockCounter *int) (<-chan storage.Record, <-chan error) {
	result := make(chan storage.Record)
	errc := make(chan error)
	if blockCounter == nil {
		blockCounter = new(int)
	}
	go func() {
		defer close(result)
		defer close(errc)
		for data := range in {
			select {
			case <-ctx.Done():
				return
			default:
				block := data.data.(*goquery.Selection)
				blockResult := make(map[string]interface{})
				for _, field := range fields {

					if isPath && strings.ToLower(field.Attrs[0]) != "path" {
						continue
					}
					err := field.extract(block, &blockResult, task.templateRequest.URL)
					if err != nil {
						errc <- err
						continue
					}
					if len(field.Details.Fields) > 0 {
						var detailsURL interface{}
						if !isPath {
							detailsURL = blockResult[field.Name+"_href"]
						} else {
							detailsURL = blockResult[field.Name+"_path"]
							field.Details.PayloadMD5 = task.rootUID
						}
						switch detailsURL.(type) {
						case string:
							field.Details.Request.URL = detailsURL.(string)
							task.jobDone.Add(1)
							if isPath {
								field.Details.blockCounter = blockCounter
							} else {
								field.Details.InitUID()
							}
							task.payloads <- field.Details
						case []string:
							for _, url := range detailsURL.([]string) {
								field.Details.Request.URL = url
								task.jobDone.Add(1)
								if isPath {
									field.Details.blockCounter = blockCounter
								} else {
									field.Details.InitUID()
								}
								task.payloads <- field.Details
							}
						}
						// save reference to  details uid to be able restore it from storage
						if !isPath {
							blockResult[field.Name+"_details"] = field.Details.PayloadMD5
						}
					}
				}
				if len(blockResult) > 0 && !isPath {
					task.mx.Lock()
					if !task.isParsed {
						task.isParsed = true
					}
					task.mx.Unlock()
					strValue, err := json.Marshal(blockResult)
					if err != nil {
						errc <- errs.ParseError{data.url, fmt.Errorf("Fail marshal result: %#v", blockResult)}
						continue
					}
					task.mx.Lock()
					recordKey := fmt.Sprintf("%s-%d", data.key, *blockCounter)
					result <- storage.Record{
						Key:     recordKey,
						Type:    storage.INTERMEDIATE,
						Value:   strValue,
						ExpTime: 0,
					}
					*blockCounter++
					task.mx.Unlock()
				}
			}
		}
	}()
	return result, errc
}

func (task *Task) saveIntermediate(ctx context.Context, in <-chan storage.Record) <-chan error {
	errc := make(chan error)
	go func() {
		defer close(errc)
		for rec := range in {
			select {
			case <-ctx.Done():
				return
			default:
				err := task.storage.Write(rec)
				if err != nil {
					errc <- err
				}
			}
		}
	}()
	return errc
}

func WaitForPipeline(errs ...<-chan error) error {
	errc := MergeErrors(errs...)
	for err := range errc {
		if err != nil {
			//return err
			logger.Sugar().Error(err)
		}
	}
	return nil
}

func MergeErrors(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error, len(cs))
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func attrOrDataValue(s *goquery.Selection) (value string) {
	if s.Length() == 0 {
		return ""
	}
	attr, exists := s.Attr("class")
	attr = strings.TrimSpace(attr)
	if exists && attr != "" { //in some cases tag is invalid f.e. <tr class>
		var re = regexp.MustCompile(`\n?\s{1,}`)
		attr = "." + re.ReplaceAllString(attr, `.`)
		return attr
	}
	attr, exists = s.Attr("id")

	if exists && attr != "" {
		return fmt.Sprintf("#%s", attr)
	}
	//if len(s.Nodes)>0 {
	return s.Nodes[0].Data
	//}
}
