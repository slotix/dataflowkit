package server

import (
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"time"

	"github.com/pkg/errors"
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

func (mw robotstxtmw) Fetch(req splash.Request) (output interface{}, err error) {
	allow := true
	robotsURL, err := NewRobotsTxt(req.URL)
	var robotsData *robotstxt.RobotsData
	if err != nil {
		logger.Println(err)
	} else {
		r := splash.Request{URL: robotsURL}
		robots, err := mw.ParseService.Fetch(r)
		if err != nil {
			logger.Println(err)
			//errors.Wrap(err, "robots.txt")
		} else {
			sResponse := robots.(*splash.Response)
			//data, err := ioutil.ReadAll(robots)
			content, err := sResponse.GetContent()
			//logger.Println(content)
			if err != nil {
				return nil, err
			}
			data, err := ioutil.ReadAll(content)
			if err != nil {
				return nil, err
			}
			robotsData = GetRobotsData(data)
			parsedURL, err := neturl.Parse(req.URL)
			if err != nil {
				logger.Println("err")
			}
			if robotsData != nil {
				allow = robotsData.TestAgent(parsedURL.Path, "DataflowKitBot")
			}
		}
	}

	//allowed ?
	if allow {
		//req.CrawlDelay = GetCrawlDelay(robotsData)
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

func NewRobotsTxt(url string) (string, error) {
	if url == "" {
		return "", errors.New("empty URL")
	}
	var robotsURL string
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return "", err
	}
	robotsURL = fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	return robotsURL, nil
}

func GetRobotsData(content []byte) *robotstxt.RobotsData {
	r, err := robotstxt.FromBytes(content)
	if err != nil {
		fmt.Println("Robots.txt error:", err)
	}
	return r
}

//GetCrawlDelay retrieves Crawl-delay directive from robots.txt. Crawl-delay is not in the standard robots.txt protocol, and according to Wikipedia, some bots have different interpretations for this value. That's why maybe many websites don't even bother defining the rate limits in robots.txt. Crawl-delay value does not have an effect on delays between consecutive requests to the same domain for the moment. FetchDelay and RandomizeFetchDelay from ScrapeOptions are used for throttling a crawler speed.
func GetCrawlDelay(r *robotstxt.RobotsData) time.Duration {
	if r != nil {
		group := r.FindGroup("DataflowKitBot")
		return group.CrawlDelay
	}
	return 0
}
