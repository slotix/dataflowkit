package fetch

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const addr = "localhost:12345"

var (
	indexContent = []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`)

	robotstxtData = "\n\t\tUser-agent: *\n\t\tAllow: /allowed\n\t\tDisallow: /disallowed\n\t\t"
	// robotstxtData = []byte(`
	// 	User-agent: *
	// 	Allow: /allowed
	// 	Disallow: /disallowed
	// 	`)
)

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
		w.Write([]byte(robotstxtData))
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
	fetcher := NewFetcher(Base)
	req := BaseFetcherRequest{
		URL:    "http://" + addr,
		Method: "GET",
	}
	resp, err := fetcher.Response(req)
	assert.Nil(t, err, "Expected no error")
	html, err := resp.GetHTML()
	assert.NoError(t, err, "Expected no error")
	data, err := ioutil.ReadAll(html)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, indexContent, data)
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
	//Test 200 response
	req = BaseFetcherRequest{
		URL: "http://" + addr,
	}

	content, err := fetcher.Fetch(req)
	assert.NoError(t, err)
	assert.NotNil(t, content, "Expected content not nil")

	//Test Host()
	req = BaseFetcherRequest{
		URL: "http://httpbin.org/status/200",
	}
	host, err := req.Host()
	assert.NoError(t, err)
	assert.Equal(t, "httpbin.org", host, "Test BaseFetcherRequest Host()")
	req = BaseFetcherRequest{
		URL: "Invalid.%$^host",
	}
	host, err = req.Host()
	assert.Error(t, err)

	//Test Type()
	assert.Equal(t, "base", req.Type(), "Test BaseFetcherRequest Type()")
	//fetch robots.txt data
	resp, err = fetcher.Response(BaseFetcherRequest{
		URL:    "http://" + addr + "/robots.txt",
		Method: "GET",
	})
	bfResponse := resp.(*BaseFetcherResponse)
	//t.Log(string(bfResponse.HTML))
	assert.Equal(t, robotstxtData, bfResponse.HTML)

}

func TestSplashFetcher_Fetch(t *testing.T) {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)

	fetcher := NewFetcher(Splash)
	//assert.Nil(t, err, "Expected no error")

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
		"http://httpbin.org/status/502",
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
	//Test Host()
	req = splash.Request{
		URL: "http://httpbin.org/status/200",
	}
	host, err := req.Host()
	assert.NoError(t, err)
	assert.Equal(t, "httpbin.org", host)
	req = splash.Request{
		URL: "Invalid.%$^host",
	}
	host, err = req.Host()
	assert.Error(t, err)

	//Test Type()
	assert.Equal(t, "splash", req.Type(), "Test splash.Request Type()")

}

func Test_parseFormData(t *testing.T) {
	formData := "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=usr&ips_password=passw&rememberMe=0"
	values := parseFormData(formData)
	assert.Equal(t,
		url.Values{"auth_key": []string{"880ea6a14ea49e853634fbdc5015a024"},
			"referer": []string{"http%3A%2F%2Fexample.com%2F"}, "ips_username": []string{"usr"},
			"ips_password": []string{"passw"},
			"rememberMe":   []string{"0"}},
		values)
}
