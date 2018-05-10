package fetch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/splash"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBaseFetcher_Fetch(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write(IndexContent)
	})
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write([]byte(RobotsContent))
	})
	r.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("allowed"))
	})
	r.HandleFunc("/disallowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("disallowed"))
	})

	r.HandleFunc("/status/{status}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		st, err := strconv.Atoi(vars["status"])
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(st)
		w.Write([]byte(vars["status"]))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()
	/////
	fetcher := NewFetcher(Base)
	req := BaseFetcherRequest{
		URL:   ts.URL,
		Method: "GET",
	}
	resp, err := fetcher.Response(req)
	assert.Nil(t, err, "Expected no error")
	html, err := resp.GetHTML()
	assert.NoError(t, err, "Expected no error")
	data, err := ioutil.ReadAll(html)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, IndexContent, data)
	assert.Equal(t, req.GetURL(), resp.GetURL())
	assert.Equal(t, time.Now().UTC().Add(24*time.Hour).Truncate(1*time.Minute), resp.GetExpires().Truncate(1*time.Minute), "Expires default value is 24 hours")

	//Test invalid Response Status codes.
	urls := []string{
		ts.URL + "/status/404",
		ts.URL + "/status/400",
		ts.URL + "/status/401",
		ts.URL + "/status/403",
		ts.URL + "/status/500",
		ts.URL + "/status/502",
		ts.URL + "/status/504",
		ts.URL + "/status/600",
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
		URL: ts.URL,
	}
	content, err := fetcher.Fetch(req)
	assert.NoError(t, err)
	assert.NotNil(t, content, "Expected content not nil")

	//Test Form Data
	req = BaseFetcherRequest{
		URL:      ts.URL,
		FormData: "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1",
	}

	content, err = fetcher.Fetch(req)
	assert.NoError(t, err)
	assert.NotNil(t, content, "Expected content not nil")

	//Test Host()
	req = BaseFetcherRequest{
		URL: "http://google.com/somepage",
	}
	host, err := req.Host()
	assert.NoError(t, err)
	assert.Equal(t, "google.com", host, "Test BaseFetcherRequest Host()")
	req = BaseFetcherRequest{
		URL: "Invalid.%$^host",
	}
	host, err = req.Host()
	assert.Error(t, err)

	//Test Type()
	assert.Equal(t, "base", req.Type(), "Test BaseFetcherRequest Type()")
	//fetch robots.txt data
	resp, err = fetcher.Response(BaseFetcherRequest{
		URL:    ts.URL + "/robots.txt",
		Method: "GET",
	})
	bfResponse := resp.(*BaseFetcherResponse)
	//t.Log(string(bfResponse.HTML))
	assert.Equal(t, RobotsContent, bfResponse.HTML)

}

func TestSplashFetcher_Fetch(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write(IndexContent)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()
	//////

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
	// urls := []string{
	// 	"http://" + addr + "/status/404",
	// 	"http://" + addr + "/status/400",
	// 	"http://" + addr + "/status/401",
	// 	"http://" + addr + "/status/unknown",
	// 	"http://" + addr + "/status/403",
	// 	"http://" + addr + "/status/500",
	// 	"http://" + addr + "/status/504",
	// 	"http://" + addr + "/status/600",
	// 	"http://google",
	// 	"google.com",
	// }
	// for _, url := range urls {
	// 	req := splash.Request{
	// 		URL: url,
	// 	}
	// 	_, err := fetcher.Fetch(req)
	// 	assert.Error(t, err, "error returned")
	// }
	//Test Host()
	req = splash.Request{
		URL: "http://httpbin.org/status/200",
		//URL: ts.URL + "/index.html",
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
