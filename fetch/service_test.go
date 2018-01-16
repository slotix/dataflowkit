package fetch

import (
	"testing"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)
}

func TestFetchService(t *testing.T) {
	var svc Service
	svc = FetchService{}
	response, err := svc.Response(splash.Request{
		URL:    "http://example.com",
		Params: "", Cookies: "", Func: "",
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, response.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")


	response, err = svc.Fetch(splash.Request{
		URL:    "http://example.com",
		Params: "", Cookies: "", Func: "",
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, response.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")
	

	response, err = svc.Response(BaseFetcherRequest{
		URL:    "http://example.com",
		Method: "GET",
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, "200 OK", response.(*BaseFetcherResponse).Status, "Expected  200 status code")

	


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
