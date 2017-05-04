package server

import (
	"fmt"
	"io"
	"io/ioutil"
	neturl "net/url"
	"time"

	"github.com/slotix/dataflowkit/splash"
	"github.com/temoto/robotstxt"
)

func robotsTxtMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return robotstxtmw{next}
	}
}

type robotstxtmw struct {
	ParseService
}

func (mw robotstxtmw) Fetch(req splash.Request) (output io.ReadCloser, err error) {
	allow := true
	robotsURL, err := NewRobotsTxt(req.URL)
	if err != nil {
		//return nil, err
		logger.Println(err)
	} else {
		r := splash.Request{URL: *robotsURL}
		//robots, err := mw.ParseService.Download(*robotsURL)
		robots, err := mw.ParseService.Fetch(r)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(robots)
		if err != nil {
			return nil, err
		}
		robotsData := GetRobotsData(data)
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

func NewRobotsTxt(url string) (*string, error) {
	var robotsURL string
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	robotsURL = fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	return &robotsURL, nil
}

func GetRobotsData(content []byte) *robotstxt.RobotsData {
	r, err := robotstxt.FromBytes(content)
	if err != nil {
		fmt.Println("Robots.txt error:", err)
	}
	return r
	//return Robots{r, parsedURL.Path}
}

func GetCrawlDelay(r *robotstxt.RobotsData) time.Duration {
	if r != nil {
		group := r.FindGroup("DataflowKitBot")
		return group.CrawlDelay
	}
	return 0
}
