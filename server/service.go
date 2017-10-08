package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/slotix/dataflowkit/robotstxt"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
)

// Define service interface
type Service interface {
	Fetch(req splash.Request) (interface{}, error)
	ParseData(payload []byte) (io.ReadCloser, error)
}

// Implement service with empty struct
type ParseService struct {
}

// create type that return function.
// this will be needed in main.go
type ServiceMiddleware func(Service) Service

//Fetch returns splash.Request
func (ParseService) Fetch(req splash.Request) (interface{}, error) {
	logger.Println("service fetch",req)
	fetcher, err := scrape.NewSplashFetcher()
	if err != nil {
		logger.Println(err)
	}
	res, err := fetcher.Fetch(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ps ParseService) ParseData(payload []byte) (io.ReadCloser, error) {
	p, err := scrape.NewPayload(payload)
	if err != nil {
		return nil, err
	}

	config, err := p.PayloadToScrapeConfig()
	if err != nil {
		return nil, err
	}
	scraper, err := scrape.New(config)
	if err != nil {
		return nil, err
	}

	req := splash.Request{URL: p.Request.URL}
	//results, err := scraper.Scrape(req, config.Opts)
	results, err := ps.scrape(req, scraper) //, config.Opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	switch config.Opts.Format {
	case "json":
		if config.Opts.PaginateResults {
			json.NewEncoder(&buf).Encode(results)
		} else {
			json.NewEncoder(&buf).Encode(results.AllBlocks())
		}
	case "csv":

		/*
			includeHeader := true
			w := csv.NewWriter(&buf)
			for i, page := range results.Results {
				if i != 0 {
					includeHeader = false
				}
				err = encodeCSV(names, includeHeader, page, ",", w)
				if err != nil {
					logger.Println(err)
				}
			}
			w.Flush()
		*/
		w := csv.NewWriter(&buf)
		err = encodeCSV(config.CSVHeader, true, results.AllBlocks(), ",", w)
		w.Flush()
	/*
		case "xmlviajson":
			var jbuf bytes.Buffer
			if config.Opts.PaginateResults {
				json.NewEncoder(&jbuf).Encode(results)
			} else {
				json.NewEncoder(&jbuf).Encode(results.AllBlocks())
			}
			//var buf bytes.Buffer
			m, err := mxj.NewMapJson(jbuf.Bytes())
			err = m.XmlIndentWriter(&buf, "", "  ")
			if err != nil {
				logger.Println(err)
			}
	*/
	case "xml":
		err = encodeXML(results.AllBlocks(), &buf)
		if err != nil {
			return nil, err
		}
	}

	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

func (ps ParseService) scrape(req splash.Request, s *scrape.Scraper) (*scrape.ScrapeResults, error) {
	url := req.URL
	if len(url) == 0 {
		return nil, errors.New("no URL provided")
	}
	//get Robotstxt Data
	robotsData, err := robotstxt.RobotsTxtData(req)
	if err != nil {
		return nil, err
	}
	res := &scrape.ScrapeResults{
		URLs:    []string{},
		Results: [][]map[string]interface{}{},
	}
	var numPages int
	//var retryTimes int
	for {
		if !robotstxt.Allowed(url, robotsData) {
			err = fmt.Errorf("%s: forbidden by robots.txt", url)
			return nil, err
		}
		// Repeat until we don't have any more URLs, or until we hit our page limit.
		if len(url) == 0 || (s.Config.Opts.MaxPages > 0 && numPages >= s.Config.Opts.MaxPages) {
			break
		}
		
		r, err := ps.Fetch(req)
	//	r, err := s.Config.Fetcher.Fetch(req)
		if err != nil {
			return nil, err
		}

		var resp io.ReadCloser
		if sResponse, ok := r.(*splash.Response); ok {
			resp, err = sResponse.GetContent()
			if err != nil {
				logger.Println(err)
			}
		}

		/*
			//fetch content
			b, err := json.Marshal(req)
			if err != nil {
				return nil, err
			}

			reader := bytes.NewReader(b)
			request, err := http.NewRequest("POST", "http://127.0.0.1:8000/app/response", reader)
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
			var sResponse splash.Response
			if err := json.Unmarshal(resp, &sResponse); err != nil {
				logger.Println("Json Unmarshall error", err)
			}
			content, err := sResponse.GetContent()
			if err != nil {
				return nil, err
			}
		*/

		// Create a goquery document.
		doc, err := goquery.NewDocumentFromReader(resp)
		//doc, err := goquery.NewDocumentFromReader(content)
		//doc, err := goquery.NewDocumentFromResponse(r)
		//resp.Close()
		if err != nil {
			return nil, err
		}
		res.URLs = append(res.URLs, url)
		results := []map[string]interface{}{}

		// Divide this page into blocks
		for _, block := range s.Config.DividePage(doc.Selection) {
			blockResults := map[string]interface{}{}

			// Process each piece of this block
			for _, piece := range s.Config.Pieces {
				//logger.Println(piece)
				sel := block
				if piece.Selector != "." {
					sel = sel.Find(piece.Selector)
				}

				pieceResults, err := piece.Extractor.Extract(sel)
				//logger.Println(attrOrDataValue(sel))
				if err != nil {
					return nil, err
				}

				// A nil response from an extractor means that we don't even include it in
				// the results.
				if pieceResults == nil {
					continue
				}

				blockResults[piece.Name] = pieceResults
			}
			if len(blockResults) > 0 {
				// Append the results from this block.
				results = append(results, blockResults)
			}
		}

		// Append the results from this page.
		res.Results = append(res.Results, results)

		numPages++

		// Get the next page.
		url, err = s.Config.Paginator.NextPage(url, doc.Selection)
		if err != nil {
			return nil, err
		}

		//every time when getting a response the next request will be filled with updated cookie information

		if sResponse, ok := r.(*splash.Response); ok {
			err := sResponse.SetCookieToRequest(&req)
			if err != nil {
				//return nil, err
				logger.Println(err)
			}
		}
		req.URL = url
		if s.Config.Opts.RandomizeFetchDelay {
			//Sleep for time equal to FetchDelay * random value between 500 and 1500 msec
			rand := scrape.Random(500, 1500)
			delay := s.Config.Opts.FetchDelay * time.Duration(rand) / 1000
			logger.Println(delay)
			time.Sleep(delay)
		} else {
			time.Sleep(s.Config.Opts.FetchDelay)
		}

	}
	// All good!
	return res, nil
}



/*
//Original
func (ps ParseService) scrape(req splash.Request, s *scrape.Scraper) (*scrape.ScrapeResults, error) {
	url := req.URL
	if len(url) == 0 {
		return nil, errors.New("no URL provided")
	}
	//get Robotstxt Data
	robotsData, err := robotstxt.RobotsTxtData(req)
	//err := scrape.AllowedByRobots(req)
	if err != nil {
		return nil, err
	}
	res := &scrape.ScrapeResults{
		URLs:    []string{},
		Results: [][]map[string]interface{}{},
	}
	var numPages int
	//var retryTimes int
	for {
		if !robotstxt.Allowed(url, robotsData) {
			err = fmt.Errorf("%s: forbidden by robots.txt", url)
			return nil, err
		}
		// Repeat until we don't have any more URLs, or until we hit our page limit.
		if len(url) == 0 || (s.Config.Opts.MaxPages > 0 && numPages >= s.Config.Opts.MaxPages) {
			break
		}
		r, err := s.Config.Fetcher.Fetch(req)
		if err != nil {
			return nil, err
		}

		var resp io.ReadCloser
		if sResponse, ok := r.(*splash.Response); ok {
			resp, err = sResponse.GetContent()
			if err != nil {
				logger.Println(err)
			}
		}

		/*
			//fetch content
			b, err := json.Marshal(req)
			if err != nil {
				return nil, err
			}

			reader := bytes.NewReader(b)
			request, err := http.NewRequest("POST", "http://127.0.0.1:8000/app/response", reader)
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
			var sResponse splash.Response
			if err := json.Unmarshal(resp, &sResponse); err != nil {
				logger.Println("Json Unmarshall error", err)
			}
			content, err := sResponse.GetContent()
			if err != nil {
				return nil, err
			}
		*/
/*
		// Create a goquery document.
		doc, err := goquery.NewDocumentFromReader(resp)
		//doc, err := goquery.NewDocumentFromReader(content)
		//doc, err := goquery.NewDocumentFromResponse(r)
		//resp.Close()
		if err != nil {
			return nil, err
		}
		res.URLs = append(res.URLs, url)
		results := []map[string]interface{}{}

		// Divide this page into blocks
		for _, block := range s.Config.DividePage(doc.Selection) {
			blockResults := map[string]interface{}{}

			// Process each piece of this block
			for _, piece := range s.Config.Pieces {
				//logger.Println(piece)
				sel := block
				if piece.Selector != "." {
					sel = sel.Find(piece.Selector)
				}

				pieceResults, err := piece.Extractor.Extract(sel)
				//logger.Println(attrOrDataValue(sel))
				if err != nil {
					return nil, err
				}

				// A nil response from an extractor means that we don't even include it in
				// the results.
				if pieceResults == nil {
					continue
				}

				blockResults[piece.Name] = pieceResults
			}
			if len(blockResults) > 0 {
				// Append the results from this block.
				results = append(results, blockResults)
			}
		}

		// Append the results from this page.
		res.Results = append(res.Results, results)

		numPages++

		// Get the next page.
		url, err = s.Config.Paginator.NextPage(url, doc.Selection)
		if err != nil {
			return nil, err
		}

		//every time when getting a response the next request will be filled with updated cookie information

		if sResponse, ok := r.(*splash.Response); ok {
			err := sResponse.SetCookieToRequest(&req)
			if err != nil {
				//return nil, err
				logger.Println(err)
			}
		}
		req.URL = url
		if s.Config.Opts.RandomizeFetchDelay {
			//Sleep for time equal to FetchDelay * random value between 500 and 1500 msec
			rand := scrape.Random(500, 1500)
			delay := s.Config.Opts.FetchDelay * time.Duration(rand) / 1000
			logger.Println(delay)
			time.Sleep(delay)
		} else {
			time.Sleep(s.Config.Opts.FetchDelay)
		}

	}
	// All good!
	return res, nil
}
*/