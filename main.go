package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/slotix/dataflowkit/extract"
	"github.com/slotix/dataflowkit/paginate"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

func main() {
	viper.Set("splash", "127.0.0.1:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")
	//fetcher, err := scrape.NewHttpClientFetcher()
	fetcher, err := scrape.NewSplashFetcher()
	if err != nil {
		fmt.Println(err)
	}
	config := &scrape.ScrapeConfig{
		Fetcher:    fetcher,
		DividePage: scrape.DividePageBySelector(".p"),

		Pieces: []scrape.Piece{
			{Name: "Price", Selector: ".pricen", Extractor: extract.Text{}},
			{Name: "Title", Selector: ".product-container a", Extractor: extract.Text{}},
			{Name: "BuyInfo", Selector: ".buy-info", Extractor: extract.Attr{Attr: "href"}},
		},
		Paginator: paginate.BySelector(".next", "href"),
	}
	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}
	req := splash.FetchRequest{URL: "https://drony.heureka.sk"}
	//req := scrape.HttpClientFetcherRequest{Method: "GET", URL: "https://drony.heureka.sk"}
	results, err := scraper.ScrapeWithOpts(
		req,
		scrape.ScrapeOptions{MaxPages: 2},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}

/*
func main1() {
	config := &scrape.ScrapeConfig{
		DividePage: scrape.DividePageBySelector(".ipb_table tr"),

		Pieces: []scrape.Piece{
			{Name: "Title_text", Selector: ".h4 a", Extractor: extract.Text{}},
			//{Name: "Title_href", Selector: ".h4 a", Extractor: extract.Attr{Attr: "href"}},
			{Name: "SubTitle", Selector: ".subforums a", Extractor: extract.MultipleText{}},
			//	{Name: "Themes", Selector: "li:nth-child(1) strong", Extractor: extract.Text{}},
		},
		//Paginator: paginate.BySelector(".next", "href"),

	}
	scraper, err := scrape.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scraper: %s\n", err)
		os.Exit(1)
	}

	results, err := scraper.Scrape("http://diesel.elcat.kg")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping: %s\n", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(results)
}
*/
