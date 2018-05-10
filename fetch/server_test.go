package fetch

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_server_Base(t *testing.T) {
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
		URL: "http://" + addr,
		//URL: "http://example.com",
		//URL: "http://github.com",
	}
	data, err := svc.Fetch(req)
	if err != nil {
		logger.Error(err)
	}
	assert.NoError(t, err, "error is nil")
	//	assert.NotNil(t, html)
	html, err := ioutil.ReadAll(data)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, indexContent, html, "Expected Hello World")
	
	//Test invalid Response Status codes.
	urls := []string{
		"http://" + addr + "/status/404",
		"http://" + addr + "/status/400",
		"http://" + addr + "/status/401",
		//"http://" + addr + "/status/unknown",
		"http://" + addr + "/status/403",
		"http://" + addr + "/status/500",
		"http://" + addr + "/status/502",
		"http://" + addr + "/status/504",
		"http://" + addr + "/status/600",
		"http://google",
		"google.com",
	}
	for _, url := range urls {
		req := BaseFetcherRequest{
			URL: url,
		}
		_, err = svc.Fetch(req)
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
		URL: "http://example.com",
	}
	r, err := svc.Response(sReq)
	if err != nil {
		logger.Error(err)
	}
	//data, err = ioutil.ReadAll(r)
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, r)
	//t.Log(string(data))
	//assert.Equal(t, indexContent, data, "Expected Hello World")

	data, err := svc.Fetch(sReq)
	if err != nil {
		logger.Error(err)
	}
	//data, err = ioutil.ReadAll(r)
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, data)

	htmlServer.Stop()
}
