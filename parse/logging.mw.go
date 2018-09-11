package parse

import (
	"io"
	"time"

	"github.com/slotix/dataflowkit/scrape"
	"go.uber.org/zap"
)

// LoggingMiddleware logs Parse Service endpoints
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

// Logging Parse Service
func (mw loggingMiddleware) Parse(payload scrape.Payload) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		output, err = mw.Service.Parse(payload)
		url := payload.Request.URL
		if err != nil {
			mw.logger.Info("Parse",
				zap.String("URL", url),
				zap.String("fetcher", payload.Request.Type),
				zap.Error(err),
				zap.Duration("took", time.Since(begin)))
		} else {
			mw.logger.Info("Parse",
				zap.String("URL", url),
				zap.String("fetcher", payload.Request.Type),
				zap.Duration("took", time.Since(begin)))
		}
	}(time.Now())
	return
}
