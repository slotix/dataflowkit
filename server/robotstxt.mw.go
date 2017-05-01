package server

import (
	"fmt"
	"io"
	"io/ioutil"
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

func (mw robotstxtmw) Fetch(req downloader.FetchRequest) (output io.ReadCloser, err error) {
	allow := true
	robotsURL, err := downloader.NewRobotsTxt(req.URL)
	if err != nil {
		//return nil, err
		logger.Println(err)
	} else {
		r := downloader.FetchRequest{URL: *robotsURL}
		//robots, err := mw.ParseService.Download(*robotsURL)
		robots, err := mw.ParseService.Fetch(r)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(robots)
		if err != nil {
			return nil, err
		}
		robotsData := downloader.GetRobotsData(data)
		parsedURL, err := neturl.Parse(req.URL)
		if err != nil {
			logger.Println("err")
		}
		if robotsData != nil {
			allow = robotsData.TestAgent(parsedURL.Path, "DataflowKitBot")
		}
	}

	//allowed ?
	if allow {
		output, err = mw.ParseService.Fetch(req)
		if err != nil {
			logger.Println(err)
		}
	} else {
		output = nil
		err = fmt.Errorf("%s: forbidden by robots.txt", req.URL)
		logger.Println(err)
	}
	return
}
