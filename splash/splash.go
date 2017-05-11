package splash

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/spf13/viper"
)

var logger *log.Logger

// ErrURLEmpty is returned when an input string is empty.
var errURLEmpty = errors.New("URL is empty")

func init() {
	logger = log.New(os.Stdout, "splash: ", log.Lshortfile)
}

//Ping returns status and maxrss from _ping  endpoint
func Ping(host string) (*PingResponse, error) {
	client := &http.Client{}
	pingURL := fmt.Sprintf("http://%s/_ping", host)
	req, err := http.NewRequest("GET", pingURL, nil)
	//req.SetBasicAuth(userName, userPass)
	resp, err := client.Do(req)
	// Check Error
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

func GetLUA(req Request) string {
	if isRobotsTxt(req.URL) {
		return robotsLUA
	}
	if req.Wait == 0 {
		req.Wait = viper.GetFloat64("splash-wait")
	}
	if req.Params == "" {
		return fmt.Sprintf(baseLUA, req.Wait)
	}
	if req.Cookies != "" {
		return fmt.Sprintf(LUASetCookie, req.Wait)
	}
	return fmt.Sprintf(LUAPostFormData, paramsToLuaTable(req.Params), req.Wait)
}

//NewSplashConn creates new connection to Splash Server
func NewSplashConn(req Request) (splashURL string, err error) {
	if req.URL == "" {
		return "", errURLEmpty
	}
	splashURL = fmt.Sprintf(
		"%sexecute?url=%s&timeout=%d&resource_timeout=%d&lua_source=%s", fmt.Sprintf("http://%s/", viper.GetString("splash")),
		neturl.QueryEscape(req.URL),
		viper.GetInt("splash-timeout"),
		viper.GetInt("splash-resource-timeout"),
		neturl.QueryEscape(GetLUA(req)))
	//	logger.Println(splashURL)
	return splashURL, nil
}

//GetResponse result is passed to  caching middleware
//to provide a RFC7234 compliant HTTP cache
func GetResponse(splashURL string) (*Response, error) {
	//logger.Println(splashURL)
	client := &http.Client{}
	request, err := http.NewRequest("GET", splashURL, nil)
	//req.SetBasicAuth(s.user, s.password)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//response from Splash service
	if resp.StatusCode == 200 {
		var sResponse Response

		if err := json.Unmarshal(res, &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		if !sResponse.Response.Ok {
			if sResponse.Response.Status != 0 {
				err = fmt.Errorf("Error: %d. %s",
					sResponse.Response.Status,
					sResponse.Response.StatusText)
			} else {
				err = fmt.Errorf("Error: %s",
					sResponse.Reason)
			}
		} else {
			err = nil
		}
		//if cacheable ?
		rv := sResponse.cacheable()
		if len(rv.OutReasons) == 0 {
			sResponse.Cacheable = true
		}
		return &sResponse, err
	}
	return nil, fmt.Errorf(string(res))

}

func (r *Response) GetContent() (io.ReadCloser, error) {
	if r == nil {
		//	logger.Println("empty response")
		return nil, errors.New("empty response")
	}
	if isRobotsTxt(r.Request.URL) {
		decoded, err := base64.StdEncoding.DecodeString(r.Response.Content.Text)
		if err != nil {
			logger.Println("decode error:", err)
			//return nil, fmt.Errorf(string(res))
			return nil, err
		}
		readCloser := ioutil.NopCloser(bytes.NewReader(decoded))
		//r := bytes.NewReader(decoded)
		return readCloser, nil
	}
	cookielua, err := r.setCookieToLUATable()
	if err != nil {
		logger.Println(err)
	}
	logger.Println(cookielua)
	readCloser := ioutil.NopCloser(strings.NewReader(r.HTML))
	return readCloser, nil
}

//cacheable check if resource is cacheable
func (r *Response) cacheable() (rv cacheobject.ObjectResults) {
//func (r *Response) cacheable() bool {

	respHeader := r.Response.Headers.(http.Header)
	reqHeader := r.Request.Headers.(http.Header)

	reqDir, err := cacheobject.ParseRequestCacheControl(reqHeader.Get("Cache-Control"))
	if err != nil {
		logger.Printf(err.Error())
	}
	resDir, err := cacheobject.ParseResponseCacheControl(respHeader.Get("Cache-Control"))
	if err != nil {
		logger.Printf(err.Error())
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

		NowUTC: time.Now().UTC(),
	}

	rv = cacheobject.ObjectResults{}
	cacheobject.CachableObject(&obj, &rv)
	cacheobject.ExpirationObject(&obj, &rv)
	//Check if it is cacheable

	expTime := rv.OutExpirationTime.Unix()
	if rv.OutExpirationTime.IsZero() {
		expTime = 0
	}
	r.CacheExpirationTime = expTime
	debug := false
	if debug {
		if rv.OutErr != nil {
			logger.Println("Errors: ", rv.OutErr)
		}
		if rv.OutReasons != nil {
			logger.Println("Reasons to not cache: ", rv.OutReasons)
		}
		if rv.OutWarnings != nil {
			logger.Println("Warning headers to add: ", rv.OutWarnings)
		}
		logger.Println("Expiration: ", rv.OutExpirationTime.String())
	}
	return rv
}

//Fetch content from url through Splash server https://github.com/scrapinghub/splash/
func Fetch(splashURL string) (io.ReadCloser, error) {
	response, err := GetResponse(splashURL)
	if err != nil {
		return nil, err
	}
	content, err := response.GetContent()
	if err == nil {
		return content, nil
	}
	return nil, err
}

func isRobotsTxt(url string) bool {
	if strings.Contains(url, "robots.txt") {
		return true
	}
	return false
}

//http://choly.ca/post/go-json-marshalling/
//UnmarshalJSON convert headers to http.Header type while unmarshaling
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

	//logger.Println(t)

	r.Request.Headers = castHeaders(r.Request.Headers)
	r.Response.Headers = castHeaders(r.Response.Headers)
	/*
		switch r.Request.URL.(type) {
		case string:
			parsedURL, err := neturl.Parse(r.Request.URL.(string))
			if err != nil {
				return err
			}
			r.Request.URL = parsedURL
		case map[string]interface{}:
			for k, v := range r.Request.URL.(map[string]interface{}) {
				logger.Println(k, v)
				r.Request.URL.(map[string]interface{})[k] = v
			}
		}
	*/
	return nil
}

//castHeaders serves for casting headers to standard http.Header type
func castHeaders(splashHeaders interface{}) (header http.Header) {
	//t := fmt.Sprintf("%T", splashHeaders)
	//	logger.Printf("%T - %s", splashHeaders, splashHeaders)
	header = make(map[string][]string)
	switch splashHeaders.(type) {
	case []interface{}:
		for _, h := range splashHeaders.([]interface{}) {
			var str []string
			v := h.(map[string]interface{})["value"].(string)
			str = append(str, v)
			header[h.(map[string]interface{})["name"].(string)] = str
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
		logger.Println()
		return nil
	}
}
