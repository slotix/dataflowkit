package fetch

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const addr = "localhost:12345"

var (
	indexContent = []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`)

	robotstxtData = []byte(`
		User-agent: *
		Allow: /allowed
		Disallow: /disallowed
		`)
)

func robotstxtHandler(w http.ResponseWriter, r *http.Request) {

}

func init() {
	server := &http.Server{}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write(indexContent)
	})
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write(robotstxtData)
	})
	http.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("allowed"))
	})
	http.HandleFunc("/disallowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("disallowed"))
	})

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
		"http://httpbin.org/status/unknown",
		"http://httpbin.org/status/403",
		"http://httpbin.org/status/500",
		"http://httpbin.org/status/504",
		"http://httpbin.org/status/600",
		"http://google",
		"google.com",
	}
	for _, url := range urls {
		req := BaseFetcherRequest{
			URL: url,
		}
		_, err := fetcher.Fetch(req)
		t.Log(err)
		assert.Error(t, err, fmt.Sprintf("%T", err)+"error returned")
	}
	//fetch robots.txt data
	resp, err = fetcher.Fetch(BaseFetcherRequest{
		URL:    "http://" + addr + "/robots.txt",
		Method: "GET",
	})
	bfResponse = resp.(*BaseFetcherResponse)
	//t.Log(string(bfResponse.HTML))
	assert.Equal(t, robotstxtData, bfResponse.HTML)
}

func TestSplashFetcher_Fetch(t *testing.T) {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)

	fetcher, err := NewFetcher(Splash)
	assert.Nil(t, err, "Expected no error")
	err = fetcher.Prepare()
	assert.Nil(t, err, "Expected no error")

	req := splash.Request{
		URL: "http://example.com",
	}
	resp, err := fetcher.Fetch(req)
	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, resp, "Expected resp not nil")

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
		req := splash.Request{
			URL: url,
		}
		_, err := fetcher.Fetch(req)
		assert.Error(t, err, "error returned")
	}
}


func TestNewFetcher_invalid(t *testing.T) {
	_, err := NewFetcher("Invalid")
	assert.NotNil(t, err, "Expected Invalid Fetcher error")
}
