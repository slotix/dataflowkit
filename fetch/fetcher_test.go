package fetch

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/utf8encoding"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestBaseFetcher_Proxy(t *testing.T) {
	viper.Set("PROXY", "http://127.0.0.1:3128")
	viper.Set("CHROME_TRACE", true)
	//viper.Set("PROXY", "")
	fetcher := newFetcher(Base)
	assert.NotNil(t, fetcher)
	fetcher = newFetcher(Chrome)
	assert.NotNil(t, fetcher)
}

func TestBaseFetcher_Fetch(t *testing.T) {
	viper.Set("PROXY", "")
	fetcher := newFetcher(Base)
	req := Request{
		URL:    tsURL + "/hello",
		Method: "GET",
	}
	html, err := fetcher.Fetch(req)
	assert.NoError(t, err, "Expected no error")
	data, err := ioutil.ReadAll(html)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, helloContent, data)

	//Test 200 response
	req = Request{
		URL: tsURL,
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
		URL: "http://google.com/somepage",
	}
	host, err := req.Host()
	assert.NoError(t, err)
	assert.Equal(t, "google.com", host, "Test BaseFetcherRequest Host()")

	req = Request{
		URL: "Invalid.%$^host",
	}
	_, err = req.Host()
	assert.Error(t, err)

	//fetch robots.txt data
	robots, _ := fetcher.Fetch(Request{
		URL:    tsURL + "/robots.txt",
		Method: "GET",
	})
	data, err = ioutil.ReadAll(robots)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, robotsContent, string(data))

}

func TestChromeFetcher_Fetch(t *testing.T) {
	viper.Set("PROXY", "")
	fetcher := newFetcher(Chrome)
	req := Request{
		Type: "chrome",
		URL:  "http://testserver:12345",
	}
	resp, err := fetcher.Fetch(req)
	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, resp, "Expected resp not nil")

	//Test Form Data
	//TODO: Add real tests here
	req = Request{
		Type:     "chrome",
		URL:      "http://testserver:12345",
		FormData: "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=userpassword&rememberMe=1",
	}

	resp, err = fetcher.Fetch(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp, "Expected content not nil")

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
	_, err = req.Host()
	assert.Error(t, err)

	//test runJSFromFile
	req = Request{
		Type: "chrome",
		URL:  "http://testserver:12345/status/200",
		//InfiniteScroll: true,
	}
	resp, err = fetcher.Fetch(req)
	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, resp, "Expected resp not nil")
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

func TestInvalidFetcher(t *testing.T) {
	var fType Type
	fType = "unknownFetcher"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	fetcher := newFetcher(fType)
	assert.NotNil(t, fetcher)
}

func TestAuthFetcher(t *testing.T) {
	viper.Set("PROXY", "")
	fetcher := newFetcher(Chrome)

	randSrc := rand.NewSource(time.Now().UnixNano())
	nRand := rand.New(randSrc)
	randValue := strconv.Itoa(nRand.Intn(1000))
	username := "AnyUserNameAcceptable_" + randValue
	req := Request{
		Type:     "chrome",
		URL:      "http://testserver:12345/login",
		FormData: "username=" + username + "&password=123",
	}

	content, err := fetcher.Fetch(req)
	assert.NoError(t, err)

	pageContent, err := ioutil.ReadAll(content)
	assert.NoError(t, err)

	assert.Equal(t, true, bytes.Contains(pageContent, []byte(">"+username+"<")))

}
func TestBaseFetcher_Encoding(t *testing.T) {
	viper.Set("PROXY", "")
	fetcher := newFetcher(Base)
	req := Request{
		URL: tsURL + "/static/html/utf8.html",
		//URL: "https://www.tvojlekar.sk/lekari.php",
		Method: "GET",
	}
	html, err := fetcher.Fetch(req)
	assert.NoError(t, err, "Expected no error")
	_, name, _, err := utf8encoding.ReaderToUtf8Encoding(html)
	assert.NoError(t, err, "Expected no error")
	// data, err := ioutil.ReadAll(r)
	// t.Log(string(data))
	assert.Equal(t,"utf-8", name, "Expected UTF-8 page")
	
	req = Request{
		URL: tsURL + "/static/html/win1250.html",
		Method: "GET",
	}
	html, err = fetcher.Fetch(req)
	assert.NoError(t, err, "Expected no error")
	_, name, _, err = utf8encoding.ReaderToUtf8Encoding(html)
	assert.NoError(t, err, "Expected no error")
	// data, err := ioutil.ReadAll(r)
	// t.Log(string(data))
	assert.Equal(t,"windows-1250", name, "Expected Win1250 page")
}
