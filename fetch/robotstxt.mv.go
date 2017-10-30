package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/splash"
	"github.com/temoto/robotstxt"
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
	url := mw.getURL(req)
	//to avoid recursion while retrieving robots.txt
	if !splash.IsRobotsTxt(url) {
		parsedURL, err := neturl.Parse(url)
		if err != nil {
			return nil, err
		}
		//generate robotsURL from req.URL
		robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
		r := splash.Request{URL: robotsURL}

		content, err := contentFromFetchService(r)
		logger.Println(err)
		if err != nil {
			return nil, err
		}
		robotsData, err := robotstxt.FromBytes(content)
		if err != nil {
			fmt.Println("Robots.txt error:", err)
		}
		if !Allowed(url, robotsData) {
			return nil, &errs.ForbiddenByRobots{url}
		}
	}
	output, err = mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	return output, err
}

//contentFromFetchService sends request to fetch service and returns robots.txt content
func contentFromFetchService(req splash.Request) ([]byte, error) {
	//fetch content
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)
	request, err := http.NewRequest("POST", "http://127.0.0.1:8000/fetch", reader)
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(request)
	logger.Println(err)
	if r != nil {
		defer r.Body.Close()
	}
	logger.Println(err)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r.Body)
	logger.Println(err)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//Allowed checks if scraping of specified URL is allowed by robots.txt
func Allowed(url string, robotsData *robotstxt.RobotsData) bool {
	if robotsData == nil {
		return true
	}
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		logger.Println("err")
	}
	return robotsData.TestAgent(parsedURL.Path, "DataflowKitBot")
}

//GetCrawlDelay retrieves Crawl-delay directive from robots.txt. Crawl-delay is not in the standard robots.txt protocol, and according to Wikipedia, some bots have different interpretations for this value. That's why maybe many websites don't even bother defining the rate limits in robots.txt. Crawl-delay value does not have an effect on delays between consecutive requests to the same domain for the moment. FetchDelay and RandomizeFetchDelay from ScrapeOptions are used for throttling a crawler speed.
func GetCrawlDelay(r *robotstxt.RobotsData) time.Duration {
	if r != nil {
		group := r.FindGroup("DataflowKitBot")
		return group.CrawlDelay
	}
	return 0
}
