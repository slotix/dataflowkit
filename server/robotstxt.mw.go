package server

import (
	"fmt"
	neturl "net/url"

	"github.com/slotix/dataflowkit/downloader"
)

func robotsTxtMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return robotstxtmw{next}
	}
}

type robotstxtmw struct {
	ParseService
}

func (mw robotstxtmw) Download(url string) (output []byte, err error) {
	robotsURL := downloader.NewRobotsTxt(url)
	robots, err := mw.ParseService.Download(robotsURL)
	robotsData := downloader.GetRobotsData(robots)
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		logger.Println("err")
	}
	allow := true
	if robotsData != nil {
		allow = robotsData.TestAgent(parsedURL.Path, "DataflowKitBot")
	}
	//allowed?
	if allow {
		output, err = mw.ParseService.Download(url)
			if err != nil {
				logger.Println(err)
			}
	} else {
		err = fmt.Errorf("%s: disallowed by robots.txt", url)
		logger.Println(err)
	}

	return
}
