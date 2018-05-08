package fetch

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var st storage.Store

func init() {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)
	viper.Set("STORAGE_TYPE", "Diskv")
	storageType, err := storage.TypeString(viper.GetString("STORAGE_TYPE"))
	if err != nil {
		logger.Error(err)
	}
	st = storage.NewStore(storageType)
}

func TestFetchService(t *testing.T) {
	var svc Service
	svc = FetchService{}
	cArr := []*http.Cookie{
		&http.Cookie{
			Name:   "cookie1",
			Value:  "cValue1",
			Domain: "example.com",
		},
		&http.Cookie{
			Name:   "cookie2",
			Value:  "cValue2",
			Domain: "example.com",
		},
	}
	userToken := "12345"
	cookies, err := json.Marshal(cArr)
	err = st.Write(userToken, cookies, 0)
	if err != nil {
		t.Log(err)
	}
	//Splash Fetcher
	response, err := svc.Response(splash.Request{
		URL:       "http://example.com",
		FormData:  "",
		LUA:       "",
		UserToken: userToken,
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, response.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")

	//BaseFetcher
	content, err := svc.Fetch(BaseFetcherRequest{
		URL:       "http://" + addr,
		Method:    "GET",
		UserToken: "123456",
	})

	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, content, "Expected content is not nill")
	content, err = svc.Fetch(BaseFetcherRequest{
		URL:    "invalid_addr",
		Method: "GET",
	})

	assert.Error(t, err, "Expected error")
}

/* func TestEncodeSplashFetcherContent(t *testing.T) {
	ctx := context.Background()
	resp := splash.Response{
		HTML: `<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`,
	}
	w := httptest.NewRecorder()

	EncodeSplashFetcherContent(ctx, w, resp)
	//r := w.Code
	//r := w.Result()
	logger.Info(w)
} */
