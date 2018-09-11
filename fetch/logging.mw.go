package fetch

import (
	"io"
	"time"

	"go.uber.org/zap"
)

// LoggingMiddleware logs Service endpoints
func LoggingMiddleware(logger *zap.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

// Make a new type and wrap into Service interface
// Add logger property to this type
type loggingMiddleware struct {
	Service
	logger *zap.Logger
}

func (mw loggingMiddleware) Fetch(req Request) (out io.ReadCloser, err error) {
	defer func(begin time.Time) {
		url := req.getURL()
		out, err = mw.Service.Fetch(req)
		if err == nil {
			mw.logger.Info("Fetch",
				zap.String("URL", url),
				zap.String("fetcher", req.Type),
				zap.Duration("took", time.Since(begin)),
			)
		} else {
			mw.logger.Error("Fetch",
				zap.String("URL", url),
				zap.String("fetcher", req.Type),
				zap.Error(err),
				zap.Duration("took", time.Since(begin)),
			)
		}
	}(time.Now())

	return
}
