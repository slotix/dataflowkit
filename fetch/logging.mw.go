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
func (mw loggingMiddleware) Fetch(req FetchRequester) (response FetchResponser, err error) {
	defer func(begin time.Time) {
		url := req.GetURL()
		mw.logger.Log(
			"function", "fetch",
			"url", url,
			//	"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	response, err = mw.Service.Fetch(req)
	return
}

func (mw loggingMiddleware) Response(req FetchRequester) (response FetchResponser, err error) {
	defer func(begin time.Time) {
		url := req.GetURL()
		mw.logger.Log(
			"function", "response",
			"url", url,
			//	"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	response, err = mw.Service.Response(req)
	return
}
