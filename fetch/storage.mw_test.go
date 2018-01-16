package fetch

import (
	"testing"

	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/storage"
	"github.com/stretchr/testify/assert"
)

func Test_storageMiddleware(t *testing.T) {
	var svc Service
	svc = FetchService{}
	storageType, err := storage.ParseType("Diskv")
	assert.Nil(t, err, "Expected no error")
	storage := storage.NewStore(storageType)
	mw := storageMiddleware{
		storage: storage,
		Service: svc,
	}
	resp, err := mw.Fetch(splash.Request{
		URL:    "http://example.com",
		Params: "", Cookies: "", Func: "",
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, resp.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")

	err = storage.EraseAll()
	assert.Nil(t, err, "Expected no error")


	resp, err = mw.Response(splash.Request{
		URL:    "http://example.com",
		Params: "", Cookies: "", Func: "",
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, resp.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")
	

	err = storage.EraseAll()
	assert.Nil(t, err, "Expected no error")


}
