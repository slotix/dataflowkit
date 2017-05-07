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

//GetResponse is needed to be passed to  httpcaching middleware
//to provide a RFC7234 compliant HTTP cache
func GetResponse(splashURL string) (*Response, error) {
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
	//if t == "[]interface {}" {
	r.Request.Headers = castHeaders(r.Request.Headers)
	r.Response.Headers = castHeaders(r.Response.Headers)
	//}
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
