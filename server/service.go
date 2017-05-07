package server

import (
	"io"

	"github.com/slotix/dataflowkit/parser"
	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/scrape"
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

func (parseService) Fetch1(req splash.Request) (io.ReadCloser, error) {
	splashURL, err := splash.NewSplashConn(req)
	content, err := splash.Fetch(splashURL)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (parseService) ParseData(payload []byte) (io.ReadCloser, error) {
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
