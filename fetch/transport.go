package fetch

import (
	"encoding/json"
	"io"
	"net/http"

	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/sirupsen/logrus"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
)

// newHttpHandler mounts all of the service endpoints into an http.Handler.
func newHttpHandler(ctx context.Context, endpoint endpoints, logger *logrus.Logger) http.Handler {
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
//if error occures, server should return 400 Bad Request
func decodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, &errs.BadRequest{err}
	}
	return request, nil
}

//EncodeFetcherContent encodes HTML Content returned by fetcher
func encodeFetcherContent(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fetcherContent, ok := response.(io.ReadCloser)
	if !ok {
		e := errs.BadGateway{What: "content"}
		encodeError(ctx, &e, w)
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
	if err == nil {
		panic("encodeError with nil error")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var httpStatus int
	switch err.(type) {
	default:
		httpStatus = http.StatusInternalServerError
	case *errs.BadRequest,
		*errs.Error:
		//return 400 Status
		httpStatus = http.StatusBadRequest
	case *errs.Unauthorized:
		//return 401 Status
		httpStatus = http.StatusUnauthorized
	case *errs.ForbiddenByRobots,
		*errs.Forbidden:
		//return 403 Status
		httpStatus = http.StatusForbidden
	case *errs.ProxyAuthenticationRequired:
		//return 407 Status
		httpStatus = http.StatusProxyAuthRequired
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
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
