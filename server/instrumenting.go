package server

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
)

func instrumentingMiddleware(
	requestCount metrics.Counter,
	requestLatency metrics.Histogram,
	countResult metrics.Histogram,
) ServiceMiddleware {
	return func(next ParseService) ParseService {
		return instrmw{requestCount, requestLatency, countResult, next}
	}
}

type instrmw struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	countResult    metrics.Histogram
	ParseService
}

func (mw instrmw) MarshalData(payload []byte) (output string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "marshaldata", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.ParseService.MarshalData(payload)
	return
}

func (mw instrmw) GetHTML(url string) (output string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "gethtml", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.ParseService.GetHTML(url)
	return
}