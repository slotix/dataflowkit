package parse

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var payload = scrape.Payload{
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
	viper.Set("FETCHER_TYPE", "base")
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host:         fetchServerAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	//Stop server
	defer fetchServer.Stop()

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
	defer parseServer.Stop()

	//create HTTPClient to send requests.
	svc, err := NewHTTPClient(parseServerAddr)
	if err != nil {
		logger.Error(err)
	}
	result, err := svc.Parse(payload)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	//Invalid request
	invPayload := scrape.Payload{
		Name: "invalid payload",
	}
	result, err = svc.Parse(invPayload)
	assert.Error(t, err)
	invPayload = scrape.Payload{
		Name: "invalid payload",
		Request: fetch.BaseFetcherRequest{
			URL: "http://books.toscrape.com",
		},
	}

	result, err = svc.Parse(invPayload)
	assert.Error(t, err)

}

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	//healthCheckHandler(w, req)
	handler := http.HandlerFunc(HealthCheckHandler)
	handler.ServeHTTP(w, req)

	// Check the status code is what we expect.
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(`{"alive": true}`), body)
}
