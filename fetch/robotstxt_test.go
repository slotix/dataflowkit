package fetch

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/temoto/robotstxt"
)

func TestIsRobotsTxt(t *testing.T) {
	assert.Equal(t, false, isRobotsTxt("http://google.com/robots.txst"))
	assert.Equal(t, true, isRobotsTxt("http://google.com/robots.txt"))

}

func TestRobotstxtData(t *testing.T) {
	addr := "localhost:12345"
	//test AllowedByRobots func
	robots, err := robotstxt.FromString(robotsContent)
	assert.NoError(t, err, "No error returned")
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	assert.Equal(t, false, AllowedByRobots("http://"+addr+"/disallowed", robots), "Test disallowed url")
	assert.Equal(t, time.Duration(0), GetCrawlDelay(robots))
	robots = nil
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	serverCfg := Config{
		Host: viper.GetString("DFK_FETCH"),
	}
	htmlServer := Start(serverCfg)

	////////
	rd, err := RobotstxtData(tsURL)

	assert.NoError(t, err, "No error returned")
	assert.NotNil(t, rd, "Not nil returned")

	_, err = RobotstxtData("invalid_host")
	assert.Error(t, err, "error returned")

	htmlServer.Stop()
}
