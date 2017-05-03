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

	"github.com/spf13/viper"
)

var logger *log.Logger

// ErrURLEmpty is returned when an input string is empty.
var errURLEmpty = errors.New("URL is empty")

func init() {
	logger = log.New(os.Stdout, "splash: ", log.Lshortfile)
}

type SplashConn struct {
	Host            string
	User            string
	Password        string
	Timeout         int
	ResourceTimeout int
	LUAScript       string
}

type Headers []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FetchRequest struct {
	URL    string  `json:"url"`
	Params string  `json:"params,omitempty"` //params used for passing formdata to LUA script
	Cookie string  `json:"cookie,omitempty"` //add cookie to request
	Func   string  `json:"func,omitempty"`
	Wait   float64 `json:"wait,omitempty"` //Time in seconds to wait until js scripts loaded. Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
}

type Cookies []struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Expires  time.Time `json:"expires,omitempty"`
	Domain   string    `json:"domain"`
	Secure   bool      `json:"secure"`
	Path     string    `json:"path"`
	HTTPOnly bool      `json:"httpOnly"`
}

type SplashResponse struct {
	HTML    string `json:"html"`
	Reason  string `json:"reason"`
	Cookies  Cookies`json:"cookies"`
	Request struct {
		Cookies     Cookies `json:"cookies"`
		Method      string        `json:"method"`
		HeadersSize int           `json:"headersSize"`
		URL         string        `json:"url"`
		HTTPVersion string        `json:"httpVersion"`
		QueryString []struct {
			Value string `json:"value"`
			Name  string `json:"name"`
		} `json:"queryString"`
		Headers  Headers `json:"headers"`
		BodySize int     `json:"bodySize"`
	} `json:"request"`
	Response struct {
		Headers Headers `json:"headers"`
		Cookies  Cookies`json:"cookies"`
		HeadersSize int  `json:"headersSize"`
		Ok          bool `json:"ok"`
		Content     struct {
			Text     string `json:"text"`
			MimeType string `json:"mimeType"`
			Size     int    `json:"size"`
			Encoding string `json:"encoding"`
		} `json:"content"`
		Status      int    `json:"status"`
		URL         string `json:"url"`
		HTTPVersion string `json:"httpVersion"`
		StatusText  string `json:"statusText"`
		RedirectURL string `json:"redirectURL"`
	} `json:"response"`
}

type SplashPingResponse struct {
	Maxrss int    `json:"maxrss"`
	Status string `json:"status"`
}

//Ping returns status and maxrss from _ping  endpoint
func Ping(host string) (*SplashPingResponse, error) {
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
	var p SplashPingResponse
	err = json.Unmarshal(buf.Bytes(), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetLUA(req FetchRequest) string {
	if isRobotsTxt(req.URL) {
		return robotsLUA
	}
	if req.Wait == 0 {
		req.Wait = viper.GetFloat64("splash-wait")
	}
	if req.Params == "" {
		return fmt.Sprintf(baseLUA, req.Wait)
	}
	if req.Cookie != "" {
		return fmt.Sprintf(LUASetCookie, req.Wait)
	}
	return fmt.Sprintf(LUAPostFormData, paramsToLuaTable(req.Params), req.Wait)
}

//NewSplashConn creates new connection to Splash Server
func NewSplashConn(req FetchRequest) (splashURL string, err error) {
	if req.URL == "" {
		return "", errURLEmpty
	}
	splashURL = fmt.Sprintf(
		"%sexecute?url=%s&timeout=%d&resource_timeout=%d&lua_source=%s", fmt.Sprintf("http://%s/", viper.GetString("splash")),
		neturl.QueryEscape(req.URL),
		viper.GetInt("splash-timeout"),
		viper.GetInt("splash-resource-timeout"),
		neturl.QueryEscape(GetLUA(req)))
	return splashURL, nil
}

//GetResponse is needed to be passed to  httpcaching middleware
//to provide a RFC7234 compliant HTTP cache
func GetResponse(splashURL string) (*SplashResponse, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", splashURL, nil)
	//req.SetBasicAuth(s.user, s.password)
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//var res []byte
	//_, err = resp.Body.Read(res)
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//response from Splash service
	if resp.StatusCode == 200 {
		var sResponse SplashResponse
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

func (r *SplashResponse) GetContent() (io.ReadCloser, error) {
	if r == nil {
		logger.Println("empty response ")
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
	cookielua, err := r.cookieToLUATable()
	if err != nil {
		logger.Println(err)
	}
	logger.Println(cookielua)
	readCloser := ioutil.NopCloser(strings.NewReader(r.HTML))
	return readCloser, nil
	//return []byte(r.HTML), nil
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
