package fetch

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const addr = "localhost:12345"

var indexContent = []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Conent-Type", "text/html")
	w.Write(indexContent)
}

func init() {
	server := &http.Server{}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", indexHandler)
	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Error("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
}

func TestBaseFetcher_Fetch(t *testing.T) {
	fetcher, err := NewFetcher(Base)
	assert.Nil(t, err, "Expected no error")
	err = fetcher.Prepare()
	assert.Nil(t, err, "Expected no error")
	req := BaseFetcherRequest{
		URL:    "http://" + addr,
		Method: "GET",
	}
	resp, err := fetcher.Fetch(req)
	assert.Nil(t, err, "Expected no error")
	bfResponse := resp.(*BaseFetcherResponse)
	assert.Equal(t, indexContent, bfResponse.HTML)
	assert.Equal(t, req.GetURL(), resp.GetURL())
	assert.Equal(t, time.Now().UTC().Add(24*time.Hour).Truncate(1*time.Minute), resp.GetExpires().Truncate(1*time.Minute), "Expires default value is 24 hours")

	//Test invalid Response Status codes.
	urls := []string{
		"http://httpbin.org/status/404",
		"http://httpbin.org/status/400",
		"http://httpbin.org/status/403",
		//"http://httpbin.org/status/500",
		"http://httpbin.org/status/504",
		"http://google",
		"google.com",
	}
	for _, url := range urls {
		req := BaseFetcherRequest{
			URL: url,
		}
		_, err := fetcher.Fetch(req)
		assert.Error(t, err, "error returned")
	}

}
