package fetch

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
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
func (mw loggingMiddleware) Fetch(req Request) (out io.ReadCloser, err error) {
	defer func(begin time.Time) {
		url := req.GetURL()
		if err == nil {
			mw.logger.WithFields(
				logrus.Fields{
					"fetcher": req.Type,
					"func":    "Fetch",
					"took":    time.Since(begin),
				}).Info("Fetch URL: ", url)
		}
		//don't log errors here. They all will be reported at transport.go func encodeError()
	}(time.Now())
	out, err = mw.Service.Fetch(req)
	return
}
