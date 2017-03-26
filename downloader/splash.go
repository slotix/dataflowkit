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

type SplashResponse struct {
	Obj struct {
		HeadersSize int    `json:"headersSize"`
		URL         string `json:"url"`
		Ok          bool   `json:"ok"`
		Content     struct {
			Text     string `json:"text"`
			Encoding string `json:"encoding"`
			Size     int    `json:"size"`
			MimeType string `json:"mimeType"`
		} `json:"content"`
		Cookies []struct {
			Expires  time.Time `json:"expires"`
			Name     string    `json:"name"`
			HTTPOnly bool      `json:"httpOnly"`
			Path     string    `json:"path"`
			Value    string    `json:"value"`
			Domain   string    `json:"domain"`
			Secure   bool      `json:"secure"`
		} `json:"cookies"`
		StatusText  string `json:"statusText"`
		HTTPVersion string `json:"httpVersion"`
		RedirectURL string `json:"redirectURL"`
		Headers     []struct {
			Value string `json:"value"`
			Name  string `json:"name"`
		} `json:"headers"`
		Status int `json:"status"`
	} `json:"1"`
}

//NewSplashConn opens new connection to Splash Server
func NewSplashConn(host, user, password string, timeout, resourceTimeout, wait int, lua string) SplashConn {

	return SplashConn{
		//	config:     cnf,
		host:            host,
		user:            user,
		password:        password,
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
		if !sResponse.Obj.Ok {
			err = fmt.Errorf("Error: %d. %s",
				sResponse.Obj.Status,
				sResponse.Obj.StatusText)
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
	decoded, err := base64.StdEncoding.DecodeString(r.Obj.Content.Text)
	if err != nil {
		logger.Println("decode error:", err)
		//return nil, fmt.Errorf(string(res))
		return nil, err
	}
	return decoded, nil
}

/*
func (s *SplashConn) Download1(addr string) ([]byte, error) {
	client := &http.Client{}
	splashURL := fmt.Sprintf("%s%s?&url=%s&timeout=%d&resource_timeout=%d&wait=%d", s.host, s.renderHTMLURL, url.QueryEscape(addr), s.timeout, s.resourceTimeout, s.wait)
	req, err := http.NewRequest("GET", splashURL, nil)
	req.SetBasicAuth(s.user, s.password)
	req.Header.Set("Content-Type", "text/plain")
	//fmt.Println(req.Header.Get("Content-Type"))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//resp.Header.Set("Content-Type", "text/plain")
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)

	//fmt.Println("CONTENT", http.DetectContentType(res))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		return res, nil
	}
	return nil, fmt.Errorf(string(res))
}
*/
