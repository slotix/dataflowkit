package server

import (
	"fmt"

	"github.com/slotix/dfk-parser/downloader"
)

func robotsTxtMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return robotstxtmw{next}
	}
}

type robotstxtmw struct {
	ParseService
}

func (mw robotstxtmw) Download(url string) (output []byte, err error) {
	robots := downloader.NewRobotsTxt(url)
	fmt.Println(robots.IsAllowed())
	fmt.Println(robots.CrawlDelay())
	output, err = mw.ParseService.Download(url)
	if err != nil {
		fmt.Println(err)
	}
	return
}
