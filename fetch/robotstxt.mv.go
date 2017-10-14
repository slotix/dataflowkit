package fetch

import (
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/robotstxt"
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
	if !robotstxt.Allowed(url, robotsData) {
		return nil, &errs.ForbiddenByRobots{url}
	}
	output, err = mw.Service.Fetch(req)
	if err != nil {
		logger.Println(err)
	}
	return
}
