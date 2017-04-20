package downloader

import (
	"fmt"
	"time"

	neturl "net/url"

	"github.com/slotix/dataflowkit/helpers"
	"github.com/temoto/robotstxt"
)

var ignoredURLs = []string{
	"127.0.0.1",
	"0.0.0.0",
	"dataflowkit.org",
}

func NewRobotsTxt(url string) (*string, error) {
	var robotsURL string
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	robotsURL = fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	if !helpers.StringInSlice(url, ignoredURLs) {
		return nil, fmt.Errorf("%s: Skipping... ", robotsURL)
	}

	return &robotsURL, nil
}

func GetRobotsData(content []byte) *robotstxt.RobotsData {
	r, err := robotstxt.FromBytes(content)
	if err != nil {
		logger.Println("Robots.txt error:", err)
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
