package fetch

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIsRobotsTxt(t *testing.T) {
	assert.Equal(t, false, IsRobotsTxt("http://google.com/robots.txst"))
	assert.Equal(t, true, IsRobotsTxt("http://google.com/robots.txt"))

}

func TestRobotstxtData(t *testing.T) {
	viper.Set("DFK_FETCH", "localhost:8000")
	robots, err := RobotstxtData("http://" + addr)
	assert.NoError(t, err, "No error returned")
	assert.Equal(t, true, AllowedByRobots("http://"+addr+"/allowed", robots), "Test allowed url")
	assert.Equal(t, false, AllowedByRobots("http://"+addr+"/disallowed", robots), "Test disallowed url")

}
