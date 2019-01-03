package parse

import (
	"io"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/slotix/dataflowkit/scrape"
)

// Implement service functions and add label method for our metrics
func (mw metricsMiddleware) Parse(payload scrape.Payload) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Parse"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		output, err = mw.Service.Parse(payload)
	}(time.Now())
	return
}

// metrics function
func Metrics(requestCount metrics.Counter,
	requestLatency metrics.Histogram) ServiceMiddleware {
	return func(next Service) Service {
		return metricsMiddleware{
			next,
			requestCount,
			requestLatency,
		}
	}
}

// Make a new type and wrap into Service interface
// Add expected metrics property to this type
type metricsMiddleware struct {
	Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}
