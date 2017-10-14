package fetch

import (
	"time"

	"github.com/go-kit/kit/log"
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
func (mw loggingMiddleware) Fetch(req interface{}) (output interface{}, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "fetch",
			"url", mw.getURL(req),
			//	"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.Service.Fetch(req)
	return
}

func (mw loggingMiddleware) Response(req interface{}) (output interface{}, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "response",
			"url", mw.getURL(req),
			//	"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.Service.Response(req)
	return
}
