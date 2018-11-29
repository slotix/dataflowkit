package fetch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
)

// newHttpHandler mounts all of the service endpoints into an http.Handler.
func newHttpHandler(ctx context.Context, endpoint endpoints) http.Handler {
	r := mux.NewRouter()
	r.UseEncodedPath()
	options := []httptransport.ServerOption{
		//httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}
	r.Methods("GET").Path("/ping").HandlerFunc(healthCheckHandler)
	r.Methods("POST").Path("/fetch").Handler(httptransport.NewServer(
		endpoint.fetchEndpoint,
		decodeRequest,
		encodeFetcherContent,
		options...,
	))
	return r
}

//DecodeRequest decodes FetcherRequest
func decodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

//EncodeFetcherContent encodes HTML Content returned by fetcher
func encodeFetcherContent(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fetcherContent, ok := response.(io.ReadCloser)
	if !ok {
		e := errors.New(http.StatusText(http.StatusBadGateway))
		encodeError(ctx, e, w)
		return nil
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err := io.Copy(w, fetcherContent)
	if err != nil {
		encodeError(ctx, err, w)
		return nil
	}
	return nil
}

// encodeError encodes erroneous responses and writes http status header.
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch e := err.(type) {
	case errs.Error:
		// We can retrieve the status here and write out a specific
		// HTTP status code.
		http.Error(w, e.Error(), e.Status())
	/* case context.DeadlineExceeded:
	http.Error(w, "Fetch timeout", 400) */
	default:
		// Any error types we don't specifically look out for default
		// to serving a HTTP 500
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

// endpoints wrapper
type endpoints struct {
	fetchEndpoint endpoint.Endpoint
}

// MakeFetchEndpoint creates Fetch Endpoint
func makeFetchEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return svc.Fetch(request.(Request))
	}
}

//healthCheckHandler is used to check if Fetch service is alive.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}
