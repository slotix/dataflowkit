package scrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/segmentio/ksuid"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "scrape: ", log.Lshortfile)
}

var (
	ErrNoParts     = errors.New("no pieces in payload")
	ErrNoSelectors = errors.New("no selectors in payload")
)

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
func (t Task) StartTime() (*time.Time, error) {
	id, err := ksuid.Parse(t.ID)
	if err != nil {
		return nil, err
	}
	idTime := id.Time()
	return &idTime, nil
}

//selectors returns selectors from payload
func (p Payload) selectors() ([]string, error) {
	selectors := []string{}
	for _, f := range p.Fields {
		selectors = append(selectors, f.Selector)
	}
	if len(selectors) == 0 {
		return nil, errNoSelectors
	}
	return selectors, nil
}

func NewTask(p Payload) (task *Task, err error) {
	scraper := &Scraper{}
	//if p.Request != nil {
		scraper, err = NewScraper(p)
		if err != nil {
			return nil, err
		}
//	}
	//https://blog.kowalczyk.info/article/JyRZ/generating-good-random-and-unique-ids-in-go.html
	id := ksuid.New()

	task = &Task{
		ID:      id.String(),
		Scraper: scraper,
	}
	return task, nil
}

//fields2parts converts payload []field to []scrape.Part
func (p Payload) fields2parts() ([]Part, error) {
	parts := []Part{}
	//Payload fields
	for _, f := range p.Fields {
		params := make(map[string]interface{})
		if f.Extractor.Params != nil {
			params = f.Extractor.Params.(map[string]interface{})
		}
		var err error
		switch eType := f.Extractor.Type; eType {

		//For Link type Two pieces as pair Text and Attr{Attr:"href"} extractors are added.
		case "link":
			l := &extract.Link{Href: extract.Attr{Attr: "href"}}
			if params != nil {
				err := FillStruct(params, l)
				if err != nil {
					logger.Println(err)
				}
			}
			parts = append(parts, Part{
				Name:      f.Name + "_text",
				Selector:  f.Selector,
				Extractor: l.Text,
			})
			//******* details
			
			if f.Details != nil {
				task := &Task{}
				detailsPayload := p
				detailsPayload.Name = f.Name + "Details"
				detailsPayload.Fields = f.Details.Fields
				//Request refers to  srarting URL here. Requests will be changed in Scrape function to Details pages afterwards
				task, err = NewTask(detailsPayload)
				if err != nil {
					return nil, err
				}
				parts = append(parts,  Part{
					Name:      f.Name + "_link",
					Selector:  f.Selector,
					Extractor: l.Href,
					Details:   task,
				})
			} 
			parts = append(parts,  Part{
				Name:      f.Name + "_link",
				Selector:  f.Selector,
				Extractor: l.Href,
				Details:   nil,
			})

			

		//For image type by default Two pieces with different Attr="src" and Attr="alt" extractors will be added for field selector.
		case "image":
			i := &extract.Image{Src: extract.Attr{Attr: "src"},
				Alt: extract.Attr{Attr: "alt"}}
			if params != nil {
				err := FillStruct(params, i)
				if err != nil {
					logger.Println(err)
				}
			}
			parts = append(parts, Part{
				Name:      f.Name + "_src",
				Selector:  f.Selector,
				Extractor: i.Src,
			}, Part{
				Name:      f.Name + "_alt",
				Selector:  f.Selector,
				Extractor: i.Alt,
			})

		default:
			var e extract.Extractor
			switch eType {
			case "const":
				//	c := &extract.Const{Val: params["value"]}
				//	e = c
				e = &extract.Const{}
			case "count":
				e = &extract.Count{}
			case "text":
				e = &extract.Text{}
			case "html":
				e = &extract.Html{}
			case "outerHtml":
				e = &extract.OuterHtml{}
			case "attr":
				e = &extract.Attr{}
			case "regex":
				r := &extract.Regex{}
				regExp := params["regexp"]
				r.Regex = regexp.MustCompile(regExp.(string))
				e = r
			}

			if params != nil {
				err := FillStruct(params, e)
				if err != nil {
					logger.Println(err)
				}
			}
			parts = append(parts, Part{
				Name:      f.Name,
				Selector:  f.Selector,
				Extractor: e,
			})
		}
	}
	// Validate payload fields
	if len(parts) == 0 {
		return nil, ErrNoParts
	}
	seenNames := map[string]struct{}{}
	for i, part := range parts {
		if len(part.Name) == 0 {
			return nil, fmt.Errorf("no name provided for part %d", i)
		}
		if _, seen := seenNames[part.Name]; seen {
			return nil, fmt.Errorf("part %s has a duplicate name", i)
		}
		seenNames[part.Name] = struct{}{}

		if len(part.Selector) == 0 {
			return nil, fmt.Errorf("no selector provided for part %d", i)
		}
	}
	return parts, nil
}

// Create a new scraper with the provided configuration.
func NewScraper(p Payload) (*Scraper, error) {
	parts, err := p.fields2parts()
	if err != nil {
		return nil, err
	}
	var paginator paginate.Paginator
	maxPages := 1 
	if p.Paginator == nil {
		paginator = &dummyPaginator{}
		
	} else {
		paginator = paginate.BySelector(p.Paginator.Selector, p.Paginator.Attribute)
		maxPages = p.Paginator.MaxPages
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
		Request: p.Request.(splash.Request),
		DividePage: dividePageFunc,
		Parts:      parts,
		Paginator:  paginator,
		Opts: ScrapeOptions{
			MaxPages:            maxPages,
			Format:              p.Format,
			PaginateResults:     *p.PaginateResults,
			FetchDelay:          p.FetchDelay,
			RandomizeFetchDelay: *p.RandomizeFetchDelay,
			RetryTimes:          p.RetryTimes,
		},
	}

	// All set!
	return scraper, nil
}

func  Scrape(task *Task) error {
	req := task.Scraper.Request
	url := req.GetURL()

	var numPages int
	opts := task.Scraper.Opts
	task.Results.Visited = make(map[string]error)

	//get Robotstxt Data
	 robots, err := fetch.RobotstxtData(url)
	 if err != nil {
		task.Results.Visited[url] = err
		logger.Println(err)
		//return err
	 }
	 task.Scraper.Robots = robots

	for {
		//check if scraping of current url is not forbidden
		if !fetch.AllowedByRobots(url, task.Scraper.Robots) {
			task.Results.Visited[url] = &errs.ForbiddenByRobots{url}
		}
		// Repeat until we don't have any more URLs, or until we hit our page limit.
		if len(url) == 0 ||
			(opts.MaxPages > 0 && numPages >= opts.MaxPages) {
			break
		}
		//call remote fetcher to download web page
		sResponse, err := responseFromFetchService(req)
		//sResponse, err := req.(splash.Request).GetResponse()
		if err != nil {
			return err
		}
		content, err := sResponse.GetContent()
		if err != nil {
			return err
		}
		// Create a goquery document.
		doc, err := goquery.NewDocumentFromReader(content)
		if err != nil {
			return err
		}

		task.Results.Visited[url] = nil
		results := []map[string]interface{}{}

		// Divide this page into blocks
		for _, block := range task.Scraper.DividePage(doc.Selection) {
			blockResults := map[string]interface{}{}

			// Process each part of this block
			for _, part := range task.Scraper.Parts {
				sel := block
				if part.Selector != "." {
					sel = sel.Find(part.Selector)
				}

				partResults, err := part.Extractor.Extract(sel)
				if err != nil {
					return err
				}

				// A nil response from an extractor means that we don't even include it in
				// the results.
				if partResults == nil {
					continue
				}

				blockResults[part.Name] = partResults

				//********* details
				if part.Details != nil {
					//fmt.Printf("%T\n", partResults)
					part.Details.Scraper.Request = splash.Request{URL: partResults.(string)}
					//go func() {
					//	err = Scrape(part.Details)
					//	if err != nil {
					//		logger.Println(err) 
					//	}
					//}()
					err = Scrape(part.Details)
					if err != nil {
						return err
					}
					//logger.Println(part.Details.Visited)
					//logger.Println(part.Details.Results)
					//part.Details.Results = append(part.Details.Results, )
				//	logger.Println(blockResults[part.Name], part.Details)
				}
			}
			if len(blockResults) > 0 {
				// Append the results from this block.
				results = append(results, blockResults)
			}
		}

		// Append the results from this page.
		task.Results.Output = append(task.Results.Output, results)

		numPages++

		// Get the next page.
		url, err = task.Scraper.Paginator.NextPage(url, doc.Selection)
		if err != nil {
			return err
		}

		//every time when getting a response the next request will be filled with updated cookie information
		err = sResponse.SetCookieToNextRequest(&req)
		if err != nil {
			logger.Println(err)
		}
		req.URL = url

		if opts.RandomizeFetchDelay {
			//Sleep for time equal to FetchDelay * random value between 500 and 1500 msec
			rand := Random(500, 1500)
			delay := opts.FetchDelay * time.Duration(rand) / 1000
			logger.Println(delay)
			time.Sleep(delay)
		} else {
			time.Sleep(opts.FetchDelay)
		}

	}
	logger.Println(task.Visited)
	// All good!
	return nil

}

//responseFromFetchService sends request to fetch service and returns *splash.Response
func responseFromFetchService(req splash.Request) (*splash.Response, error) {

	//fetch content
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)
	addr := "http://" + viper.GetString("DFK_FETCH") + "/response/splash"
	request, err := http.NewRequest("POST", addr, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(request)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		panic(err)
	}
	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var sResponse *splash.Response
	if err := json.Unmarshal(resp, &sResponse); err != nil {
		logger.Println("Json Unmarshall error", err)
	}
	//	content, err := sResponse.GetContent()
	//	if err != nil {
	//		return nil, err
	//	}
	//return content, nil
	return sResponse, nil
}

//PartNames returns Part Names which are used as a header of output CSV
func (s Scraper) PartNames() []string {
	names := []string{}
	for _, part := range s.Parts {
		names = append(names, part.Name)
	}
	return names
}
