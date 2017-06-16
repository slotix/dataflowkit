package server

import (
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/slotix/dataflowkit/splash"
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

func (mw logmw) ParseData(payload []byte) (output io.ReadCloser, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "parse",
			//"input", payload,
			//"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.ParseService.ParseData(payload)
	return
}

func (mw logmw) Fetch(req splash.Request) (output interface{}, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "fetch",
			"input", req.URL,
			//	"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	output, err = mw.ParseService.Fetch(req)
	return
}

/*
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
*/
