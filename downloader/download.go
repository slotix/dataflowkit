package downloader

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/net/html/charset"
)

type FetchRequest struct {
	URL    string `json:"url"`
	Params string `json:"params,omitempty"` //params used for passing formdata to LUA script
	Func   string `json:"func,omitempty"`
}

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "downloader: ", log.Lshortfile)
}

// ErrURLEmpty is returned when an input string is empty.
var errURLEmpty = errors.New("URL is empty")

type directConn struct {
}

func GetLUA(req FetchRequest, wait int) string {
	if isRobotsTxt(req.URL) {
		return robotsLUA
	}
	if req.Params == "" {
		return fmt.Sprintf(generalLUA, wait)
	}
	return fmt.Sprintf(generalLUAWithAuth, paramsToLuaTable(req.Params), wait)
}

//GetResponse is needed to be passed to  httpcaching middleware
//to provide a RFC7234 compliant HTTP cache
func GetResponse(req FetchRequest) (*SplashResponse, error) {
	if req.URL == "" {
		return nil, errURLEmpty
	}
	wait := viper.GetInt("splash-wait")
	conn := SplashConn{
		Host:            fmt.Sprintf("http://%s/", viper.GetString("splash")),
		Timeout:         viper.GetInt("splash-timeout"),
		ResourceTimeout: viper.GetInt("splash-resource-timeout"),
		Wait:            wait, //Sometimes wait parameter should be set to more than default 0,5. It allows to finish js scripts execution on a web page.
		LUAScript:       GetLUA(req, wait),
	}
	s, err := NewSplashConn(conn)
	if err != nil {
		return nil, err
	}
	response, err := s.GetResponse(req)
	return response, err
}

//Fetch content from url.
//If no data is pulled through Splash server https://github.com/scrapinghub/splash/ .
func Fetch(req FetchRequest) (io.ReadCloser, error) {
	response, err := GetResponse(req)
	if err != nil {
		return nil, err
	}
	content, err := response.GetContent()
	if err == nil {
		return content, nil
	}
	return nil, err
}

//Fetch gets content directly. Obsolete...
func (d directConn) Fetch(url string) ([]byte, error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // disable verify
	}

	timeout := time.Duration(15) * time.Second
	client := &http.Client{Transport: transCfg,
		Timeout: timeout}
	response, err := client.Get(url)
	// Check Error
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	utfBody, err := charset.NewReader(response.Body, response.Header.Get("Content-Type"))
	//NewReader returns an io.Reader that converts the content to UTF-8. It calls DetermineEncoding to find out what r's encoding is. https://godoc.org/golang.org/x/net/html/charset
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(utfBody)
	return buf.Bytes(), nil
}
