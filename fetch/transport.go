package fetch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"
)

//decodeFetchRequest
//if error is not nil, server should return
//400 Bad Request
func DecodeSplashFetchRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request splash.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		//logger.Printf("Type: %T\n", err)
		return nil, &errs.BadRequest{err} //err
	}
	return request, nil
}

//decodeFetchRequest
//if error is not nil, server should return
//400 Bad Request
func DecodeBaseFetchRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request BaseFetcherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, &errs.BadRequest{err} //err
	}
	return request, nil
}

func EncodeSplashFetchResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse, ok := response.(*splash.Response)
	if !ok {
		return errors.New("invalid Splash Response")
	}
	if fetcherResponse.Error != "" {
		return errors.New(fetcherResponse.Error)
	}
	content, err := fetcherResponse.GetContent()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, content)
	if err != nil {
		return err
	}
	return nil
}

func EncodeBaseFetchResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse, ok := response.(*BaseFetcherResponse)
	if !ok {
		return errors.New("invalid HttpClient Response")
	}
	if fetcherResponse.StatusCode != 200 {
		return errors.New(fetcherResponse.Status)
	}
	r := bytes.NewReader(fetcherResponse.HTML)
	readCloser := ioutil.NopCloser(r)

	_, err := io.Copy(w, readCloser)
	if err != nil {
		return err
	}

	return nil
}

func EncodeSplashResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse := response.(*splash.Response)
	if fetcherResponse.Error != "" {
		return errors.New(fetcherResponse.Error)
	}

	data, err := json.Marshal(fetcherResponse)
	if err != nil {
		return err
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err = w.Write(data)

	if err != nil {
		return err
	}
	return nil
}

func EncodeBaseResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse := response.(*BaseFetcherResponse)
	if fetcherResponse.Response.StatusCode != 200 {
		return errors.New(fetcherResponse.Response.Status)
	}
	
	data, err := json.Marshal(fetcherResponse)
	if err != nil {
		return err
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	SplashFetchEndpoint    endpoint.Endpoint
	SplashResponseEndpoint endpoint.Endpoint
	BaseFetchEndpoint      endpoint.Endpoint
	BaseResponseEndpoint   endpoint.Endpoint
	//ParseEndpoint endpoint.Endpoint
}

// creating Fetch Endpoint
func MakeSplashFetchEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(splash.Request)
		v, err := svc.Fetch(req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// creating Fetch Endpoint
func MakeBaseFetchEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BaseFetcherRequest)
		v, err := svc.Fetch(req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// creating Response Endpoint
func MakeSplashResponseEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(splash.Request)
		v, err := svc.Response(req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

// creating Response Endpoint
func MakeBaseResponseEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BaseFetcherRequest)
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

	r.Methods("POST").Path("/fetch/splash").Handler(httptransport.NewServer(
		endpoint.SplashFetchEndpoint,
		DecodeSplashFetchRequest,
		EncodeSplashFetchResponse,
		options...,
	))

	r.Methods("POST").Path("/fetch/base").Handler(httptransport.NewServer(
		endpoint.BaseFetchEndpoint,
		DecodeBaseFetchRequest,
		EncodeBaseFetchResponse,
		options...,
	))

	r.Methods("POST").Path("/response/splash").Handler(httptransport.NewServer(
		endpoint.SplashResponseEndpoint,
		DecodeSplashFetchRequest,
		EncodeSplashResponse,
		options...,
	))
	r.Methods("POST").Path("/response/base").Handler(httptransport.NewServer(
		endpoint.BaseResponseEndpoint,
		DecodeBaseFetchRequest,
		EncodeBaseResponse,
		options...,
	))
	r.Methods("GET").Path("/ping").HandlerFunc(HealthCheckHandler)

	return r
}
