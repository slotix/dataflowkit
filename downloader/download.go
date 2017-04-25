package downloader

import (
	"bytes"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/net/html/charset"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "downloader: ", log.Lshortfile)
}

// ErrURLEmpty is returned when an input string is empty.
var errURLEmpty = errors.New("Empty string. URL")

type directConn struct {
}

func isRobotsTxt(url string) bool {
	if strings.Contains(url, "robots.txt") {
		return true
	}
	return false
}

func getLUA(url string) string {
	if isRobotsTxt(url) {
		return robotsLUA
	}
	return generalLUA
}

//GetResponse is needed to be passed to  httpcaching middleware
//to provide a RFC7234 compliant HTTP cache
func GetResponse(url string) (*SplashResponse, error) {
	if url == "" {
		return nil, errURLEmpty
	}
	wait := viper.GetInt("splash-wait")
	s := NewSplashConn(
		fmt.Sprintf("http://%s/", viper.GetString("splash")),
		viper.GetInt("splash-timeout"),
		viper.GetInt("splash-resource-timeout"),
		wait, //wait parameter should be something more than default 0,5 value as it is not enough to load js scripts
		fmt.Sprintf(getLUA(url), wait),
	)

	response, err := s.GetResponse(url)
	//logger.Println(response)
	return response, err
}

//Fetch gets content from url.
//If no data is pulled through Splash server https://github.com/scrapinghub/splash/ .
func Fetch(url string) ([]byte, error) {
	response, err := GetResponse(url)
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
