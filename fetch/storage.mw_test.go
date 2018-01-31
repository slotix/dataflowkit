package fetch

import (
	"testing"
	"time"

	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	mw storageMiddleware
	s storage.Store
)

func init() {
	var svc Service
	svc = FetchService{}
	s = storage.NewStore(storage.Diskv)
	mw = storageMiddleware{
		storage: s,
		Service: svc,
	}
}

func Test_storageMiddleware(t *testing.T) {	
	req := splash.Request{
		URL:    "http://example.com",
		Params: "", Cookies: "", LUA: "",
	}
	//Loading from remote server
	start := time.Now()
	resp, err := mw.Response(req)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, resp.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")
	elapsed1 := time.Since(start)
	t.Log("Loading from remote server... ", elapsed1)

	//Loading from cached storage
	start = time.Now()
	resp, err = mw.Response(req)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, resp.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")
	elapsed2 := time.Since(start)
	t.Log("Loading from remote server... ", elapsed2)
	assert.Equal(t, true, elapsed1>elapsed2, "it takes longer to load a webpage from remote server")

	err = s.DeleteAll()
	assert.Nil(t, err, "Expected no error")
}

func Test_IGNORE_CACHE_INFO(t *testing.T) {
	viper.Set("IGNORE_CACHE_INFO", true)
	req := splash.Request{
		URL:    "http://google.com",
		Params: "", Cookies: "", LUA: "",
	}
	//Loading from remote server
	start := time.Now()
	resp, err := mw.Response(req)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, resp.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")
	elapsed1 := time.Since(start)
	t.Log("Loading from remote server... ", elapsed1)

	//Loading from cached storage
	start = time.Now()
	resp, err = mw.Response(req)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, resp.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")
	elapsed2 := time.Since(start)
	t.Log("Loading from remote server... ", elapsed2)
	assert.Equal(t, true, elapsed1>elapsed2, "it takes longer to load a webpage from remote server")

	err = s.DeleteAll()
	assert.Nil(t, err, "Expected no error")
}