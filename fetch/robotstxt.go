package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/temoto/robotstxt"
)

//IsRobotsTxt returns true if resource is robots.txt file
func IsRobotsTxt(url string) bool {
	if strings.HasSuffix(url, "robots.txt") {
		return true
	}
	return false
}

//fetchRobots retrieves content of robots.txt with BaseFetcher.
func fetchRobots(req FetchRequester) ([]byte, error) {
	//fetch content
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)

	//https://gitlab.com/gitlab-org/gitlab-ce/issues/33534
	//Use BaseFetcher to get robots.txt content
	addr := "http://" + viper.GetString("DFK_FETCH") + "/fetch/base"
	request, err := http.NewRequest("POST", addr, reader)
	if err != nil {
		return nil, err
	}
	//request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(request)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//Response is used for getting HEAD response to check if a request is redirected.
func Response(req FetchRequester) (*BaseFetcherResponse, error) {
	//fetch content
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)

	//https://gitlab.com/gitlab-org/gitlab-ce/issues/33534
	//Use BaseFetcher to get robots.txt content
	addr := "http://" + viper.GetString("DFK_FETCH") + "/response/base"
	request, err := http.NewRequest("POST", addr, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(request)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	bfResponse := &BaseFetcherResponse{}

	if err := json.Unmarshal(data, &bfResponse); err != nil {
		return nil, err
	}
	return bfResponse, nil
}

//RobotstxtData generates robots.txt url, retrieves its content through API fetch endpoint.
func RobotstxtData(url string) (robotsData *robotstxt.RobotsData, err error) {

	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	//generate robotsURL from req.URL
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	r := BaseFetcherRequest{URL: robotsURL, Method: "GET"}

	content, err := fetchRobots(r)
	//content, err := Response(r)

	if err != nil {
		return nil, err
	}
	robotsData, err = robotstxt.FromBytes(content)
	if err != nil {
		fmt.Println("Robots.txt error:", err)
	}

	//
	// r = BaseFetcherRequest{URL: url, Method: "HEAD"}
	// d, err := head(r)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// logger.Println(string(d))
	return

}

//Allowed checks if scraping of specified URL is allowed by robots.txt
func AllowedByRobots(url string, robotsData *robotstxt.RobotsData) bool {
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
