package parse

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// endpoints wrapper
type Endpoints struct {
	ParseEndpoint endpoint.Endpoint
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
