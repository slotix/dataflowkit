package fetch

import (
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestBaseFetcher_Proxy(t *testing.T) {
	viper.Set("PROXY", "http://127.0.0.1:3128")
	//viper.Set("PROXY", "")
	fetcher := NewFetcher(Base)
	assert.NotNil(t, fetcher)
}

func TestBaseFetcher_Fetch(t *testing.T) {
	viper.Set("PROXY", "")
	fetcher := NewFetcher(Base)
	req := Request{
		Type: "base",
		//URL:    ts.URL,
		URL:    tsURL + "/hello",
		Method: "GET",
	}
	//resp, err := fetcher.Response(req)
	//assert.Nil(t, err, "Expected no error")
	//html, err := resp.GetHTML()
	html, err := fetcher.Fetch(req)
	assert.NoError(t, err, "Expected no error")
	data, err := ioutil.ReadAll(html)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, helloContent, data)
	//assert.Equal(t, req.GetURL(), resp.GetURL())
	//assert.Equal(t, time.Now().UTC().Add(24*time.Hour).Truncate(1*time.Minute), resp.GetExpires().Truncate(1*time.Minute), "Expires default value is 24 hours")
	
	//Test 200 response
	req = Request{
		Type: "base",
		URL:  tsURL,
	}
	content, err := fetcher.Fetch(req)
	assert.NoError(t, err)
	assert.NotNil(t, content, "Expected content not nil")

	//Test Form Data
	req = Request{
		Type:     "base",
		URL:      tsURL,
		FormData: "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1",
	}

	content, err = fetcher.Fetch(req)
	assert.NoError(t, err)
	assert.NotNil(t, content, "Expected content not nil")

	//Test Host()
	req = Request{
		Type: "base",
		URL:  "http://google.com/somepage",
	}
	host, err := req.Host()
	assert.NoError(t, err)
	assert.Equal(t, "google.com", host, "Test BaseFetcherRequest Host()")
	req = Request{
		Type: "base",
		URL:  "Invalid.%$^host",
	}
	host, err = req.Host()
	assert.Error(t, err)

	//fetch robots.txt data
	robots, err := fetcher.Fetch(Request{
		Type:   "base",
		URL:    tsURL + "/robots.txt",
		Method: "GET",
	})
	data, err = ioutil.ReadAll(robots)
	assert.NoError(t, err, "Expected no error")
	// resp, err := fetcher.Response(BaseFetcherRequest{
	// 	URL:    tsURL + "/robots.txt",
	// 	Method: "GET",
	// })
	//bfResponse := resp.(*BaseFetcherResponse)
	assert.Equal(t, robotsContent, string(data))

}

func TestChromeFetcher_Fetch(t *testing.T) {
	viper.Set("PROXY", "")
	fetcher := NewFetcher(Chrome)
	req := Request{
		Type: "chrome",
		URL:  "http://testserver:12345",
	}
	resp, err := fetcher.Fetch(req)
	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, resp, "Expected resp not nil")
	req = Request{
		Type: "chrome",
		URL:  "http://testserver:12345/status/200",
		//URL: ts.URL + "/index.html",
	}
	host, err := req.Host()
	assert.NoError(t, err)
	assert.Equal(t, "testserver:12345", host)
	req = Request{
		Type: "chrome",
		URL:  "Invalid.%$^host",
	}
	host, err = req.Host()
	assert.Error(t, err)
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
