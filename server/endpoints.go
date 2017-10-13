package server

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
)

// endpoints wrapper
type Endpoints struct {
	FetchEndpoint endpoint.Endpoint
	ParseEndpoint endpoint.Endpoint
}

// creating Fetch Endpoint
func MakeFetchEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(splash.Request)
		//req := request
		v, err := svc.Fetch(req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// creating Parse Endpoint
func MakeParseEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		v, err := svc.ParseData(request.(scrape.Payload))
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
