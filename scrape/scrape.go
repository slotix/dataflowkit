package scrape

// The following code was sourced and modified from the
// https://github.com/andrew-d/goscrape package governed by MIT license.

//TODO: add paginator to details
import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/slotix/dataflowkit/splash"

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

func init() {
	logger = log.NewLogger(true)
}

// NewTask creates new task to parse fetched page following the rules from Payload.
func NewTask(p Payload) *Task {
	//https://blog.kowalczyk.info/article/JyRZ/generating-good-random-and-unique-ids-in-go.html
	id := ksuid.New()
	//tQueue := make(chan *Scraper, 100)
	return &Task{
		ID:      id.String(),
		Payload: p,
		Visited: make(map[string]error),
		Robots:  make(map[string]*robotstxt.RobotsData),
	}

}

// Parse processes specified task which parses fetched page.
func (task *Task) Parse() (io.ReadCloser, error) {
	scraper, err := task.Payload.newScraper()
	if err != nil {
		return nil, err
	}
	//scrape request and return results.
	results, err := task.scrape(scraper)
	if err != nil {
		if task.Payload.Request.Type() == "splash" {
			return nil, err
		}
		logger.Error(err)
		task.Payload.FetcherType = "splash"
		request, err := task.Payload.initRequest()
		if err != nil {
			return nil, err
		}
		task.Payload.Request = request
		scraper.Request = request
		results, err = task.scrape(scraper)
		if err != nil {
			return nil, err
		}
	}
	//logger.Info(task.Visited)

	var e encoder
	switch strings.ToLower(task.Payload.Format) {
	case "csv":
		e = CSVEncoder{
			comma:     ",",
			partNames: scraper.partNames(),
		}
	case "json":
		e = JSONEncoder{
			paginateResults: *task.Payload.PaginateResults,
		}
	case "xml":
		e = XMLEncoder{}
	default:
		return nil, errors.New("invalid output format specified")
	}
	//logger.Info(results)
	r, err := e.Encode(results)
	if err != nil {
		return nil, err
	}
	return r, err
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
	//TODO: need to test the case when there are no selectors found in payload.
	var dividePageFunc DividePageFunc
	if len(selectors) == 0 {
		dividePageFunc = DividePageBySelector("body")
	} else {
		dividePageFunc = DividePageByIntersection(selectors)
	}

	scraper := &Scraper{
		Request:    p.Request,
		DividePage: dividePageFunc,
		Parts:      parts,
		Paginator:  paginator,
	}

	// All set!
	return scraper, nil
}

//fields2parts converts payload []field to []scrape.Part
func (p Payload) fields2parts() ([]Part, error) {
	parts := []Part{}
	//Payload fields
	for _, f := range p.Fields {
		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params
		}
		var err error

		for _, t := range f.Extractor.Types {
			part := Part{
				Name:     f.Name + "_" + t,
				Selector: f.Selector,
			}

			var e extract.Extractor
			switch strings.ToLower(t) {
			case "text":
				e = &extract.Text{
					Filters: f.Extractor.Filters,
				}
			case "href", "src":
				e = &extract.Attr{
					Attr: t,
					//BaseURL: p.Request.URL,
				}
				scraper := &Scraper{}
				//******* details
				if f.Details != nil {
					detailsPayload := p
					detailsPayload.Name = f.Name + "Details"
					detailsPayload.Fields = f.Details.Fields
					detailsPayload.Paginator = f.Details.Paginator
					//Request refers to  srarting URL here. Requests will be changed in Scrape function to Details pages afterwards
					scraper, err = detailsPayload.newScraper()
					if err != nil {
						return nil, err
					}
				} else {
					scraper = nil
				}
				part.Details = scraper
			case "alt":
				e = &extract.Attr{
					Attr:    t,
					Filters: f.Extractor.Filters,
				}
			case "width", "height":
				e = &extract.Attr{Attr: t}
			case "regex":
				r := &extract.Regex{}
				regExp := params["regexp"]
				r.Regex = regexp.MustCompile(regExp.(string))
				//it is obligatory parameter and we don't need to add it again in further fillStruct() func. So we can delete it here
				delete(params, "regexp")
				e = r
			case "const":
				//	c := &extract.Const{Val: params["value"]}
				//	e = c
				e = &extract.Const{}
			case "count":
				e = &extract.Count{}
			case "html":
				e = &extract.Html{}
			case "outerHtml":
				e = &extract.OuterHtml{}

			default:
				logger.Error(errors.New(t + ": Unknown selector type"))
				continue
			}
			part.Extractor = e

			if params != nil {
				err := fillStruct(params, e)
				if err != nil {
					logger.Error(err)
				}
			}
			//logger.Info(e)
			parts = append(parts, part)
		}
	}
	// Validate payload fields
	if len(parts) == 0 {
		return nil, &errs.BadPayload{errs.ErrNoParts}
	}
	seenNames := map[string]struct{}{}
	for i, part := range parts {
		if len(part.Name) == 0 {
			return nil, fmt.Errorf("no name provided for part %d", i)
		}
		if _, seen := seenNames[part.Name]; seen {
			return nil, fmt.Errorf("part %s has a duplicate name", part.Name)
		}
		seenNames[part.Name] = struct{}{}

		if len(part.Selector) == 0 {
			return nil, fmt.Errorf("no selector provided for part %d", i)
		}
	}
	return parts, nil
}

// scrape is a core function which follows the rules listed in task payload, processes all pages/ details pages. It stores parsed results to Task.Results
func (t *Task) scrape(scraper *Scraper) (*Results, error) {
	//logger.Info(time.Now())
	output := [][]map[string]interface{}{}

	req := scraper.Request
	//req := t.Payload.Request
	url := req.GetURL()

	var numPages int
	//get Robotstxt Data
	host, err := req.Host()
	if err != nil {
		t.Visited[url] = err
		logger.Error(err)
		//return err
	}
	if _, ok := t.Robots[host]; !ok {
		robots, err := fetch.RobotstxtData(url)
		if err != nil {
			robotsURL, err1 := fetch.AssembleRobotstxtURL(url)
			if err1 != nil {
				return nil, err1
			}
			t.Visited[url] = err
			logger.WithFields(
				logrus.Fields{
					"err": err,
				}).Warn("Robots.txt URL: ", robotsURL)
			//logger.Warning(err)
			//return err
		}
		t.Robots[host] = robots
	}

	for {
		results := []map[string]interface{}{}
		//check if scraping of current url is not forbidden
		if !fetch.AllowedByRobots(url, t.Robots[host]) {
			t.Visited[url] = &errs.ForbiddenByRobots{url}
		}
		// Repeat until we don't have any more URLs, or until we hit our page limit.
		if len(url) == 0 ||
			(t.Payload.Paginator != nil && (t.Payload.Paginator.MaxPages > 0 && numPages >= t.Payload.Paginator.MaxPages)) {
			break
		}
		//call remote fetcher to download web page
		content, err := fetchContent(req)
		if err != nil {
			return nil, err
		}

		// Create a goquery document.
		doc, err := goquery.NewDocumentFromReader(content)
		if err != nil {
			return nil, err
		}

		t.Visited[url] = nil

		// Divide this page into blocks
		for _, block := range scraper.DividePage(doc.Selection) {
			blockResults := map[string]interface{}{}

			// Process each part of this block
			for _, part := range scraper.Parts {
				sel := block
				if part.Selector != "." {
					sel = sel.Find(part.Selector)
				}
				//update base URL to reflect attr relative URL change
				//fmt.Println(reflect.TypeOf(part.Extractor))
				switch part.Extractor.(type) {
				case *extract.Attr:
					//	if part.Extractor.GetType() == "attr" {
					attr := part.Extractor.(*extract.Attr)
					if attr.Attr == "href" || attr.Attr == "src" {
						attr.BaseURL = url
					}
					//	}
				}
				extractedPartResults, err := part.Extractor.Extract(sel)
				if err != nil {
					return nil, err
				}

				// A nil response from an extractor means that we don't even include it in
				// the results.
				if extractedPartResults == nil {
					continue
				}
				blockResults[part.Name] = extractedPartResults

				//********* details
				//part.Details = nil
				if part.Details != nil {
					var requests []fetch.FetchRequester

					switch extractedPartResults.(type) {
					case string:
						var rq fetch.FetchRequester
						switch t.Payload.Request.Type() {
						case "base":
							rq = &fetch.BaseFetcherRequest{URL: extractedPartResults.(string)}
						case "splash":
							rq = &splash.Request{URL: extractedPartResults.(string)}
						default:
							err := errors.New("invalid fetcher type specified")
							logger.Error(err.Error())
							return nil, err
						}
						requests = append(requests, rq)
					case []string:
						for _, r := range extractedPartResults.([]string) {

							var rq fetch.FetchRequester
							switch t.Payload.Request.Type() {
							case "base":
								rq = &fetch.BaseFetcherRequest{URL: r}
							case "splash":
								rq = &splash.Request{URL: r}
							default:
								err := errors.New("invalid fetcher type specified")
								logger.Error(err.Error())
								return nil, err
							}

							requests = append(requests, rq)

						}
					}
					for _, r := range requests {
						//part.Details.Request = splash.Request{
						//	URL: extractedPartResults.(string),
						//}
						part.Details.Request = r
						//check if domain is the same for initial URL and details' URLs
						//If original host is the same as details' host sleep for some time before  fetching of details page  to avoid ban and other sanctions
						detailsHost, err := part.Details.Request.Host()
						if err != nil {
							logger.Error(err)
						}
						if detailsHost == host {
							if !viper.GetBool("IGNORE_FETCH_DELAY") {
								if *t.Payload.RandomizeFetchDelay {
									//Sleep for time equal to FetchDelay * random value between 500 and 1500 msec
									rand := utils.Random(500, 1500)
									delay := *t.Payload.FetchDelay * time.Duration(rand) / 1000
									logger.Infof("%s -> %v", delay, part.Details.Request.GetURL())
									time.Sleep(delay)
								} else {
									time.Sleep(*t.Payload.FetchDelay)
								}
							}
						}

						resDetails, err := t.scrape(part.Details)
						if err != nil {
							return nil, err
						}
						blockResults[part.Name+"_details"] = resDetails.AllBlocks()
					}
				}
				//********* end details
			}
			if len(blockResults) > 0 {
				// Append the results from this block.
				results = append(results, blockResults)
			}
		}
		if len(results) != 0 {
			output = append(output, results)
		}
		numPages++

		// Get the next page. If empty URL is returned there is no Next Pages to proceed.
		if t.Payload.Paginator == nil {
			url = ""
		}
		if t.Payload.Paginator != nil {
			if !t.Payload.Paginator.InfiniteScroll {
				url, err = scraper.Paginator.NextPage(url, doc.Selection)
				if err != nil {
					return nil, err
				}
			} else {
				url = ""
			}
		}

		if url != "" {
			var rq fetch.FetchRequester
			switch req.Type() {
			case "splash":
				rq = &splash.Request{URL: url}
			case "base":
				rq = &fetch.BaseFetcherRequest{URL: url}
			default:
				err := errors.New("invalid fetcher type specified")
				logger.Error(err.Error())
				return nil, err
			}
			req = rq

			if !viper.GetBool("IGNORE_FETCH_DELAY") {
				if *t.Payload.RandomizeFetchDelay {
					//Sleep for time equal to FetchDelay * random value between 500 and 1500 msec
					rand := utils.Random(500, 1500)
					delay := *t.Payload.FetchDelay * time.Duration(rand) / 1000
					logger.Infof("%s -> %v", delay, req.GetURL())
					time.Sleep(delay)
				} else {
					time.Sleep(*t.Payload.FetchDelay)
				}
			}
		}

	}
	if len(output) == 0 {
		return nil, &errs.BadPayload{errs.ErrEmptyResults}
	}

	// All good!
	return &Results{output}, err

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
func fetchContent(req fetch.FetchRequester) (io.ReadCloser, error) {
	svc, err := fetch.NewHTTPClient(viper.GetString("DFK_FETCH") /*, gklog.NewNopLogger()*/)
	if err != nil {
		logger.Error(err)
	}
	resp, err := svc.Response(req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return resp.GetHTML()
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
func (t Task) startTime() (*time.Time, error) {
	id, err := ksuid.Parse(t.ID)
	if err != nil {
		return nil, err
	}
	idTime := id.Time()
	return &idTime, nil
}
