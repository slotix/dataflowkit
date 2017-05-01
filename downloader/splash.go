package downloader

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

type SplashConn struct {
	Host            string
	User            string
	Password        string
	Timeout         int
	ResourceTimeout int
	Wait            int
	LUAScript       string
}

type Headers []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SplashResponse struct {
	HTML    string `json:"html"`
	Reason  string `json:"reason"`
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
	} `json:"request"`
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

//NewSplashConn creates new connection to Splash Server
func NewSplashConn(c SplashConn) (*SplashConn, error) {
	//resp, err := c.ping()
	//if err != nil {
	//	return nil, err
	//}
	//fmt.Println(resp)
	return &c, nil
}

func (s *SplashConn) GetResponse(req FetchRequest) (*SplashResponse, error) {
	client := &http.Client{}
	splashURL := fmt.Sprintf(
		"%sexecute?url=%s&timeout=%d&resource_timeout=%d&wait=%d&lua_source=%s", s.Host,
		neturl.QueryEscape(req.URL),
		s.Timeout,
		s.ResourceTimeout,
		s.Wait,
		neturl.QueryEscape(s.LUAScript))

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
	readCloser := ioutil.NopCloser(strings.NewReader(r.HTML))
	return readCloser, nil
	//return []byte(r.HTML), nil
}
