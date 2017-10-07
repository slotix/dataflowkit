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

//403 Forbidden
type errorForbiddenByRobots struct {
	URL string
}

func (e *errorForbiddenByRobots) Error() string { return e.URL + ": forbidden by robots.txt" }

func (mw robotstxtMiddleware) Fetch(req splash.Request) (output interface{}, err error) {
	robotsData, err := robotstxt.RobotsTxtData(req)
	if err != nil {
		return nil, err
	}
	if !robotstxt.Allowed(req.URL, robotsData) {
		return nil, &errorForbiddenByRobots{req.URL}
	}
	output, err = mw.Service.Fetch(req)
	if err != nil {
		logger.Println(err)
	}
	return
}
