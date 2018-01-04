package splash

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)
}

func TestSplashRenderHTMLEndpoint(t *testing.T) {
	//Splash running inside Docker container cannot render a page on a localhost. It leads to rendering page errors https://github.com/scrapinghub/splash/issues/237 .
	//Only URLs on the web are available for testing.
	sReq := []byte(`{"url": "http://example.com", "wait": 0.5}`)
	reader := bytes.NewReader(sReq)
	splashExecuteURL := "http://" + viper.GetString("SPLASH") + "/render.html"
	client := &http.Client{}
	req, err := http.NewRequest("POST", splashExecuteURL, reader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Error(err)
	}
	statusCode := resp.StatusCode
	assert.Equal(t, statusCode, 200)
	//	logger.Info("Status code:", statusCode)

	//res, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	logger.Error(err)
	//}
	//logger.Info(string(res))

}
