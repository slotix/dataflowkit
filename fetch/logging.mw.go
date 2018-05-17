package fetch

import (
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

// Fetch logs requests to Fetch endpoint
// func (mw loggingMiddleware) Fetch(req FetchRequester) (out io.ReadCloser, err error) {
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
// 		if err == nil {
// 			mw.logger.WithFields(
// 				logrus.Fields{
// 					"fetcher": fetcher,
// 					"func":    "Fetch",
// 					"took":    time.Since(begin),
// 				}).Info("Fetch URL: ", url)
// 		}
// 		//don't log errors here. They all will be reported at transport.go func encodeError()
// 	}(time.Now())
// 	out, err = mw.Service.Fetch(req)
// 	return
// }

// Fetch logs requests to Response endpoint
func (mw loggingMiddleware) Response(req FetchRequester) (response FetchResponser, err error) {
	defer func(begin time.Time) {
		response, err = mw.Service.Response(req)
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
		if err == nil {
			mw.logger.WithFields(
				logrus.Fields{
					"fetcher": fetcher,
					"func":    "Response",
					"took":    time.Since(begin),
				}).Info("Fetch URL: ", url)
		} else {
			mw.logger.WithFields(
				logrus.Fields{
					"err":   err,
					"fetcher": fetcher,
					"func":    "Response",
					"took":    time.Since(begin),
				}).Error("Fetch URL: ", url)
		}
	}(time.Now())

	return
}
