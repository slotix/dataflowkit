package server

import (
	"fmt"
	"io"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/slotix/dataflowkit/splash"
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

func (mw instrmw) ParseData(payload []byte) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "marshaldata", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.ParseService.ParseData(payload)
	return
}

func (mw instrmw) Fetch(req splash.Request) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "gethtml", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.ParseService.Fetch(req)
	return
}
