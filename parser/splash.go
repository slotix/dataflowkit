package parser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type SplashConn struct {
	host            string
	renderHTMLURL   string
	user            string
	password        string
	timeout         int
	resourceTimeout int
	wait            int
}

//NewSplashConn opens new connection to Splash Server 
func NewSplashConn(host, renderHTMLURL, user, password string, timeout, resourceTimeout, wait int) SplashConn {

	return SplashConn{
		//	config:     cnf,
		host:            host,
		renderHTMLURL:   renderHTMLURL,
		user:            user,
		password:        password,
		timeout:         timeout,
		resourceTimeout: resourceTimeout,
		wait:            wait,
	}
}

func (s *SplashConn) getHTML(addr string) ([]byte, error) {
	client := &http.Client{}
	splashURL := fmt.Sprintf("%s%s?&url=%s&timeout=%d&resource_timeout=%d&wait=%d", s.host, s.renderHTMLURL, url.QueryEscape(addr), s.timeout, s.resourceTimeout, s.wait)
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
	if resp.StatusCode == 200 {
		return res, nil
	}
	return nil, fmt.Errorf(string(res))
}
