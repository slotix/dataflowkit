package fetch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	IndexContent = []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`)

	RobotsContent = "\n\t\tUser-agent: *\n\t\tAllow: /allowed\n\t\tDisallow: /disallowed\n\t\t"
	// robotstxtData = []byte(`
	// 	User-agent: *
	// 	Allow: /allowed
	// 	Disallow: /disallowed
	// 	`)
)

func Test_server_Base(t *testing.T) {
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
		w.WriteHeader(403)
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

	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServer := viper.GetString("DFK_FETCH")
	serverCfg := Config{
		Host:         fetchServer,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	//viper.Set("SKIP_STORAGE_MW", true)
	htmlServer := Start(serverCfg)

	//create HTTPClient to send requests.
	svc, err := NewHTTPClient(fetchServer)
	if err != nil {
		logger.Error(err)
	}
	////////

	//send request to base fetcher endpoint
	req := BaseFetcherRequest{
		URL: ts.URL,
		//URL: "http://example.com",
		//URL: "http://github.com",
	}
	resp, err := svc.Response(req)
	assert.NoError(t, err, "error is nil")
	data, err := resp.GetHTML()
	assert.NoError(t, err)
	//	assert.NotNil(t, html)
	html, err := ioutil.ReadAll(data)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, IndexContent, html, "Expected Hello World")
	
	//test forbidden 
	req = BaseFetcherRequest{
		URL: ts.URL + "/disallowed",
	}
	resp, err = svc.Response(req)
	assert.Error(t, err, "returned error")

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
		_, err = svc.Response(req)
		t.Log(err)
		assert.Error(t, err)
	}

	
	htmlServer.Stop()
}

func Test_server_Splash(t *testing.T) {
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServer := viper.GetString("DFK_FETCH")
	serverCfg := Config{
		Host:         fetchServer, //"localhost:5000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	htmlServer := Start(serverCfg)

	//create HTTPClient to send requests.
	svc, err := NewHTTPClient(fetchServer)
	if err != nil {
		logger.Error(err)
	}
	////////
	//send request to splash fetcher endpoint
	sReq := splash.Request{
		//URL: "http://" + addr,
		//URL: "http://testserver:12345",
		URL: "http://example.com",
	}
	resp, err := svc.Response(sReq)
	if err != nil {
		logger.Error(err)
	}
	//data, err = ioutil.ReadAll(r)
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, resp)

	//assert.Equal(t, indexContent, data, "Expected Hello World")

	// data, err := svc.Response(sReq)
	// if err != nil {
	// 	logger.Error(err)
	// }
	// //data, err = ioutil.ReadAll(r)
	// assert.NoError(t, err, "Expected no error")
	// assert.NotNil(t, data)

	htmlServer.Stop()
}
