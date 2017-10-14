package parse

import (
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/slotix/dataflowkit/scrape"
)

// Make a new type and wrap into Service interface
// Add logger property to this type
type loggingMiddleware struct {
	Service
	logger log.Logger
}

// implement function to return ServiceMiddleware
func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

// Implement Service Interface for LoggingMiddleware
func (mw loggingMiddleware) ParseData(payload scrape.Payload) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "parse",
			//"input", payload,
			//"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.Service.ParseData(payload)
	return
}
