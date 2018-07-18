package fetch

import (
	"io/ioutil"
	"testing"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)



func Test_server_Base(t *testing.T) {
	// r := mux.NewRouter()
	// r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Conent-Type", "text/html")
	// 	w.Write(IndexContent)
	// })
	// r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Conent-Type", "text/html")
	// 	w.Write([]byte(RobotsContent))
	// })
	// r.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(200)
	// 	w.Write([]byte("allowed"))
	// })
	// r.HandleFunc("/disallowed", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(403)
	// 	w.Write([]byte("disallowed"))
	// })

	// r.HandleFunc("/status/{status}", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	st, err := strconv.Atoi(vars["status"])
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	w.WriteHeader(st)
	// 	w.Write([]byte(vars["status"]))
	// })

	// ts := httptest.NewServer(r)
	// defer ts.Close()

	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServer := viper.GetString("DFK_FETCH")
	serverCfg := Config{
		Host: fetchServer,
	}
	//viper.Set("SKIP_STORAGE_MW", true)
	htmlServer := Start(serverCfg)
	defer htmlServer.Stop()
	//time.Sleep(5 * time.Second)
	//create HTTPClient to send requests.
	svc, err := NewHTTPClient(fetchServer)
	if err != nil {
		logger.Error(err)
	}
	////////

	//send request to base fetcher endpoint
	req := BaseFetcherRequest{
		URL: tsURL + "/hello",
		//URL: "http://example.com",
		//URL: "http://github.com",
	}
	// resp, err := svc.Response(req)
	// assert.NoError(t, err, "error is nil")
	// data, err := resp.GetHTML()
	// assert.NoError(t, err)
	//	assert.NotNil(t, html)
	data, err := svc.Fetch(req)
	assert.NoError(t, err, "error is nil")
	html, err := ioutil.ReadAll(data)
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`), html, "Expected Hello World")

	//test forbidden
	req = BaseFetcherRequest{
		URL: tsURL + "/disallowed",
	}
	data, err = svc.Fetch(req)
	assert.Error(t, err, "returned error")

	//Test invalid Response Status codes.
	urls := []string{
		tsURL + "/status/404",
		tsURL + "/status/400",
		tsURL + "/status/401",
		tsURL + "/status/403",
		tsURL + "/status/500",
		tsURL + "/status/502",
		tsURL + "/status/504",
		tsURL + "/status/600",
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

}

func Test_server_Splash(t *testing.T) {
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServer := viper.GetString("DFK_FETCH")
	serverCfg := Config{
		Host: fetchServer, //"localhost:5000",
	}
	viper.Set("SKIP_STORAGE_MW", true)
	htmlServer := Start(serverCfg)
	defer htmlServer.Stop()

	//time.Sleep(1 * time.Second)
	//create HTTPClient to send requests.
	svc, err := NewHTTPClient(fetchServer)
	if err != nil {
		logger.Error(err)
	}
	////////
	//send request to splash fetcher endpoint
	sReq := splash.Request{
		//URL: "http://" + addr,
		URL: "http://testserver:12345",
		//URL: "http://example.com",
	}
	resp, err := svc.Fetch(sReq)
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

}
