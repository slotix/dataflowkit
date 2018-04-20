package parse

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slotix/dataflowkit/scrape"
)

// LoggingMiddleware logs Parse Service endpoints
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

// Logging Parse Service
func (mw loggingMiddleware) Parse(payload scrape.Payload) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		// mw.logger.Log(
		// 	"parsed URL", payload.Request.GetURL(),
		// 	"err", err,
		// 	"took", time.Since(begin),
		// )
		url := payload.Request.GetURL()
		if err != nil {
			mw.logger.WithFields(
				logrus.Fields{
					"err":     err,
					"fetcher": payload.FetcherType,
					"took":    time.Since(begin),
				}).Error("Parse URL: ", url)
		} else {
			mw.logger.WithFields(
				logrus.Fields{
					"fetcher": payload.FetcherType,
					"took":    time.Since(begin),
				}).Info("Fetch URL: ", url)
		}
	}(time.Now())
	output, err = mw.Service.Parse(payload)
	return
}
