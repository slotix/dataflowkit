package server

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	jujuratelimit "github.com/juju/ratelimit"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
)

func proxyingMiddleware(ctx context.Context, instances string, logger log.Logger) ServiceMiddleware {
	// If instances is empty, don't proxy.
	if instances == "" {
		logger.Log("proxy_to", "none")
		return func(next ParseService) ParseService { return next }
	}

	// Set some parameters for our client.
	var (
		qps         = 100                    // beyond which we will return an error
		maxAttempts = 3                      // per request, before giving up
		maxTime     = 2500 * time.Millisecond // wallclock time, before giving up
	)

	// Otherwise, construct an endpoint for each instance in the list, and add
	// it to a fixed set of endpoints. In a real service, rather than doing this
	// by hand, you'd probably use package sd's support for your service
	// discovery system.
	var (
		instanceList = split(instances)
		subscriber   sd.FixedSubscriber
	)
	logger.Log("proxy_to", fmt.Sprint(instanceList))
	for _, instance := range instanceList {
		var e endpoint.Endpoint
		//e = makeGetHTMLProxy(ctx, instance)
		e = makeMarshalDataProxy(ctx, instance)
		e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
		e = ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))(e)
		subscriber = append(subscriber, e)
	}

	// Now, build a single, retrying, load-balancing endpoint out of all of
	// those individual endpoints.
	balancer := lb.NewRoundRobin(subscriber)
	retry := lb.Retry(maxAttempts, maxTime, balancer)

	// And finally, return the ServiceMiddleware, implemented by proxymw.
	return func(next ParseService) ParseService {
		return proxymw{ctx, next, retry}
	}
}

// proxymw implements StringService, forwarding Uppercase requests to the
// provided endpoint, and serving all other (i.e. Count) requests via the
// next StringService.
type proxymw struct {
	ctx   context.Context
	next  ParseService      // Serve most requests via this service...
	parse endpoint.Endpoint // ...except Uppercase, which gets served by this endpoint
}

func (mw proxymw) GetHTML(url string) ([]byte, error) {
	return mw.next.GetHTML(url)
}

func (mw proxymw) MarshalData(payload []byte) ([]byte, error) {
	response, err := mw.parse(mw.ctx, payload)
	if err != nil {
		return nil, err
	}
	return response.([]byte), nil
}

func makeMarshalDataProxy(ctx context.Context, instance string) endpoint.Endpoint {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		panic(err)
	}
	if u.Path == "" {
		u.Path = "/app/marshaldata"
	}
    fmt.Println("u", u)
    
	return httptransport.NewClient(
		"POST",
		u,
		encodeRequest,
		decodeMarshalDataResponse,
	).Endpoint()
}

func (mw proxymw) CheckServices() map[string]string {
	return mw.next.CheckServices()
}


func split(s string) []string {
	a := strings.Split(s, ",")
	for i := range a {
		a[i] = strings.TrimSpace(a[i])
	}
	return a
}
