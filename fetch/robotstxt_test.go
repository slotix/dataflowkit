package fetch

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/temoto/robotstxt"
)

func TestIsRobotsTxt(t *testing.T) {
	assert.Equal(t, false, IsRobotsTxt("http://google.com/robots.txst"))
	assert.Equal(t, true, IsRobotsTxt("http://google.com/robots.txt"))

}

func TestRobotstxtData(t *testing.T) {
	//test AllowedByRobots func
	robots, err := robotstxt.FromString(robotstxtData)
	assert.NoError(t, err, "No error returned")
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	assert.Equal(t, false, AllowedByRobots("http://"+addr+"/disallowed", robots), "Test disallowed url")
	assert.Equal(t, time.Duration(0), GetCrawlDelay(robots))
	robots = nil
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")


	viper.Set("DFK_FETCH", "http://127.0.0.1:8000")
	//rd, err := RobotstxtData("http://" + addr)
	rd, err := RobotstxtData("https://google.com")
	assert.NoError(t, err, "No error returned")
	assert.NotNil(t, rd, "No error returned")

	rd, err = RobotstxtData("invalid_host")
	assert.Error(t, err, "No error returned")
}