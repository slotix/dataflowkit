package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/slotix/dataflowkit/parser"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/goscrape/extract"
	"github.com/slotix/goscrape/paginate"
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
		pieces = append(pieces, scrape.Piece{
			Name:      f.Name,
			Selector:  f.Selector,
			Extractor: extract.Text{},
		})
		selectors = append(selectors, f.Selector)
	}

	paginator := pl.Paginator
	logger.Println(paginator)
	config := &scrape.ScrapeConfig{
		Fetcher:    fetcher,
		//DividePage: scrape.DividePageBySelector(".p"),
		DividePage: scrape.DividePageByIntersection(selectors),
		Pieces:    pieces,
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

//func (parseService) CheckServices() (status map[string]string) {
//	return CheckServices() //, allAlive
//}

// ServiceMiddleware is a chainable behavior modifier for ParseService.
type ServiceMiddleware func(ParseService) ParseService
