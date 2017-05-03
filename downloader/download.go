package downloader

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"time"

	"golang.org/x/net/html/charset"
)

type directConn struct {
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
