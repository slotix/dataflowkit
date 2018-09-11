package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	st            storage.Store
	tsURL         string
	robotsContent = "\n\t\tUser-agent: *\n\t\tAllow: /allowed\n\t\tDisallow: /disallowed\n\t\tDisallow: /redirect\n\t\t"
	helloContent  = []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`)
)

func init() {
	viper.Set("STORAGE_TYPE", "Diskv")
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("PROXY", "")
	viper.Set("CHROME", "http://127.0.0.1:9222")
	viper.Set("CHROME_SCRIPTS", "../cmd/fetch.d/chrome")
	st = storage.NewStore(viper.GetString("STORAGE_TYPE"))
	tsURL = "http://localhost:12345"
}

func TestFetchServiceMW(t *testing.T) {
	//start fetch server
	fetchServer := viper.GetString("DFK_FETCH")
	serverCfg := Config{
		Host: fetchServer,
	}
	htmlServer := Start(serverCfg)
	defer htmlServer.Stop()

	svc, _ := NewHTTPClient(fetchServer)
	svc = RobotsTxtMiddleware()(svc)
	svc = LoggingMiddleware(logger)(svc)

	cArr := []*http.Cookie{
		{
			Name:   "cookie1",
			Value:  "cValue1",
			Domain: "localhost:12345",
		},
		{
			Name:   "cookie2",
			Value:  "cValue2",
			Domain: "localhost:12345",
		},
	}
	userToken := "12345"
	cookies, _ := json.Marshal(cArr)
	rec := storage.Record{
		Key:     userToken,
		Type:    "Cookies",
		Value:   cookies,
		ExpTime: 0,
	}
	//write cookies to storage
	err := st.Write(rec)
	if err != nil {
		t.Log(err)
	}

	data, err := svc.Fetch(Request{
		Type:      "base",
		URL:       tsURL + "/hello",
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, data, "Expected response is not nil")

	//read cookies
	data, err = svc.Fetch(Request{
		Type:      "base",
		URL:       tsURL,
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, data, "Expected response is not nil")

	//Test invalid Response Status codes.
	urls := []string{
		tsURL + "/status/404",
		tsURL + "/status/400",
		tsURL + "/status/401",
		tsURL + "/status/403",
		tsURL + "/status/407",
		tsURL + "/status/500",
		tsURL + "/status/502",
		tsURL + "/status/504",
		tsURL + "/status/600",
		"http://google",
		"google.com",
	}
	for _, url := range urls {
		req := Request{
			Type: "base",
			URL:  url,
		}
		_, err := svc.Fetch(req)
		t.Log(err)
		assert.Error(t, err, fmt.Sprintf("%T", err)+"error returned")
	}

	//invalid URL
	_, err = svc.Fetch(Request{
		Type:   "base",
		URL:    "invalid_addr",
		Method: "GET",
	})
	assert.Error(t, err, "Expected error")

	//invalid Fetcher type
	_, err = svc.Fetch(Request{
		Type:   "invalid",
		URL:    "invalid_addr",
		Method: "GET",
	})
	assert.Error(t, err, "Expected error")

	//disallowed by robots
	_, err = svc.Fetch(Request{
		Type:      "base",
		URL:       tsURL + "/disallowed",
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Error(t, err, "Expected error")

	//disallowed by robots
	_, err = svc.Fetch(Request{
		Type:      "base",
		URL:       tsURL + "/redirect",
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Error(t, err, "Expected error")

	//Test Chrome Fetcher
	//svcChrome := FetchService{}
	_, err = svc.Fetch(Request{
		Type:      "chrome",
		URL:       "http://testserver:12345",
		FormData:  "",
		UserToken: userToken,
	})
	assert.Nil(t, err, "Expected no error")

	svc1 := FetchService{}
	//Pass invalid Fetcher type directly to service skipping NewHTTPClient
	_, err = svc1.Fetch(Request{
		Type:   "invalid",
		URL:    "invalid_addr",
		Method: "GET",
	})
	assert.Error(t, err, "Expected error")

	//Test decodeChromeFetcherContent
	//Chrome returns empty result for erroneous pages: <html><head></head><body></body></html>
	//And returns no error
	data, err = svc.Fetch(Request{
		Type: "chrome",
		URL:  "http://testserver:12345/status/404",
		//URL:    "http://httpbin.org/status/404",
		Method: "GET",
	})
	assert.NoError(t, err, "No error")
	buf := new(bytes.Buffer)
	buf.ReadFrom(data)
	s := buf.String()
	t.Log(s)

	//remove cookies from storage
	err = st.Delete(rec)
	if err != nil {
		t.Log(err)
	}
}
