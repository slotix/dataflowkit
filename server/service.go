package server

import (
	"github.com/slotix/dataflowkit/downloader"
	"github.com/slotix/dataflowkit/parser"
)

// ParseService provides operations on strings.
type ParseService interface {
	GetResponse(string) (*downloader.SplashResponse, error)
	Download(string) ([]byte, error)
	ParseData(payload []byte) ([]byte, error)
//	CheckServices() (status map[string]string)
	//	Save(payload []byte) (string, error)
}

type parseService struct{}

func (parseService) GetResponse(url string) (*downloader.SplashResponse, error) {
	response, err := downloader.GetResponse(url)
	return response, err
}

func (parseService) Download(url string) ([]byte, error) {
	//defer func(begin time.Time) {
	//	fmt.Println("took", time.Since(begin))
	//}(time.Now())
	content, err := downloader.Download(url)
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
