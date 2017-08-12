package server

import (
	"context"

	"github.com/go-kit/kit/endpoint"
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
		v, err := svc.Fetch(req)
		//logger.Println(err)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// creating Parse Endpoint
func MakeParseEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		v, err := svc.ParseData(request.([]byte))
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}