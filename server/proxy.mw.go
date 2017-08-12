package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/slotix/dataflowkit/splash"
)

type proxyingMiddleware struct {
	ctx   context.Context
	next  Service
	fetch endpoint.Endpoint
}


func (mw proxyingMiddleware) Fetch(req splash.Request) (interface{}, error) {
	res, err := mw.fetch(mw.ctx, req)
	
	//res, err := mw.next.Fetch(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (mw proxyingMiddleware) ParseData(payload []byte) (io.ReadCloser, error) {
	return mw.next.ParseData(payload)
}

func ProxyingMiddleware(ctx context.Context, proxyURL string) ServiceMiddleware {
	if proxyURL == "" {
		logger.Println("proxy_to", "none")
		return func(next Service) Service { return next }
	}
	return func(next Service) Service {
		var e endpoint.Endpoint
		e = makeFetchEndpointProxy(ctx, proxyURL)
		return proxyingMiddleware{ctx, next, e}
	}
}


func makeFetchEndpointProxy(ctx context.Context, instance string) endpoint.Endpoint {
	
	u, err := url.Parse(instance)
	if err != nil {
		panic(err)
	}

	u.Path = "/app/fetch"

	return httptransport.NewClient(
		"GET", u,
		encodeRequest,
		decodeFetchResponse,
	).Endpoint()
}

func decodeFetchResponse(_ context.Context, r *http.Response) (interface{}, error) {
	var response splash.Response
	logger.Println(r.Status)
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		//logger.Println(response)
		return nil, err
	}

	return response, nil
}

func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	logger.Println(buf.String())
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeFetchRequest(_ context.Context, r *http.Request, request interface{}) error {
	req := request.(splash.Request)
	var buf bytes.Buffer
	//if err := json.NewEncoder(&buf).Encode(request); err != nil {
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Body = ioutil.NopCloser(&buf)
	return nil
}
