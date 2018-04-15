package fetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/temoto/robotstxt"
)

func TestIsRobotsTxt(t *testing.T) {
	assert.Equal(t, false, IsRobotsTxt("http://google.com/robots.txst"))
	assert.Equal(t, true, IsRobotsTxt("http://google.com/robots.txt"))

}

func TestRobotstxtData(t *testing.T) {
	robots, err := robotstxt.FromString(robotstxtData)
	assert.NoError(t, err, "No error returned")
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	assert.Equal(t, false, AllowedByRobots("http://"+addr+"/disallowed", robots), "Test disallowed url")
	assert.Equal(t, time.Duration(0), GetCrawlDelay(robots))

}

/* func TestRobotstxtData(t *testing.T) {
	viper.Set("DFK_FETCH", "localhost:8000")
	robots, err := RobotstxtData("http://" + addr)
	assert.NoError(t, err, "No error returned")
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	assert.Equal(t, false, AllowedByRobots("http://"+addr+"/disallowed", robots), "Test disallowed url")

}
*/
