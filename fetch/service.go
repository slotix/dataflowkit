package fetch

import (
	"io"

	"github.com/slotix/dataflowkit/splash"
)

// Service defines Fetch service interface
type Service interface {
	Response(req FetchRequester) (FetchResponser, error)
	Fetch(req FetchRequester) (io.ReadCloser, error)
}

// FetchService implements service with empty struct
type FetchService struct {
}

// ServiceMiddleware defines a middleware for a Fetch service
type ServiceMiddleware func(Service) Service

//Response returns splash.Response
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
		logger.Error(err)
	}
	//res, err := fetcher.Fetch(req)
	res, err := fetcher.Response(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//Fetch downloads web page content and returns it
func (fs FetchService) Fetch(req FetchRequester) (io.ReadCloser, error) {
	res, err := fs.Response(req)
	if err != nil {
		return nil, err
	}
	return res.GetHTML()
}
