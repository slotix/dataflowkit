package server

import (
	"github.com/slotix/dataflowkit/downloader"
	"github.com/slotix/dataflowkit/parser"
)

// ParseService provides operations on strings.
type ParseService interface {
	GetResponse(string) (*downloader.SplashResponse, error)
	Download(url string) ([]byte, error)
	ParseData(payload []byte) ([]byte, error)
	//	CheckServices() (status map[string]string)
	//	Save(payload []byte) (string, error)

}
/*
type Params []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type fetchRequest struct {
	URL string `json:"url"`
	Params Params `json:"params,omitempty"`
	Func string `json:"func,omitempty"`
}
*/

type parseService struct{}

func (parseService) GetResponse(url string) (*downloader.SplashResponse, error) {
	response, err := downloader.GetResponse(url)
	return response, err
}

func (parseService) Download(url string ) ([]byte, error) {
//func (parseService) Download(request []byte) ([]byte, error) {
	
	content, err := downloader.Fetch(url)
	if err != nil {
		return nil, err
	}
	return content, nil

	//	return nil, fmt.Errorf("%s: forbidden by robots.txt", url)
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
