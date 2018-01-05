package splash

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var resp *Response

func init() {
	viper.Set("SPLASH", "127.0.0.1:8050")
	viper.Set("SPLASH_TIMEOUT", 20)
	viper.Set("SPLASH_RESOURCE_TIMEOUT", 30)
	viper.Set("SPLASH_WAIT", 0.5)

	//Splash running inside Docker container cannot render a page on a localhost. It leads to rendering page errors https://github.com/scrapinghub/splash/issues/237 .
	//Only URLs on the web are available for testing.
	req := Request{
		URL: "http://example.com",
	}
	resp, _ = req.GetResponse()
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

func TestGetResponse(t *testing.T) {
	statusCode := resp.Response.Status
	assert.Equal(t, statusCode, 200)
	respURL := resp.GetURL()
	assert.Equal(t, respURL, "http://example.com/")
	expires := resp.GetExpires()
	tp := fmt.Sprintf("%T", expires)
	assert.Equal(t, "time.Time", tp)
	reasons := resp.GetReasonsNotToCache()
	logger.Info(reasons)
}

func TestGetContent(t *testing.T) {
	resp := Response{
		HTML: `<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`,
	}
	readCloser, _ := resp.GetContent()
	buf := new(bytes.Buffer)
	buf.ReadFrom(readCloser)
	s := buf.String()
	assert.Equal(t, `<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`, s)
}

func TestReqGetURL(t *testing.T) {
	req := Request{
		URL: "   http://example.com/	",
	}
	assert.Equal(t, "http://example.com", req.GetURL())
}

func TestHost(t *testing.T) {
	req := Request{
		URL: "   http://example.com/	",
	}
	host, _ := req.Host()
	assert.Equal(t, "example.com", host)
}

func TestPing(t *testing.T) {
	host := viper.GetString("SPLASH")
	pr, _ := Ping(host)
	assert.Equal(t, "ok", pr.Status)
}

func TestParamsToLUATable(t *testing.T) {
	params := "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fexample.com%2F&ips_username=user&ips_password=password&rememberMe=1"
	p := paramsToLuaTable(params)
	assert.Equal(t, `{"auth_key":"880ea6a14ea49e853634fbdc5015a024","referer":"http%3A%2F%2Fexample.com%2F","ips_username":"user","ips_password":"password","rememberMe":"1"}`, p)

}

func TestGenerateCookie(t *testing.T) {
	cookie := `example_uzt=72e3502635d3af8fa2916cf397e93fee; expires=Tue, 04-Jul-2017 13:28:36 GMT; Max-Age=2592000; path=/; domain=.example.com; HttpOnly
heureka_s=1; expires=Mon, 04-Jun-2018 13:28:36 GMT; Max-Age=31536000; path=/; domain=.example.com`
	s, err := generateCookie(cookie)
	if err != nil {
		logger.Println(nil)
	}
	assert.Equal(t, `[{"name":"example_uzt", "value":"72e3502635d3af8fa2916cf397e93fee", "path":"/", "domain":".example.com", "expires":"Tue, 04-Jul-2017 13:28:36 GMT", "httpOnly":true, "secure":false},{"name":"heureka_s", "value":"1", "path":"/", "domain":".example.com", "expires":"Mon, 04-Jun-2018 13:28:36 GMT", "httpOnly":false, "secure":false}]`, s)
}




//Script for LUA debug
var debugLUA = `
treat = require("treat")
function string.ends(String,End)
  return End=='' or string.sub(String,-string.len(End))==End
end
function remove_trailing_slash(text)
  if string.ends(text, "/") then
    text = text:sub(1, -2)
  end
  return text
end

function main(splash)
  local url = splash.args.url
  local urls = {}
	local resp_urls = {}
  splash:on_response(function (response)
  url = remove_trailing_slash(url)
  resp_url = remove_trailing_slash(response.info.url)
	table.insert(urls, url)
	table.insert(resp_urls, resp_url)
  if resp_url == url then
    status = response.info.status
    is_redirect = status == 301 or status == 302
    if is_redirect then
      url = response.info.redirectURL
    elseif status == 200 then
  			r = response
    end
  end
  end)
  local ok, reason = splash:go(url)
  assert(splash:wait(1))
  if not ok then
       return {
        reason = reason,
     --   request = r.request.info,
     --   response = r.info,
      }
  end
  return {
     -- request = r.request.info,
     -- response = r.info,
			urls = treat.as_array(urls),
			resp_urls = treat.as_array(resp_urls),
	    html = splash:html(),       
  } 
end
`

/*
func TestParamsToJSON(t *testing.T) {
	//params := "auth_key=test&referer=http%3A%2F%2Fexample.com%2F&ips_username=test&ips_password=test&rememberMe=1"
	params := "auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fdiesel.elcat.kg%2F&ips_username=dm_&ips_password=dmsoft&rememberMe=1"
//	logger.Println(paramsToJSON(params))
	assert.Equal(t,
	`{"auth_key":["880ea6a14ea49e853634fbdc5015a024"],"ips_password":["dmsoft"],"ips_username":["dm_"],"referer":["http://diesel.elcat.kg/"],"rememberMe":["1"]}`, paramsToJSON(params))

}
*/
