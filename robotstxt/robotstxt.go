package robotstxt

import (
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"time"

	"strings"

	"github.com/pkg/errors"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
	"github.com/temoto/robotstxt"
)

//RobotsTxtData returns RobotsTxtData structure or an error otherwise
//func RobotsTxtData(req interface{}) (robotsData *robotstxt.RobotsData, err error) {
func RobotsTxtData(URL string) (robotsData *robotstxt.RobotsData, err error) {
	
	//url := strings.TrimSpace(req.(scrape.HttpClientFetcherRequest).URL)
	url := strings.TrimSpace(URL)

	if url == "" {
		return nil, errors.New("empty URL")
	}
	var robotsURL string
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	//generate robotsURL from req.URL
	robotsURL = fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	
	//fetch robots.txt
	r := splash.Request{URL: robotsURL}
	fetcher, err := scrape.NewSplashFetcher()

	//r := scrape.HttpClientFetcherRequest{URL: robotsURL}
	//fetcher, err := scrape.NewHttpClientFetcher()
	if err != nil {
		return nil, err
	}
	robots, err := fetcher.Fetch(r)
	
	if err != nil {
		logger.Println(err)
		//return nil, err
	} else {
		//	robotsData, err = robotstxt.FromResponse(robots.(*http.Response))

		sResponse := robots.(*splash.Response)
		content, err := sResponse.GetContent()
		if err != nil {
			return nil, err
		}
		
		data, err := ioutil.ReadAll(content)
		if err != nil {
			return nil, err
		}
		robotsData, err = robotstxt.FromBytes(data)
		if err != nil {
			fmt.Println("Robots.txt error:", err)
		}

	}
	//logger.Println(robotsData)
	return robotsData, nil
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
