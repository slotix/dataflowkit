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

//DecodeSplashFetcherRequest decodes request sent to remote Splash service
//if error occures, server returns 400 Bad Request
func DecodeSplashFetcherRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request splash.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, &errs.BadRequest{err} //err
	}
	return request, nil
}

//DecodeBaseFetcherRequest decodes BaseFetcherRequest
//if error occures, server should return 400 Bad Request
func DecodeBaseFetcherRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request BaseFetcherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, &errs.BadRequest{err} //err
	}
	return request, nil
}

//EncodeSplashFetcherContent encodes HTML Content returned by Splash service endpoint
func EncodeSplashFetcherContent(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse, ok := response.(*splash.Response)
	if !ok {
		encodeError(ctx, errors.New("invalid SplashFetcher Response"), w)
		//return errors.New("invalid Splash Response")
		return nil
	}
	if fetcherResponse.Error != "" {
		encodeError(ctx, errors.New(fetcherResponse.Error), w)
		//return errors.New(fetcherResponse.Error)
		return nil
	}
	content, err := fetcherResponse.GetContent()
	if err != nil {
		encodeError(ctx, err, w)
		return nil
		//return err
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err = io.Copy(w, content)
	if err != nil {
		encodeError(ctx, err, w)
		return nil
		//return err
	}
	return nil
}

//EncodeBaseFetcherContent encodes HMTL Content returned by Base Fetcher
func EncodeBaseFetcherContent(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse, ok := response.(*BaseFetcherResponse)
	if !ok {
		encodeError(ctx, errors.New("invalid BaseFetcher Response"), w)
		//return errors.New("invalid Base Response")
		return nil

	}
	if fetcherResponse.StatusCode != 200 {
		encodeError(ctx, errors.New(fetcherResponse.Status), w)
		//return errors.New(fetcherResponse.Status)
		return nil
	}
	r := bytes.NewReader(fetcherResponse.HTML)
	readCloser := ioutil.NopCloser(r)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err := io.Copy(w, readCloser)
	if err != nil {
		return err
	}

	return nil
}

//EncodeSplashFetcherContent encodes response returned by Splash service endpoint
func EncodeSplashFetcherResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
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

//EncodeBaseFetcherResponse encodes response returned by Base Fetcher
func EncodeBaseFetcherResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherResponse := response.(*BaseFetcherResponse)
	if fetcherResponse.StatusCode != 200 {
		return errors.New(fetcherResponse.Status)
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

// encodeError encodes erroneous responses and writes http status header.
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
	case *errs.BadGateway:
		//return 502 Status
		httpStatus = http.StatusBadGateway
	case *errs.GatewayTimeout:
		//return 504 Status
		httpStatus = http.StatusGatewayTimeout
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

// Endpoints wrapper
type Endpoints struct {
	SplashFetchEndpoint    endpoint.Endpoint
	SplashResponseEndpoint endpoint.Endpoint
	BaseFetchEndpoint      endpoint.Endpoint
	BaseResponseEndpoint   endpoint.Endpoint
}

// MakeSplashFetchEndpoint creates Splash Fetch Endpoint
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

// MakeBaseFetchEndpoint creates BaseFetch Endpoint
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

// MakeSplashResponseEndpoint creates SplashResponse Endpoint
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

// MakeBaseResponseEndpoint creates BaseResponse Endpoint
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

//healthCheckHandler is used to check if Fetch service is alive.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

// MakeHttpHandler mounts all of the service endpoints into an http.Handler.
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
		DecodeSplashFetcherRequest,
		EncodeSplashFetcherContent,
		options...,
	))

	r.Methods("POST").Path("/fetch/base").Handler(httptransport.NewServer(
		endpoint.BaseFetchEndpoint,
		DecodeBaseFetcherRequest,
		EncodeBaseFetcherContent,
		options...,
	))

	r.Methods("POST").Path("/response/splash").Handler(httptransport.NewServer(
		endpoint.SplashResponseEndpoint,
		DecodeSplashFetcherRequest,
		EncodeSplashFetcherResponse,
		options...,
	))
	r.Methods("POST").Path("/response/base").Handler(httptransport.NewServer(
		endpoint.BaseResponseEndpoint,
		DecodeBaseFetcherRequest,
		EncodeBaseFetcherResponse,
		options...,
	))
	r.Methods("GET").Path("/ping").HandlerFunc(healthCheckHandler)

	return r
}
