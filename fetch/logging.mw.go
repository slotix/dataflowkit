package fetch

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slotix/dataflowkit/splash"
)

// LoggingMiddleware logs Service endpoints
func LoggingMiddleware(logger *logrus.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

// Make a new type and wrap into Service interface
// Add logger property to this type
type loggingMiddleware struct {
	Service
	logger *logrus.Logger
}

// Logging Service Fetches
func (mw loggingMiddleware) Fetch(req FetchRequester) (out io.ReadCloser, err error) {
	defer func(begin time.Time) {
		url := req.GetURL()
		var fetcher string
		switch req.(type) {
		case BaseFetcherRequest:
			fetcher = "base"
		case splash.Request:
			fetcher = "splash"
		default:
			panic("invalid fetcher request")
		}
		if err != nil {
			mw.logger.WithFields(
				logrus.Fields{
					"fetcher": fetcher,
					"err":     err,
					"took":    time.Since(begin),
				}).Error("Fetch URL: ", url)
		} else {
			mw.logger.WithFields(
				logrus.Fields{
					"fetcher": fetcher,
					"took":    time.Since(begin),
				}).Info("Fetch URL: ", url)
		}
	}(time.Now())
	out, err = mw.Service.Fetch(req)
	return
}

// Logging Service Responses
// func (mw loggingMiddleware) Response(req FetchRequester) (response FetchResponser, err error) {
// 	defer func(begin time.Time) {
// 		url := req.GetURL()
// 		var fetcher string
// 		switch req.(type) {
// 		case BaseFetcherRequest:
// 			fetcher = "base"
// 		case splash.Request:
// 			fetcher = "splash"
// 		default:
// 			panic("invalid fetcher request")
// 		}
// 		mw.logger.Info(
// 			"url", url,
// 			"fetcher", fetcher,
// 			//	"output", output,
// 			"function", "response",
// 			"err", err,
// 			"took", time.Since(begin),
// 		)
// 	}(time.Now())
// 	response, err = mw.Service.Response(req)
// 	return
// }
