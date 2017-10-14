package fetch

import (
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
)

// Define service interface
type Service interface {
	Fetch(req interface{}) (interface{}, error)
	getURL(req interface{}) string
	Response(req interface{}) (interface{}, error)
}

// Implement service with empty struct
type FetchService struct {
}

// create type that return function.
// this will be needed in main.go
type ServiceMiddleware func(Service) Service

func (fs FetchService) getURL(req interface{}) string {
	var url string
	switch req.(type) {
	case splash.Request:
		url = req.(splash.Request).URL
	case scrape.HttpClientFetcherRequest:
		url = req.(scrape.HttpClientFetcherRequest).URL
	}
	return url
}

//Fetch returns splash.Response
//see transport.go encodeFetchResponse for more details about retured values. 
func (fs FetchService) Fetch(req interface{}) (interface{}, error) {
	// request := req.(splash.Request)
	// fetcher, err := scrape.NewSplashFetcher()
	// if err != nil {
	// 	logger.Println(err)
	// }
	// res, err := fetcher.Fetch(request)
	// if err != nil {
	// 	return nil, err
	// }
	//return res, nil
	
	res, err := fs.Response(req)
	content, err := res.(*splash.Response).GetContent()
	if err != nil {
		return nil, err
	}
	//data, err := ioutil.ReadAll(content)

	//if err != nil {
	//	return err
	//}
	//return res, nil
	return content, nil
}

//Response returns splash.Response
//see transport.go encodeResponse for more details about retured values. 
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