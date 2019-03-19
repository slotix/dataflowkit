package parse

import (
	"testing"
	"time"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/scrape"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

//Chrome fetcher and test server run in their own docker containers. So to make URLs of testserver visible they shoould look like http://testserver:12345/...
//Base fetcher on the contrary сan see http://localhost:12345 like URLs
var (
	payloadChrome = scrape.Payload{
		Name: "test",
		Request: fetch.Request{
			Type:      "chrome",
			URL:       "http://testserver:12345/persons/page-0",
			UserToken: "12345",
		},
		Fields: []scrape.Field{
			{
				Name:        "Name",
				CSSSelector: "#cards a",
				Attrs:       []string{"text", "href"},
				Filters:     []scrape.Filter{scrape.Filter{"trim", ""}},
			},
			{
				Name:        "Image",
				CSSSelector: ".card-img-top",
				Attrs:       []string{"src", "alt"},
			},
		},
		Format: "json",
	}

	payloadBase = scrape.Payload{
		Name: "test",
		Request: fetch.Request{
			Type:      "base",
			URL:       "http://127.0.0.1:12345",
			UserToken: "12345",
		},
		Fields: []scrape.Field{
			{
				Name:        "alert",
				CSSSelector: ".alert-info",
				Attrs:       []string{"text"},
				Filters:     []scrape.Filter{scrape.Filter{"trim", ""}},
			},
		},
		Format: "json",
	}
)

func init() {
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("DFK_PARSE", "127.0.0.1:8001")
	viper.Set("STORAGE_TYPE", "MongoDB")
	viper.Set("FETCHER_TYPE", "base")
	viper.Set("RESULTS_DIR", "results")
	viper.Set("CHROME", "http://127.0.0.1:9222")
	viper.Set("PAYLOAD_POOL_SIZE", 50)
	viper.Set("PAYLOAD_WORKERS_NUM", 100)
}
func Test_service(t *testing.T) {
	//start fetch server
	fetchServerAddr := viper.GetString("DFK_FETCH")
	fetchServerCfg := fetch.Config{
		Host: fetchServerAddr,
	}
	fetchServer := fetch.Start(fetchServerCfg)
	//Stop server
	defer fetchServer.Stop()
	time.Sleep(500 * time.Millisecond)
	svc := ParseService{}
	result, err := svc.Parse(payloadBase)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	//start parse server
	parseServerAddr := "127.0.0.1:8001"
	serverCfg := Config{
		Host: parseServerAddr,
	}
	parseServer := Start(serverCfg)
	defer parseServer.Stop()
	time.Sleep(500 * time.Millisecond)

	//create HTTPClient to send requests.
	svc1, _ := NewHTTPClient(parseServerAddr)
	result, err = svc1.Parse(payloadChrome)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	//Invalid request
	invPayload := scrape.Payload{
		Name: "invalid payload",
	}
	_, err = svc1.Parse(invPayload)
	assert.Error(t, err)

	//Invalid Payload - no fields
	invPayload = scrape.Payload{
		Name: "invalid payload",
		Request: fetch.Request{
			URL: "http://127.0.0.1:12345",
		},
	}

	_, err = svc.Parse(invPayload)
	assert.Error(t, err)

}
