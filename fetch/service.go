package fetch

import (
	"github.com/slotix/dataflowkit/splash"
)

// Define fetch service interface
type Service interface {
	Fetch(req FetchRequester) (FetchResponser, error)
	Response(req FetchRequester) (FetchResponser, error)
}

// Implement service with empty struct
type FetchService struct {
}


type ServiceMiddleware func(Service) Service

//Fetch returns splash.Response
//see transport.go encodeFetchResponse for more details about retured value.  
func (fs FetchService) Fetch(req FetchRequester) (FetchResponser, error) {
	res, err := fs.Response(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//Response returns splash.Response
//see transport.go encodeResponse for more details about retured value.
func (fs FetchService) Response(req FetchRequester) (FetchResponser, error) {

	var err error
	var fetcher Fetcher
	switch req.(type) {
	case BaseFetcherRequest:
		fetcher, err = NewFetcher(Base)
	case splash.Request:
		fetcher, err = NewFetcher(Splash)
	default:
		panic("invalid fetcher request")
	}
	
	if err != nil {
		logger.Println(err)
	}
	res, err := fetcher.Fetch(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
