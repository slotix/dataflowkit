package downloader

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"time"
)

type SplashConn struct {
	host            string
	user            string
	password        string
	timeout         int
	resourceTimeout int
	wait            int
	lua             string
}

type Headers []struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

type SplashResponse struct {
	Request struct {
		Cookies     []interface{} `json:"cookies"`
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
	} `json:"1"`
	Response struct {
		Headers Headers `json:"headers"`
		Cookies []struct {
			Name     string    `json:"name"`
			Value    string    `json:"value"`
			Expires  time.Time `json:"expires"`
			Domain   string    `json:"domain"`
			Secure   bool      `json:"secure"`
			Path     string    `json:"path"`
			HTTPOnly bool      `json:"httpOnly"`
		} `json:"cookies"`
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
	} `json:"2"`
}

//NewSplashConn opens new connection to Splash Server
func NewSplashConn(host string, timeout, resourceTimeout, wait int, lua string) SplashConn {

	return SplashConn{
		//	config:     cnf,
		host:            host,
		timeout:         timeout,
		resourceTimeout: resourceTimeout,
		wait:            wait,
		lua:             lua,
	}
}

func (s *SplashConn) GetResponse(url string) (*SplashResponse, error) {
	client := &http.Client{}
	//splashURL := fmt.Sprintf("%s%s?&url=%s&timeout=%d&resource_timeout=%d&wait=%d", s.host, s.renderHTMLURL, url.QueryEscape(addr), s.timeout, s.resourceTimeout, s.wait)

	splashURL := fmt.Sprintf(
		"%sexecute?url=%s&timeout=%d&resource_timeout=%d&wait=%d&lua_source=%s", s.host,
		neturl.QueryEscape(url),
		s.timeout,
		s.resourceTimeout,
		s.wait,
		neturl.QueryEscape(s.lua))

	req, err := http.NewRequest("GET", splashURL, nil)
	req.SetBasicAuth(s.user, s.password)

	resp, err := client.Do(req)
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
		var sResponse SplashResponse
		if err := json.Unmarshal(res, &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		//if response returned by Splash is bad
		/*
			if !sResponse.Obj.Ok {
				return nil, fmt.Errorf("Error: %d. %s",
					sResponse.Obj.Status,
					sResponse.Obj.StatusText)
			}*/
		//return &sResponse, nil
		if !sResponse.Response.Ok {
			err = fmt.Errorf("Error: %d. %s",
				sResponse.Response.Status,
				sResponse.Response.StatusText)
		} else {
			err = nil
		}
		return &sResponse, err
	}
	return nil, fmt.Errorf(string(res))

}

func (r *SplashResponse) GetHTML() ([]byte, error) {
	if r == nil {
		logger.Println("empty response ")
		return nil, errors.New("empty response")
	}
	decoded, err := base64.StdEncoding.DecodeString(r.Response.Content.Text)
	if err != nil {
		logger.Println("decode error:", err)
		//return nil, fmt.Errorf(string(res))
		return nil, err
	}
	return decoded, nil
}
