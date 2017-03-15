package parser

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"time"

	"golang.org/x/net/html/charset"
)

// ErrURLEmpty is returned when an input string is empty.
var errURLEmpty = errors.New("Empty string. URL")
var errRedisSet = errors.New("Redis. Set Value failed")

type directConn struct {
}


//GetHTML gets content from url.
//It checks if there is local copy of content in Redis cache
//If no data is pulled through Splash server https://github.com/scrapinghub/splash/ .
//Then it pushes content to Redis to make it available localy for 3600 seconds by default
func GetHTML(url string) ([]byte, error) {
	if url == "" {
		return nil, errURLEmpty
	}

	redisURL := "localhost:6379"
	redisPassword := ""
	redis := NewRedisConn(redisURL, redisPassword, "", 0)
	content, err := redis.GetValue(url)
	if err == nil {
		return content, nil
	}
	s := NewSplashConn(
		"http://localhost:8050/",
		"render.html",
		"user",
		"userpass",
		20,
		30,
		1, //wait parameter should be something more than default 0,5 value as it is not enough to load js scripts
	)

	content, err = s.getHTML(url)
	if err == nil {

		err1 := redis.SetValue(url, content)
		if err1 != nil {
			fmt.Printf("%s: %s", errRedisSet, err1.Error())
		}
		return content, nil
	}
	return nil, err
}

//getHTML gets content directly. Obsolete...
func (d directConn) getHTML(url string) ([]byte, error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // disable verify
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
