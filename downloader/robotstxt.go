package downloader

import (
	"fmt"
	"strings"
	"time"

	neturl "net/url"

	"github.com/temoto/robotstxt"
)

func isRobotsTxt(url string) bool {
	if strings.Contains(url, "robots.txt") {
		return true
	}
	return false
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
