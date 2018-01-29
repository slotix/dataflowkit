package fetch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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

//EncodeFetcherContent encodes HTML Content returned by SplashFetcher/ BaseFetcher
func EncodeFetcherContent(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	fetcherContent, ok := response.(io.ReadCloser)
	if !ok {
		encodeError(ctx, errors.New("invalid Content from SplashFetcher"), w)
		//return errors.New("invalid Splash Response")
		return nil
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err := io.Copy(w, fetcherContent)
	if err != nil {
		encodeError(ctx, err, w)
		return nil
		//return err
	}
	return nil
}

//EncodeSplashFetcherResponse encodes response returned by Splash server
func EncodeFetcherResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	//	fetcherResponse, ok := response.(*splash.Response)
	fetcherResponse, ok := response.(FetchResponser)
	if !ok {
		return errors.New("invalid Fetcher Response")
	}
	//statusCode := fetcherResponse.GetStatusCode()
	//if statusCode != 200 {
	//	return errors.New(strconv.Itoa(statusCode))
	//}
	//if fetcherResponse.Error != "" {
	//	return errors.New(fetcherResponse.Error)
	//}

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

// NewHttpHandler mounts all of the service endpoints into an http.Handler.
func NewHttpHandler(ctx context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
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
		EncodeFetcherContent,
		options...,
	))

	r.Methods("POST").Path("/fetch/base").Handler(httptransport.NewServer(
		endpoint.BaseFetchEndpoint,
		DecodeBaseFetcherRequest,
		EncodeFetcherContent,
		options...,
	))

	r.Methods("POST").Path("/response/splash").Handler(httptransport.NewServer(
		endpoint.SplashResponseEndpoint,
		DecodeSplashFetcherRequest,
		EncodeFetcherResponse,
		options...,
	))
	r.Methods("POST").Path("/response/base").Handler(httptransport.NewServer(
		endpoint.BaseResponseEndpoint,
		DecodeBaseFetcherRequest,
		EncodeFetcherResponse,
		options...,
	))
	r.Methods("GET").Path("/ping").HandlerFunc(healthCheckHandler)

	return r
}

// NewHTTPClient returns an AddService backed by an HTTP server living at the
// remote instance. We expect instance to come from a service discovery system,
// so likely of the form "host:port". We bake-in certain middlewares,
// implementing the client library pattern.
func NewHTTPClient(instance string, logger log.Logger) (Service, error) {
	//func NewHTTPClient(instance string, logger log.Logger) (Service, error) {
	// Quickly sanitize the instance string.
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	//	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var splashFetchEndpoint endpoint.Endpoint
	{
		splashFetchEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/fetch/splash"),
			encodeHTTPGenericRequest,
			decodeSplashFetcherContent,
		).Endpoint()
		//	splashFetchEndpoint = limiter(splashFetchEndpoint)
		//	splashFetchEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
		//		Name:    "Splash Fetch",
		//		Timeout: 30 * time.Second,
		//	}))(splashFetchEndpoint)
	}

	var splashResponseEndpoint endpoint.Endpoint
	{
		splashResponseEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/response/splash"),
			encodeHTTPGenericRequest,
			decodeSplashFetcherResponse,
		).Endpoint()
	}

	var baseFetchEndpoint endpoint.Endpoint
	{
		baseFetchEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/fetch/base"),
			encodeHTTPGenericRequest,
			decodeBaseFetcherContent,
		).Endpoint()
	}

	var baseResponseEndpoint endpoint.Endpoint
	{
		baseResponseEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/response/base"),
			encodeHTTPGenericRequest,
			decodeBaseFetcherResponse,
		).Endpoint()
	}
	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return Endpoints{
		SplashFetchEndpoint:    splashFetchEndpoint,
		SplashResponseEndpoint: splashResponseEndpoint,
		BaseFetchEndpoint:      baseFetchEndpoint,
		BaseResponseEndpoint:   baseResponseEndpoint,
		//	ConcatEndpoint: concatEndpoint,
	}, nil
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// decodeSplashFetcherContent is a transport/http.DecodeResponseFunc that decodes a
// JSON-encoded sum response from the HTTP response body. If the response has a
// non-200 status code, we will interpret that as an error and attempt to decode
// the specific error message from the response body. Primarily useful in a
// client.
func decodeSplashFetcherContent(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	//logger.Info(string(data))
	return data, nil
	//sResponse := &splash.Response{}

	//if err := json.Unmarshal(data, &sResponse); err != nil {
	//	return nil, err
	//}
	//return sResponse, nil
	//var resp splash.Response
	//err := json.NewDecoder(r.Body).Decode(&resp)
	//return resp, err
	//reader := r.Body
}

func decodeSplashFetcherResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp splash.Response
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeBaseFetcherContent(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp BaseFetcherResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func decodeBaseFetcherResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp BaseFetcherResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func (e Endpoints) Fetch(req FetchRequester) (io.ReadCloser, error) {
	/* ctx := context.Background()
	resp, err := e.SplashFetchEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}
	response := resp.(FetchResponser)
	return response, nil */
	r, err := e.Response(req)
	if err != nil {
		return nil, err
	}
	return r.GetHTML()
}

func (e Endpoints) Response(req FetchRequester) (FetchResponser, error) {
	ctx := context.Background()
	resp, err := e.SplashResponseEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}
	response := resp.(splash.Response)
	return &response, nil
}
