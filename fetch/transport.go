package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"

	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/sirupsen/logrus"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"
)

// NewHttpHandler mounts all of the service endpoints into an http.Handler.
func NewHttpHandler(ctx context.Context, endpoint Endpoints, logger *logrus.Logger) http.Handler {
	r := mux.NewRouter()
	r.UseEncodedPath()
	options := []httptransport.ServerOption{
		//httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("GET").Path("/proxy").HandlerFunc(proxyHandler)
	r.Methods("GET").Path("/ping").HandlerFunc(healthCheckHandler)
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
	return r
}

//DecodeSplashFetcherRequest decodes request sent to remote Splash service
//if error occures, server returns 400 Bad Request
func DecodeSplashFetcherRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request splash.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, &errs.BadRequest{err}
	}
	return request, nil
}

//DecodeBaseFetcherRequest decodes BaseFetcherRequest
//if error occures, server should return 400 Bad Request
func DecodeBaseFetcherRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request BaseFetcherRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, &errs.BadRequest{err}
	}
	return request, nil
}

//EncodeFetcherContent encodes HTML Content returned by fetcher
func EncodeFetcherContent(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fetcherContent, ok := response.(io.ReadCloser)
	if !ok {
		e := errs.BadGateway{What: "content"}
		encodeError(ctx, &e, w)
		return nil
		//encodeError(ctx, errors.New("invalid Content from Fetcher"), w)
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

//EncodeFetcherResponse encodes response returned by fetcher
func EncodeFetcherResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fetcherResponse, ok := response.(FetchResponser)
	if !ok {
		//return errors.New("invalid Fetcher Response")
		e := errs.BadGateway{What: "response"}
		encodeError(ctx, &e, w)
		return nil
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
		encodeError(ctx, err, w)
		return nil
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err = w.Write(data)

	if err != nil {
		encodeError(ctx, err, w)
		return nil
	}
	return nil
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
	//logger.Error(err)
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
		resp, err := svc.Response(req)
		if err != nil {
			return nil, err
		}
		return resp.GetHTML()
	}
}

// MakeBaseFetchEndpoint creates BaseFetch Endpoint
func MakeBaseFetchEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BaseFetcherRequest)
		resp, err := svc.Response(req)
		if err != nil {
			return nil, err
		}
		return resp.GetHTML()
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

//healthCheckHandler is used to check if Fetch service is alive.
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	u := r.URL.Query().Get("url")
	fetcher := NewFetcher(Base)
	req := BaseFetcherRequest{
		URL:    u,
		Method: "GET",
	}
	resp, err := fetcher.Response(req)
	if err != nil {
		e := errs.BadGateway{What: "response"}
		encodeError(ctx, &e, w)
		return
	}
	if err != nil {
		return
	}

	extension := filepath.Ext(u)

	var mimetype string
	switch extension {
	case ".html":
		mimetype = "text/html"
	case ".ico":
		mimetype = "image/x-icon"
	case ".jpg":
		mimetype = "image/jpeg"
	case ".png":
		mimetype = "image/png"
	case ".gif":
		mimetype = "image/gif"
	case ".css":
		mimetype = "text/css"
	case ".js":
		mimetype = "text/javascript"
	default:
		mimetype = "text/plain"
	}
	w.Header().Set("Content-Type", mimetype)
	w.WriteHeader(http.StatusOK)
	rc, err := resp.GetHTML()
	if err != nil {
		e := errs.BadGateway{What: "content"}
		encodeError(ctx, &e, w)
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	io.WriteString(w, buf.String())
}
