package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/andrew-d/goscrape"
	"github.com/andrew-d/goscrape/extract"
	"github.com/andrew-d/goscrape/paginate"
)

func main() {
	config := &scrape.ScrapeConfig{
		DividePage: scrape.DividePageBySelector("tr:nth-child(3) tr:nth-child(3n-2):not([style='height:10px'])"),

		Pieces: []scrape.Piece{
			{Name: "title", Selector: "td.title > a", Extractor: extract.Text{}},
			{Name: "link", Selector: "td.title > a", Extractor: extract.Attr{Attr: "href"}},
			{Name: "rank", Selector: "td.title[align='right']",
				Extractor: extract.Regex{Regex: regexp.MustCompile(`(\d+)`)}},
		},

		Paginator: paginate.BySelector("a[rel='nofollow']:last-child", "href"),
	}

	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}

	results, err := scraper.ScrapeWithOpts(
		"https://news.ycombinator.com",
		scrape.ScrapeOptions{MaxPages: 3},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}
