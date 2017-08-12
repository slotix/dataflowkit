package server

import (
	"fmt"

	"github.com/slotix/dataflowkit/robotstxt"
	"github.com/slotix/dataflowkit/splash"
)

func RobotsTxtMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return robotstxtMiddleware{next}
	}
}

type robotstxtMiddleware struct {
	Service
}

func (mw robotstxtMiddleware) Fetch(req splash.Request) (output interface{}, err error) {
	robotsData, err := robotstxt.RobotsTxtData(req)
	if err != nil {
		return nil, err
	}
	if !robotstxt.Allowed(req.URL, robotsData) {
		err = fmt.Errorf("%s: forbidden by robots.txt", req.URL)
		return nil, err
	}
	output, err = mw.Service.Fetch(req)
	if err != nil {
		logger.Println(err)
	}
	return	
}
