package downloader

import (
	"fmt"
	"log"
	"net/http"

	"github.com/temoto/robotstxt"
	neturl "net/url"
	"time"
)

type Robots struct{
	Robotstxt *robotstxt.RobotsData
	Path string
}

func NewRobotsTxt(url string) Robots{
	var robotsURL string
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		fmt.Println("err")
	}
	robotsURL = fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	response, err := http.Get(robotsURL)
	if err != nil {
		log.Fatalln("HTTP error:", err)
	}

	r, err := robotstxt.FromResponse(response)
	if err != nil {
		log.Fatalln("Robots.txt error:", err)
	}
	
	return Robots{r, parsedURL.Path}
}

func (r Robots) IsAllowed() bool{
	group := r.Robotstxt.FindGroup("DataflowKitBot")
	allowed := group.Test(r.Path)
	return allowed	
}

func (r Robots) CrawlDelay() time.Duration{
	group := r.Robotstxt.FindGroup("DataflowKitBot")
	return group.CrawlDelay	
}