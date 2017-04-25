package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrew-d/goscrape"
	"github.com/andrew-d/goscrape/extract"
	"github.com/andrew-d/goscrape/paginate"
)

func main() {
	config := &scrape.ScrapeConfig{
		DividePage: scrape.DividePageBySelector(".p"),

		Pieces: []scrape.Piece{
			{Name: "Price", Selector: ".pricen", Extractor: extract.Text{}},
			{Name: "Title", Selector: ".product-container a", Extractor: extract.Attr{Attr: "href"}},
			{Name: "Reviews", Selector: ".review-count a", Extractor: extract.Text{}},
		},
		Paginator: paginate.BySelector(".listing a", "href"),
	//	Paginator: paginate.ByQueryParam("f"),
    }

	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}

	results, err := scraper.ScrapeWithOpts(
		"https://drony.heureka.sk",
		scrape.ScrapeOptions{MaxPages: 3},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}
