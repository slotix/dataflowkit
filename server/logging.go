package server

import (
	"time"

	"github.com/go-kit/kit/log"
)

func loggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next ParseService) ParseService {
		return logmw{logger, next}
	}
}

type logmw struct {
	logger log.Logger
	ParseService
}

func (mw logmw) ParseData(payload []byte) (output []byte, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "marshaldata",
			//"input", payload,
			//"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.ParseService.ParseData(payload)
	return
}

func (mw logmw) Download(url string) (output []byte, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "gethtml",
			"input", url,
			//	"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.ParseService.Download(url)
	return
}

func (mw logmw) CheckServices() (output map[string]string) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "chkservices",
			"took", time.Since(begin),
		)
	}(time.Now())

	output = mw.ParseService.CheckServices()
	return
}
