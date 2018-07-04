package splash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	neturl "net/url"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/sirupsen/logrus"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/logger"
	"github.com/spf13/viper"
)

var logger *logrus.Logger

func init() {
	viper.AutomaticEnv() // read in environment variables that match
	logger = log.NewLogger(true)
}

// Options struct inclued parameters for Splash Connection
type Options struct {
	host  string //splash host address
	proxy string
	//Splash connection parameters:
	timeout         int
	resourceTimeout int
	// Time in seconds to wait until java scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
	wait float64
	LUA  string
}

//NewSplash creates new connection to Splash Server
//func NewSplash(req Request, setters ...Option) (splashURL string) {
func NewSplash(req Request) (splashURL string) {
	//Default options

	args := &Options{
		host:            viper.GetString("SPLASH"),
		timeout:         viper.GetInt("SPLASH_TIMEOUT"),
		resourceTimeout: viper.GetInt("SPLASH_RESOURCE_TIMEOUT"),
		wait:            viper.GetFloat64("SPLASH_WAIT"),
		proxy:           viper.GetString("PROXY"),
	}
	/*args.host = "127.0.0.1:8050"
	args.timeout = 20
	args.resourceTimeout = 30
	args.wait = 1.0
	*/

	//Generating Splash URL
	/*
	   	//"Set-Cookie" from response headers should be sent when accessing for the same domain second time
	   	cookie := `PHPSESSID=ef75e2737a14b06a2749d0b73840354f; path=/; domain=.acer-a500.ru; HttpOnly
	   dle_user_id=deleted; expires=Thu, 01-Jan-1970 00:00:01 GMT; Max-Age=0; path=/; domain=.acer-a500.ru; httponly
	   dle_password=deleted; expires=Thu, 01-Jan-1970 00:00:01 GMT; Max-Age=0; path=/; domain=.acer-a500.ru; httponly
	   dle_hash=deleted; expires=Thu, 01-Jan-1970 00:00:01 GMT; Max-Age=0; path=/; domain=.acer-a500.ru; httponly
	   dle_forum_sessions=ef75e2737a14b06a2749d0b73840354f; expires=Wed, 06-Jun-2018 19:13:00 GMT; Max-Age=31536000; path=/; domain=.acer-a500.ru; httponly
	   forum_last=1496801580; expires=Wed, 06-Jun-2018 19:13:00 GMT; Max-Age=31536000; path=/; domain=.acer-a500.ru; httponly`
	   	//cookie := ""

	   	req.Cookies, err = generateCookie(cookie)
	   	if err != nil {
	   		logger.Println(err)
	   	}
	   	//logger.Println(req.Cookies)

	   	//---------
	*/
	//req.Params = `"auth_key=880ea6a14ea49e853634fbdc5015a024&referer=http%3A%2F%2Fdiesel.elcat.kg%2F&ips_username=dm_&ips_password=asfwwe!444D&rememberMe=1"`
	var LUAScript string
	if req.LUA != "" {
		LUAScript = req.LUA
	} else {
		LUAScript = neturl.QueryEscape(baseLUA)
	}

	cookies := ""
	if len(req.Cookies) > 0 {
		cookies = `{"Cookie":"`
		for _, c := range req.Cookies {
			cookies += fmt.Sprintf(`%s=%s;`,
				c.Name, strings.Replace(c.Value, `"`, `\"`, -1))
		}
		cookies = strings.TrimSuffix(cookies, ";") + `"}`
	}

	var infiniteScroll string
	if req.InfiniteScroll {
		infiniteScroll = "true"
	} else {
		infiniteScroll = "false"
	}

	splashURL = fmt.Sprintf(
		"http://%s/execute?url=%s&proxy=%s&timeout=%d&resource_timeout=%d&wait=%.1f&headers=%s&formdata=%s&lua_source=%s&scroll2bottom=%s",
		args.host,
		neturl.QueryEscape(req.URL),
		neturl.QueryEscape(args.proxy),
		args.timeout,
		args.resourceTimeout,
		args.wait,
		//neturl.QueryEscape(req.Cookies),
		neturl.QueryEscape(cookies),
		neturl.QueryEscape(paramsToLuaTable(req.FormData)),
		LUAScript,
		infiniteScroll)

	return
}

//GetResponse result is passed to storage middleware
//to provide a RFC7234 compliant HTTP cache
func (req Request) GetResponse() (*Response, error) {
	//URL validation
	if _, err := url.ParseRequestURI(strings.TrimSpace(req.URL)); err != nil {
		return nil, &errs.BadRequest{err}
	}
	splashURL := NewSplash(req)
	client := &http.Client{}
	request, err := http.NewRequest("GET", splashURL, nil)
	//req.SetBasicAuth(s.user, s.password)
	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//response from Splash service.
	statusCode := resp.StatusCode
	if statusCode != 200 {
		switch statusCode {
		case 504:
			return nil, &errs.GatewayTimeout{}
		default:
			return nil, fmt.Errorf(string(res))
		}
	}

	var sResponse Response

	if err := json.Unmarshal(res, &sResponse); err != nil {
		logger.Error("Json Unmarshall error", err)
	}

	//if response status code is not 200
	if sResponse.Error != "" {
		switch sResponse.Error {
		case "http404":
			return nil, &errs.NotFound{req.URL}
		case "http403":
			return nil, &errs.Forbidden{req.URL}
		case "network3":
			return nil, &errs.BadRequest{errors.New(sResponse.Error)}
		default:
			return nil, &errs.Error{sResponse.Error}
		}
		//return nil, fmt.Errorf("%s", sResponse.Error)
	}
	// Sometimes no response, request returned  by Splash.
	// gc (garbage collection) method should be called to clear WebKit caches and then
	// GetResponse again. See more at https://github.com/scrapinghub/splash/issues/613

	if sResponse.Response == nil || sResponse.Request == nil && sResponse.HTML != "" {

		var response *Response
		gcResponse, err := gc(viper.GetString("SPLASH"))
		if err == nil && gcResponse.Status == "ok" {
			response, err = req.GetResponse()
			if err != nil {
				return nil, err
			}
		}
		return response, nil
	}

	if !sResponse.Response.Ok {
		// if sResponse.Response.Status == 0 {
		// 	err = fmt.Errorf("%s",
		// 		//sResponse.Error)
		// 		sResponse.Response.StatusText)
		// } else {
		// 	err = fmt.Errorf("%d. %s",
		// 		sResponse.Response.Status,
		// 		sResponse.Response.StatusText)
		// }
		return nil, err
	}
	//is resource cacheable ?
	sResponse.SetCacheInfo()

	return &sResponse, nil

}

// GetHTML returns HTML content from Splash Response
func (r *Response) GetHTML() (io.ReadCloser, error) {
	if r == nil {
		return nil, errors.New("empty response")
	}
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	/* if IsRobotsTxt(r.Request.URL) {
		decoded, err := base64.StdEncoding.DecodeString(r.Response.Content.Text)
		if err != nil {
			logger.Println("decode error:", err)
			return nil, err
		}
		readCloser := ioutil.NopCloser(bytes.NewReader(decoded))
		return readCloser, nil
	} */

	readCloser := ioutil.NopCloser(strings.NewReader(r.HTML))
	return readCloser, nil
}

// SetCacheInfo checks if resource is cacheable.
// Respource is cachable if length of ReasonsNotToCache is zero.
// ReasonsNotToCache and Expires values are filled here.
func (r *Response) SetCacheInfo() {
	respHeader := r.Response.Headers.(http.Header)
	reqHeader := r.Request.Headers.(http.Header)
	//	respHeader := r.Response.castHeaders()
	//	reqHeader := r.Request.castHeaders()

	reqDir, err := cacheobject.ParseRequestCacheControl(reqHeader.Get("Cache-Control"))
	if err != nil {
		logger.Error(err.Error())
	}
	resDir, err := cacheobject.ParseResponseCacheControl(respHeader.Get("Cache-Control"))
	if err != nil {
		logger.Error(err.Error())
	}
	//logger.Println(respHeader)
	expiresHeader, _ := http.ParseTime(respHeader.Get("Expires"))
	dateHeader, _ := http.ParseTime(respHeader.Get("Date"))
	lastModifiedHeader, _ := http.ParseTime(respHeader.Get("Last-Modified"))
	obj := cacheobject.Object{
		//	CacheIsPrivate:         false,
		RespDirectives:         resDir,
		RespHeaders:            respHeader,
		RespStatusCode:         r.Response.Status,
		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,

		ReqDirectives: reqDir,
		ReqHeaders:    reqHeader,
		ReqMethod:     r.Request.Method,
		NowUTC:        time.Now().UTC(),
	}

	rv := cacheobject.ObjectResults{}
	cacheobject.CachableObject(&obj, &rv)
	cacheobject.ExpirationObject(&obj, &rv)

	//Check if it is cacheable

	if len(rv.OutReasons) == 0 {
		//r.Cacheable = true
		if rv.OutExpirationTime.IsZero() {
			//if time is zero than set it to current time plus 24 hours.
			r.Expires = time.Now().UTC().Add(time.Hour * 24)
		} else {
			r.Expires = rv.OutExpirationTime
		}
		//logger.Info("Current Time: ", time.Now().UTC())
		//logger.Info(r.Request.URL, r.Expires)
	} else {
		//if resource is not cacheable set expiration time to the current time.
		//This way web page will be downloaded every time.
		r.Expires = time.Now().UTC()
		r.ReasonsNotToCache = rv.OutReasons
	}

}

// GetURL returns URL from Request
func (req Request) GetURL() string {
	//trim trailing slash if any.
	//aws s3 bucket item name cannot contain slash at the end.
	return strings.TrimRight(strings.TrimSpace(req.URL), "/")
}

//  GetFormData returns form data from Splash Request
func (req Request) GetFormData() string {
	return req.FormData
}

// Host returns Host value from Request
func (req Request) Host() (string, error) {
	u, err := url.Parse(req.GetURL())
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

//Type return fetcher type
func (req Request) Type() string {
	return "splash"
}

// UnmarshalJSON convert headers to http.Header type
// http://choly.ca/post/go-json-marshalling/
func (r *Response) UnmarshalJSON(data []byte) error {
	type Alias Response
	aux := &struct {
		Headers interface{} `json:"headers"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if r.Request != nil {
		r.Request.Headers = castHeaders(r.Request.Headers)
	}
	if r.Response != nil {
		r.Response.Headers = castHeaders(r.Response.Headers)
	}
	return nil
}

//castHeaders serves for casting headers returned by Splash to standard http.Header type
func castHeaders(splashHeaders interface{}) (header http.Header) {
	header = make(map[string][]string)
	switch splashHeaders.(type) {
	case []interface{}:
		for _, h := range splashHeaders.([]interface{}) {
			//var str []string
			str := []string{}
			v, ok := h.(map[string]interface{})["value"].(string)
			if ok {
				str = append(str, v)
				header[h.(map[string]interface{})["name"].(string)] = str
			}
		}
		return header
	case map[string]interface{}:
		for k, v := range splashHeaders.(map[string]interface{}) {
			var str []string
			for _, vv := range v.([]interface{}) {
				str = append(str, vv.(string))
			}
			header[k] = str
		}
		return header
	default:
		return nil
	}
}

//Splash splash:history() may return empty list as it cannot retrieve cached results
//in case request the same page twice. So the only solution for the moment is to call
//http://localhost:8050/_gc if an error occures. It runs the Python garbage collector
//and clears internal WebKit caches. See more at https://github.com/scrapinghub/splash/issues/613
func gc(host string) (*gcResponse, error) {
	client := &http.Client{}
	gcURL := fmt.Sprintf("http://%s/_gc", host)
	req, err := http.NewRequest("POST", gcURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var gc gcResponse
	err = json.Unmarshal(buf.Bytes(), &gc)
	if err != nil {
		return nil, err
	}
	return &gc, nil
}

//Ping returns status and maxrss from _ping  endpoint
func Ping(host string) (*PingResponse, error) {
	client := &http.Client{}
	pingURL := fmt.Sprintf("http://%s/_ping", host)
	req, err := http.NewRequest("GET", pingURL, nil)
	if err != nil {
		return nil, err
	}
	//req.SetBasicAuth(userName, userPass)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var p PingResponse
	err = json.Unmarshal(buf.Bytes(), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

//GetExpires returns Response Expires value.
func (r Response) GetExpires() time.Time {
	return r.Expires
}

//GetReasonsNotToCache returns an array of reasons why a response should not be cached.
func (r Response) GetReasonsNotToCache() []cacheobject.Reason {
	return r.ReasonsNotToCache
}

//GetURL returns URL after all redirects
//TODO: test it
func (r Response) GetURL() string {
	return r.Response.URL
}

func (r Request) GetUserToken() string {
	return r.UserToken
}
