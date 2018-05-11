package fetch

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	//"github.com/slotix/dataflowkit/testserver"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/temoto/robotstxt"
)

func TestIsRobotsTxt(t *testing.T) {
	assert.Equal(t, false, IsRobotsTxt("http://google.com/robots.txst"))
	assert.Equal(t, true, IsRobotsTxt("http://google.com/robots.txt"))

}

func TestRobotstxtData(t *testing.T) {
	addr := "localhost:12345"
	//test AllowedByRobots func
	robots, err := robotstxt.FromString(RobotsContent)
	assert.NoError(t, err, "No error returned")
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	assert.Equal(t, false, AllowedByRobots("http://"+addr+"/disallowed", robots), "Test disallowed url")
	assert.Equal(t, time.Duration(0), GetCrawlDelay(robots))
	robots = nil
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")

	//start serving content
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write(IndexContent)
	})
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write([]byte(RobotsContent))
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	//rd, err := RobotstxtData("https://google.com")
	//viper.Set("DFK_FETCH", ts.URL)
	
	//start fetch server
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	fetchServer := viper.GetString("DFK_FETCH")
	serverCfg := Config{
		Host:         fetchServer,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	viper.Set("SKIP_STORAGE_MW", true)
	htmlServer := Start(serverCfg)

	////////
	rd, err := RobotstxtData(ts.URL)

	assert.NoError(t, err, "No error returned")
	assert.NotNil(t, rd, "Not nil returned")

	rd, err = RobotstxtData("invalid_host")
	assert.Error(t, err, "error returned")

	htmlServer.Stop()
}
