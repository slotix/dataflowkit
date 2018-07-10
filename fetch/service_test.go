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

var (
	st            storage.Store
	tsURL         string
	robotsContent = "\n\t\tUser-agent: *\n\t\tAllow: /allowed\n\t\tDisallow: /disallowed\n\t\tDisallow: /redirect\n\t\t"
	helloContent  = []byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`)
)

func init() {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)
	viper.Set("STORAGE_TYPE", "Diskv")
	viper.Set("DFK_FETCH", "127.0.0.1:8000")
	viper.Set("PROXY", "")
	st = storage.NewStore(viper.GetString("STORAGE_TYPE"))
	tsURL = "http://localhost:12345"
}

func TestFetchService(t *testing.T) {
	var svc Service
	svc = FetchService{}
	svc = RobotsTxtMiddleware()(svc)
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
	rec := storage.Record{
		Key:     userToken,
		Type:    "Cookies",
		Value:   cookies,
		ExpTime: 0,
	}
	err = st.Write(rec)
	if err != nil {
		t.Log(err)
	}

	//BaseFetcher
	// r := mux.NewRouter()
	// r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Conent-Type", "text/html")
	// 	w.Write(IndexContent)
	// })
	// ts := httptest.NewServer(r)
	// defer ts.Close()

	resp, err := svc.Response(BaseFetcherRequest{
		URL:       tsURL + "/hello",
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, resp, "Expected response is not nil")

	//read cookies
	resp, err = svc.Response(BaseFetcherRequest{
		URL:       tsURL,
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, resp, "Expected response is not nil")

	//invalid URL
	resp, err = svc.Response(BaseFetcherRequest{
		URL:    "invalid_addr",
		Method: "GET",
	})
	assert.Error(t, err, "Expected error")

	//disallowed by robots
	resp, err = svc.Response(BaseFetcherRequest{
		URL:       tsURL + "/disallowed",
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Error(t, err, "Expected error")

	//disallowed by robots
	resp, err = svc.Response(BaseFetcherRequest{
		URL:       tsURL + "/redirect",
		Method:    "GET",
		UserToken: "12345",
	})

	assert.Error(t, err, "Expected error")

	//Splash Fetcher
	svcSplash := FetchService{}
	response, err := svcSplash.Response(splash.Request{
		URL:       "http://testserver:12345",
		FormData:  "",
		LUA:       "",
		UserToken: userToken,
	})
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, 200, response.(*splash.Response).Response.Status, "Expected Splash server returns 200 status code")

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
