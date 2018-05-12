package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/slotix/dataflowkit/splash"
)

// NewHTTPClient returns an Fetch Service backed by an HTTP server living at the
// remote instance. We expect instance to come from a service discovery system,
// so likely of the form "host:port". We bake-in certain middlewares,
// implementing the client library pattern.
func NewHTTPClient(instance string) (Service, error) {
	// Quickly sanitize the instance string.
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	// var splashFetchEndpoint endpoint.Endpoint
	// {
	// 	splashFetchEndpoint = httptransport.NewClient(
	// 		"POST",
	// 		copyURL(u, "/fetch/splash"),
	// 		encodeHTTPGenericRequest,
	// 		decodeSplashFetcherContent,
	// 	).Endpoint()
	// }

	var splashResponseEndpoint endpoint.Endpoint
	{
		splashResponseEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/response/splash"),
			encodeHTTPGenericRequest,
			decodeSplashFetcherResponse,
		).Endpoint()
	}

	// var baseFetchEndpoint endpoint.Endpoint
	// {
	// 	baseFetchEndpoint = httptransport.NewClient(
	// 		"POST",
	// 		copyURL(u, "/fetch/base"),
	// 		encodeHTTPGenericRequest,
	// 		decodeBaseFetcherContent,
	// 	).Endpoint()
	// }

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
		//SplashFetchEndpoint:    splashFetchEndpoint,
		SplashResponseEndpoint: splashResponseEndpoint,
		//BaseFetchEndpoint:      baseFetchEndpoint,
		BaseResponseEndpoint: baseResponseEndpoint,
	}, nil
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(ctx context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// decodeSplashFetcherContent is a transport/http.DecodeResponseFunc that decodes a
// JSON-encoded splash fetcher response from the HTTP response body. If the response has a
// non-200 status code, we will interpret that as an error and attempt to decode
// the specific error message from the response body. Primarily useful in a
// client.
// func decodeSplashFetcherContent(ctx context.Context, r *http.Response) (interface{}, error) {
// 	if r.StatusCode != http.StatusOK {
// 		return nil, errors.New(r.Status)
// 	}
// 	data, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

func decodeSplashFetcherResponse(ctx context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	var resp splash.Response
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// func decodeBaseFetcherContent(ctx context.Context, r *http.Response) (interface{}, error) {
// 	if r.StatusCode != http.StatusOK {
// 		return nil, errors.New(r.Status)
// 	}
// 	data, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

func decodeBaseFetcherResponse(ctx context.Context, r *http.Response) (interface{}, error) {
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

// func (e Endpoints) Fetch(req FetchRequester) (io.ReadCloser, error) {
// 	// ctx := context.Background()
// 	// var resp interface{}
// 	// var err error
// 	// switch req.Type() {
// 	// case "base":
// 	// 	resp, err = e.BaseFetchEndpoint(ctx, req)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// case "splash":
// 	// 	resp, err = e.SplashFetchEndpoint(ctx, req)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// }
// 	// readCloser := ioutil.NopCloser(bytes.NewReader(resp.([]byte)))
// 	// return readCloser, nil
// 	resp, err := e.Response(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp.GetHTML()
// }

func (e Endpoints) Response(req FetchRequester) (FetchResponser, error) {
	ctx := context.Background()
	var r FetchResponser
	var err error
	switch req.Type() {
	case "base":
		resp, err := e.BaseResponseEndpoint(ctx, req)
		if err != nil {
			return nil, err
		}
		response := resp.(BaseFetcherResponse)
		r = &response
	//case splash.Request, *splash.Request:
	case "splash":
		resp, err := e.SplashResponseEndpoint(ctx, req)
		if err != nil {
			return nil, err
		}
		response := resp.(splash.Response)
		r = &response
	}
	return r, err
}
