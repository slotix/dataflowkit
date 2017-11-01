package fetch

import (
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
)

// Define service interface
type Service interface {
	Fetch(req interface{}) (interface{}, error)
	Response(req interface{}) (interface{}, error)
}

// Implement service with empty struct
type FetchService struct {
}

// create type that return function.
// this will be needed in main.go
type ServiceMiddleware func(Service) Service


//Fetch returns splash.Response
//see transport.go encodeFetchResponse for more details about retured value.

func (fs FetchService) Fetch(req interface{}) (interface{}, error) {
	res, err := fs.Response(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//Response returns splash.Response
//see transport.go encodeResponse for more details about retured value.
func (fs FetchService) Response(req interface{}) (interface{}, error) {
	request := req.(splash.Request)
	fetcher, err := scrape.NewSplashFetcher()
	if err != nil {
		logger.Println(err)
	}
	res, err := fetcher.Fetch(request)
	if err != nil {
		return nil, err
	}
	return res, nil
}
