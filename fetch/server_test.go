package fetch

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_server(t *testing.T) {
	//start fetch server
	fetchServer := "127.0.0.1:8002"
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

	//send request to base fetcher endpoint
	req := BaseFetcherRequest{
		URL: "http://" + addr,
		//URL: "http://google.com",

	}
	html, err := svc.Fetch(req)
	if err != nil {
		logger.Error(err)
	}
	assert.NoError(t, err, "error is nil")
	data, err := ioutil.ReadAll(html)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, indexContent, data, "Expected Hello World")

	//send request to splash fetcher endpoint
	sReq := splash.Request{
		//URL: "http://" + addr,
		URL: "http://example.com",
	}
	r, err := svc.Fetch(sReq)
	if err != nil {
		logger.Error(err)
	}
	//data, err = ioutil.ReadAll(r)
	assert.NoError(t, err, "Expected no error")
	assert.NotNil(t, r)
	//t.Log(string(data))
	//assert.Equal(t, indexContent, data, "Expected Hello World")

	// //Test forbidden by robots
	// req = BaseFetcherRequest{
	// 	URL: "https://github.com",
	// }
	// html, err = svc.Fetch(req)
	// if err != nil {
	// 	logger.Error(err)
	// }
	// assert.NoError(t, err, "error is nil")

	//Stop fetch server
	htmlServer.Stop()
}
