package parse

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/endpoint"

	"context"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/scrape"
)

//DecodeParseRequest decodes request sent to Parser
//if error occures, server returns 400 Bad Request
func DecodeParseRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var p scrape.Payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return nil, err
	}
	return p, nil
}

//EncodeParseResponse encodes response returned by Parser
func EncodeParseResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	ctx := context.Background()
	data, err := ioutil.ReadAll(response.(io.Reader))
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch e := err.(type) {
	case errs.Error:
		// We can retrieve the status here and write out a specific
		// HTTP status code.
		http.Error(w, e.Error(), e.Status())
	//case errs.BadPayload:
	//	http.Error(w, e.Error(), e.Status())
	default:
		// Any error types we don't specifically look out for default
		// to serving a HTTP 500
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
	}
}

// Endpoints wrapper
type Endpoints struct {
	ParseEndpoint endpoint.Endpoint
}

// MakeParseEndpoint creates Parse Endpoint
func MakeParseEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		v, err := svc.Parse(request.(scrape.Payload))
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

//HealthCheckHandler is used to check if Parse service is alive
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

// NewHttpHandler mounts all of the service endpoints into an http.Handler.
func NewHttpHandler(ctx context.Context, endpoint Endpoints) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/parse").Handler(httptransport.NewServer(
		endpoint.ParseEndpoint,
		DecodeParseRequest,
		EncodeParseResponse,
		options...,
	))

	r.Methods("GET").Path("/ping").HandlerFunc(HealthCheckHandler)
	return r
}
