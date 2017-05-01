package server

import (
	"io"

	"github.com/slotix/dataflowkit/downloader"
	"github.com/slotix/dataflowkit/parser"
)

// ParseService provides operations on strings.
type ParseService interface {
	GetResponse(req downloader.FetchRequest) (*downloader.SplashResponse, error)
	Fetch(req downloader.FetchRequest) (io.ReadCloser, error)
	ParseData(payload []byte) ([]byte, error)
	//	CheckServices() (status map[string]string)
}

type parseService struct{}

func (parseService) GetResponse(req downloader.FetchRequest) (*downloader.SplashResponse, error) {
	response, err := downloader.GetResponse(req)
	return response, err
}

func (parseService) Fetch(req downloader.FetchRequest) (io.ReadCloser, error) {
	content, err := downloader.Fetch(req)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (parseService) ParseData(payload []byte) ([]byte, error) {
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
