package parse

import (
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/slotix/dataflowkit/scrape"
)

// LoggingMiddleware logs Parse Service endpoints
func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

// Make a new type and wrap into Service interface
// Add logger property to this type
type loggingMiddleware struct {
	Service
	logger log.Logger
}

// Logging Parse Service
func (mw loggingMiddleware) Parse(payload scrape.Payload) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			//"function", "parse",
			//"input", payload,
			//"output", output,
			"parsing", payload.Request.URL,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.Service.Parse(payload)
	return
}
