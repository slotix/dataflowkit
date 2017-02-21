package parser

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"golang.org/x/net/html/charset"
)

//TODO:
//if error don't put to REDIS !!!
/*{
  "type": "GlobalTimeoutError",
   "info":  {
    "timeout":  10.0
  },
   "error":  504,
   "description":  "Timeout exceeded rendering page"
}
errors:
  errNoSelectors: No selectors found
  errURLEmpty: Empty string. URL
  errRedisConn: Cannot connect to Redis
  errWriteRow: Cannot write Excel Row

*/

// ErrURLEmpty is returned when an input string is empty.
var errURLEmpty = errors.New("Empty string. URL")
var ErrRedisDown = errors.New("Redis. Cannot connect")
var ErrRedisSave = errors.New("Redis. Save failed")

type HTMLGetter interface {
	getHTML(url string) ([]byte, error)
}

type redisConn struct {
	conn     redis.Conn
	protocol string
	addr     string
}

type splashConn struct {
	timeout         int
	resourceTimeout int
	wait            int
	userName        string
	userPass        string
}

type directConn struct {
}

func (r redisConn) getHTML(url string) ([]byte, error) {
	//Get a key
	content, err := redis.Bytes(r.conn.Do("GET", url))
	if err == nil {
		return content, nil
	}
	return nil, err
}

//SetHTML pushes html buffer to Redis
func (r redisConn) setHTML(url string, buf []byte) error {
	reply, err := r.conn.Do("SET", url, buf)
	if err != nil {
		return err
	}
	if reply.(string) == "OK" {
		//set 1 hour 3600 before html content key expiration
		r.conn.Do("EXPIRE", url, viper.GetInt("redis.expire"))
	}
	return nil
}

func (s splashConn) getHTML(url string) ([]byte, error) {
	client := &http.Client{}
	splashURL := fmt.Sprintf("%s%s?url=%s&timeout=%d&resource_timeout=%d&wait=%d", viper.GetString("splash.base-url"), viper.GetString("splash.render-html"), url, s.timeout, s.resourceTimeout, s.wait)
	req, err := http.NewRequest("GET", splashURL, nil)
	req.SetBasicAuth(s.userName, s.userPass)
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

//GetHTML gets content from url.
//At first it checks if there is local copy of content in Redis cache
//Secondly it pulls data from Splash server https://github.com/scrapinghub/splash/ running localy.
//Then it pushes content to Redis to make it available localy for 3600 seconds
func GetHTML(url string) ([]byte, error) {
	if url == "" {
		return nil, errURLEmpty
	}
	rc := redisConn{
		protocol: viper.GetString("redis.protocol"),
		addr:     viper.GetString("redis.address")}
	var err error
	rc.conn, err = redis.Dial(rc.protocol, rc.addr)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", ErrRedisDown, err.Error())
	}
	defer rc.conn.Close()
	//Get html content from Redis
	content, err := rc.getHTML(url)
	if err == nil {
		return content, nil
	}
	//if there is nothing in Redis get content from Splash server - javascript rendering service
	s := splashConn{
		timeout:         viper.GetInt("splash.timeout"),
		resourceTimeout: viper.GetInt("splash.resource-timeout"),
		wait:            viper.GetInt("splash.wait"), //wait parameter should be something more than default 0,5 value as it is not enough to load js scripts
		userName:        viper.GetString("splash.username"),
		userPass:        viper.GetString("splash.userpass")}
	content, err = s.getHTML(url)
	if err == nil {
		//push html content to redis
		err1 := rc.setHTML(url, content)
		if err1 != nil {
			fmt.Printf("%s: %s", ErrRedisSave, err1.Error())
		}
		return content, nil
	}
	return nil, err
}

//getHTML gets content directly before running javascripts. It is not used for the moment
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
