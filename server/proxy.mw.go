package server

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
	"time"

	jujuratelimit "github.com/juju/ratelimit"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/slotix/dataflowkit/splash"
	"github.com/sony/gobreaker"
)

type proxyingMiddleware struct {
	ctx   context.Context
	next  Service
	fetch endpoint.Endpoint
}

func (mw proxyingMiddleware) Fetch(req splash.Request) (interface{}, error) {
	logger.Println("proxy before endpoit")
	res, err := mw.fetch(mw.ctx, req)
	logger.Println("proxy after endpoit")
	//res, err := mw.next.Fetch(req)
	if err != nil {
		return nil, err
	}
	resp := res.(splash.Response)
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}

func (mw proxyingMiddleware) ParseData(payload []byte) (io.ReadCloser, error) {
	return mw.next.ParseData(payload)
}

func ProxyingMiddleware(ctx context.Context, instances string) ServiceMiddleware {
	// If instances is empty, don't proxy.
	if instances == "" {
		logger.Println("proxy_to", "none")
		return func(next Service) Service { return next }
	}
	// Set some parameters for our client.
	var (
		qps         = 100                    // beyond which we will return an error
		maxAttempts = 3                      // per request, before giving up
		maxTime     = 250 * time.Millisecond // wallclock time, before giving up
	)
	// Otherwise, construct an endpoint for each instance in the list, and add
	// it to a fixed set of endpoints. In a real service, rather than doing this
	// by hand, you'd probably use package sd's support for your service
	// discovery system.
	var (
		instanceList = split(instances)
		subscriber   sd.FixedSubscriber
	)
	logger.Println("proxy_to", fmt.Sprint(instanceList))
	for _, instance := range instanceList {
		var e endpoint.Endpoint
		e = makeFetchEndpointProxy(ctx, instance)
		e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
		e = ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))(e)
		subscriber = append(subscriber, e)
	}
	// Now, build a single, retrying, load-balancing endpoint out of all of
	// those individual endpoints.
	balancer := lb.NewRoundRobin(subscriber)
	retry := lb.Retry(maxAttempts, maxTime, balancer)

	// And finally, return the ServiceMiddleware, implemented by proxymw.
	return func(next Service) Service {
		return proxyingMiddleware{ctx, next, retry}
	}

	//return func(next Service) Service {
	//	var e endpoint.Endpoint
	//	e = makeFetchEndpointProxy(ctx, instances)
	//	return proxyingMiddleware{ctx, next, e}
	//}
}

func makeFetchEndpointProxy(ctx context.Context, instance string) endpoint.Endpoint {
	logger.Println(instance)
	u, err := url.Parse(instance)
	if err != nil {
		panic(err)
	}

	u.Path = "/app/fetch"
	//logger.Println(u)
	return httptransport.NewClient(
		"POST", u,
		encodeRequest,
		decodeFetchResponse,
	).Endpoint()
}

func decodeFetchResponse(_ context.Context, r *http.Response) (interface{}, error) {
	var response splash.Response
	logger.Println("proxy response", r.Request.URL, r.Status)
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		logger.Println(err)
		return nil, err
	}
	logger.Println(response)
	return response, nil
}

func encodeRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	logger.Println("proxy request", buf.String())
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

func split(s string) []string {
	a := strings.Split(s, ",")
	for i := range a {
		a[i] = strings.TrimSpace(a[i])
	}
	return a
}
