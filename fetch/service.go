package fetch

import "github.com/slotix/dataflowkit/splash"

// Define service interface
type Service interface {
	Fetch(req splash.Request) (interface{}, error)
}

// Implement service with empty struct
type FetchService struct {
}

// create type that return function.
// this will be needed in main.go
type ServiceMiddleware func(Service) Service

//Fetch returns splash.Request
func (FetchService) Fetch(req splash.Request) (interface{}, error) {
	fetcher, err := NewSplashFetcher()
	if err != nil {
		logger.Println(err)
	}
	res, err := fetcher.Fetch(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

