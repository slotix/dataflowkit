package fetch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"context"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"
)

//decodeFetchRequest
//if error is not nil, server should return
//400 Bad Request
func DecodeFetchRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request splash.Request
	//var request scrape.HttpClientFetcherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		//logger.Printf("Type: %T\n", err)
		return nil, &errs.BadRequest{err} //err
	}
	return request, nil
}

func EncodeFetchResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	sResponse, ok := response.(*splash.Response)
	if !ok {
		return errors.New("invalid Splash Response")
	}
	if sResponse.Error != "" {
		return errors.New(sResponse.Error)
	}
	content, err := sResponse.GetContent()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, content)
	if err != nil {
		return err
	}
	return nil
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	sResponse := response.(*splash.Response)
	if sResponse.Error != "" {
		return errors.New(sResponse.Error)
	}

	data, err := json.Marshal(sResponse)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	if err != nil {
		return err
	}
	return nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error.
type errorer interface {
	error() error
}

// encode error
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
//	logger.Printf("Type: %T\n", err)

	var httpStatus int
	switch err.(type) {
	default:
		httpStatus = http.StatusInternalServerError
	case *errs.BadRequest,
		*errs.InvalidHost,
		*errs.Error:
		//return 400 Status
		httpStatus = http.StatusBadRequest
	case *errs.ForbiddenByRobots,
		*errs.Forbidden:
		//return 403 Status
		httpStatus = http.StatusForbidden
	case *errs.NotFound:
		//return 404 Status
		httpStatus = http.StatusNotFound
	case *errs.GatewayTimeout:
		//return 504 Status
		httpStatus = http.StatusGatewayTimeout
	}

	w.WriteHeader(httpStatus)
	//AWS error payload should looks like
	//{
	//"errorType": "BadRequest",
	//"httpStatus": httpStatus,
	//"requestId" : "<context.awsRequestId>",
	//"message": err.Error(),
	//}
	//according to the information from https://aws.amazon.com/blogs/compute/error-handling-patterns-in-amazon-api-gateway-and-aws-lambda/

	//But it seems enough to w.WriteHeader(httpStatus) and send an error only
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// endpoints wrapper
type Endpoints struct {
	FetchEndpoint    endpoint.Endpoint
	ResponseEndpoint endpoint.Endpoint
	//ParseEndpoint endpoint.Endpoint
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



// creating Response Endpoint
func MakeResponseEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(splash.Request)
		//req := request
		v, err := svc.Response(req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    io.WriteString(w, `{"alive": true}`)
}

// Make Http Handler
func MakeHttpHandler(ctx context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	/*
		router := httprouter.New()
		var svc Service
		fetchHandler := httptransport.NewServer(
			MakeFetchEndpoint(svc),
			decodeFetchRequest,
			encodeFetchResponse,
		)
		router.Handler("POST", "/app/fetch", fetchHandler)
		return router
	*/
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//"/app/parse"
	r.Methods("POST").Path("/fetch").Handler(httptransport.NewServer(
		endpoint.FetchEndpoint,
		DecodeFetchRequest,
		EncodeFetchResponse,
		options...,
	))

	r.Methods("POST").Path("/response").Handler(httptransport.NewServer(
		endpoint.ResponseEndpoint,
		DecodeFetchRequest,
		EncodeResponse,
		options...,
	))
	r.Methods("GET").Path("/ping").HandlerFunc(HealthCheckHandler)

	return r
}
