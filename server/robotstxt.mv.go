package server

import (
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

func (mw robotstxtMiddleware) Fetch(req interface{}) (output interface{}, err error) {
	//robotsData, err := robotstxt.RobotsTxtData(req)
	url := mw.getURL(req) 
	robotsData, err := robotstxt.RobotsTxtData(url)
	if err != nil {
		return nil, err
	}
	logger.Println(robotsData)
	if !robotstxt.Allowed(url, robotsData) {
		return nil, &splash.ErrorForbiddenByRobots{url}
	}
	output, err = mw.Service.Fetch(req)
	if err != nil {
		logger.Println(err)
	}
	return
}
