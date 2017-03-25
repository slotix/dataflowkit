package downloader

import (
	"fmt"
	"log"

	neturl "net/url"

	"github.com/temoto/robotstxt"
)

func NewRobotsTxt(url string) string {
	var robotsURL string
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		log.Println(err)
	}
	robotsURL = fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	return robotsURL
}

func GetRobotsData(content []byte) *robotstxt.RobotsData {
	r, err := robotstxt.FromBytes(content)
	if err != nil {
		logger.Println("Robots.txt error:", err)
	}
	return r
	//return Robots{r, parsedURL.Path}
}

/*
func CrawlDelay() time.Duration {
	if r.Robotstxt != nil {
		group := r.Robotstxt.FindGroup("DataflowKitBot")
		return group.CrawlDelay
	}
	return 0
}
*/
