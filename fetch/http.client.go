package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
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

	var baseFetchEndpoint endpoint.Endpoint
	{
		baseFetchEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/fetch/base"),
			encodeRequest,
			decodeBaseFetcherContent,
		).Endpoint()
	}

	var chromeFetchEndpoint endpoint.Endpoint
	{
		chromeFetchEndpoint = httptransport.NewClient(
			"POST",
			copyURL(u, "/fetch/chrome"),
			encodeRequest,
			decodeChromeFetcherContent,
		).Endpoint()
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return Endpoints{
		BaseFetchEndpoint:   baseFetchEndpoint,
		ChromeFetchEndpoint: chromeFetchEndpoint,
	}, nil
}

// encodeRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeRequest(ctx context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func decodeBaseFetcherContent(ctx context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func decodeChromeFetcherContent(ctx context.Context, r *http.Response) (interface{}, error) {
	//Chrome returns no error for pages with non 200 pages. So we don't need to this check 
	// if r.StatusCode != http.StatusOK {
	// 	return nil, errors.New(r.Status)
	// }
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func copyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func (e Endpoints) Fetch(req Request) (io.ReadCloser, error) {
	ctx := context.Background()
	var resp interface{}
	var err error
	switch req.Type {
	case "base":
		resp, err = e.BaseFetchEndpoint(ctx, req)
		if err != nil {
			return nil, err
		}
	case "chrome":
		resp, err = e.ChromeFetchEndpoint(ctx, req)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid fetcher type specified %s", req.Type)
	}
	readCloser := ioutil.NopCloser(bytes.NewReader(resp.([]byte)))
	return readCloser, nil
}
