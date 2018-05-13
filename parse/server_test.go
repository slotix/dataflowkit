package parse

import (
	"testing"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var pJSON = scrape.Payload{
	Name: "books-to-scrape",
	// Request: splash.Request{
	// 	URL: "http://books.toscrape.com",
	// },
	FetcherType: "base",
	Request: fetch.BaseFetcherRequest{
		URL: "http://books.toscrape.com",
	},
	Fields: []scrape.Field{
		scrape.Field{
			Name:     "Title",
			Selector: "h3 a",
			Extractor: scrape.Extractor{
				Types:   []string{"text", "href"},
				Filters: []string{"trim"},
			},
		},
		scrape.Field{
			Name:     "Price",
			Selector: ".price_color",
			Extractor: scrape.Extractor{
				Types: []string{"regex"},
				Params: map[string]interface{}{
					"regexp": "([\\d\\.]+)",
				},
				Filters: []string{"trim"},
			},
		},
		scrape.Field{
			Name:     "Image",
			Selector: ".thumbnail",
			Extractor: scrape.Extractor{
				Types:   []string{"src", "alt"},
				Filters: []string{"trim"},
			},
		},
	},
	Format: "json",
}

func Test_server(t *testing.T) {
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("SKIP_STORAGE_MW", false)
	viper.Set("STORAGE_TYPE", "Diskv")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	fetchServer := fetch.Start(fetchServerCfg)

	////////
	//start parse server
	parseServerAddr := "127.0.0.1:8001"
	serverCfg := Config{
		Host:         parseServerAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	//viper.Set("SKIP_STORAGE_MW", true)
	parseServer := Start(serverCfg)

	//create HTTPClient to send requests.
	svc, err := NewHTTPClient(parseServerAddr)
	if err != nil {
		logger.Error(err)
	}
	result, err := svc.Parse(pJSON)
	if err != nil {
		logger.Error(err)
	}
	assert.NotNil(t, result)
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(result)
	//t.Log(buf.String())

	//Stop servers
	fetchServer.Stop()
	parseServer.Stop()

}
