package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/parser"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
)

// ParseService provides operations on strings.
type ParseService interface {
	GetResponse(req splash.Request) (*splash.Response, error)
	Fetch(req splash.Request) (io.ReadCloser, error)
	ParseData(payload []byte) (io.ReadCloser, error)
	//	CheckServices() (status map[string]string)
}

type parseService struct {
	//Fetcher scrape.Fetcher
}

func (parseService) GetResponse(req splash.Request) (*splash.Response, error) {
	splashURL, err := splash.NewSplashConn(req)
	response, err := splash.GetResponse(splashURL)
	return response, err
}

func (parseService) Fetch(req splash.Request) (io.ReadCloser, error) {
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

func (parseService) ParseData(payload []byte) (io.ReadCloser, error) {
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	fetcher, err := scrape.NewSplashFetcher()
	if err != nil {
		logger.Println(err)
	}
	pieces := []scrape.Piece{}
	pl := p.Payloads[0]
	selectors := []string{}

	for _, f := range pl.Fields {
		var extractor scrape.PieceExtractor
		switch f.Extractor.Type {
		case "text":
			t := extract.Text{}
			if f.Extractor.Params != nil {
				err := t.FillStruct(f.Extractor.Params.(map[string]interface{}))
				if err != nil {
					logger.Println(err)
				}
			}
			extractor = t
			/*
				t := extract.Text{}
				if f.Extractor.Params != nil {
					err := parser.FillStruct(f.Extractor.Params.(map[string]interface{}), &t)
					if err != nil {
						logger.Println(err)
					}
				}
				extractor = t
			*/
		case "attr":
			a := extract.Attr{}
			if f.Extractor.Params != nil {
				err := parser.FillStruct(f.Extractor.Params.(map[string]interface{}), &a)

				if err != nil {
					logger.Println(err)
				}
			}
			extractor = a
		case "regex":
			r := extract.Regex{}
			if f.Extractor.Params != nil {
				err := parser.FillStruct(f.Extractor.Params.(map[string]interface{}), &r)
				if err != nil {
					logger.Println(err)
				}
				regExp := f.Extractor.Params.(map[string]interface{})["regexp"]
				//r.Regex = regexp.MustCompile(`(\d+)`)
				r.Regex = regexp.MustCompile(regExp.(string))
			}
			extractor = r
		}
		pieces = append(pieces, scrape.Piece{
			Name:      f.Name,
			Selector:  f.Selector,
			Extractor: extractor,
		})
		selectors = append(selectors, f.Selector)
	}

	paginator := pl.Paginator
	//logger.Println(paginator)
	config := &scrape.ScrapeConfig{
		Fetcher: fetcher,
		//DividePage: scrape.DividePageBySelector(".p"),
		DividePage: scrape.DividePageByIntersection(selectors),
		Pieces:     pieces,
		//Paginator: paginate.BySelector(".next", "href"),
		Paginator: paginate.BySelector(paginator.Selector, paginator.Attribute),
		Opts:      scrape.ScrapeOptions{MaxPages: paginator.MaxPages},
	}
	scraper, err := scrape.New(config)
	if err != nil {
		return nil, err
	}
	req := splash.Request{URL: pl.URL}
	results, err := scraper.Scrape(req)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(results)
	readCloser := ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	return readCloser, nil
}

/*
func (parseService) Fetch_old(req splash.Request) (io.ReadCloser, error) {
	splashURL, err := splash.NewSplashConn(req)
	content, err := splash.Fetch(splashURL)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (parseService) ParseData_old(payload []byte) (io.ReadCloser, error) {
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	res, err := p.MarshalData()
	if err != nil {
		logger.Println(res, err)
		return nil, err
	}
	return res, nil
}
*/
//func (parseService) CheckServices() (status map[string]string) {
//	return CheckServices() //, allAlive
//}

// ServiceMiddleware is a chainable behavior modifier for ParseService.
type ServiceMiddleware func(ParseService) ParseService
