package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

type healthChecker interface {
	isAlive() error
	serviceName() string
}

type redisConn struct {
	conn     redis.Conn
	protocol string
	addr     string
}

type splashConn struct {
	url      string
	userName string
	userPass string
}

func (r redisConn) serviceName() string {
	return "Redis"
}

func (s splashConn) serviceName() string {
	return "Splash"
}

func (r redisConn) isAlive() error {
	var err error
	r.conn, err = redis.Dial(r.protocol, r.addr)
	if err != nil {
		return err
	}
	defer r.conn.Close()
	res, err := r.conn.Do("PING")
	if err != nil {
		return err
	}
	if res == "PONG" {
		return nil
	}
	return err
}

func (s splashConn) isAlive() error {

	var p pingSplashServerResponse
	err := p.pingSplashServer(s.url)
	if err != nil {
		return err
	}
	if p.Status == "ok" {
		return nil
	}
	return err
}

type pingSplashServerResponse struct {
	Maxrss int    `json:"maxrss"`
	Status string `json:"status"`
}

//pingSplashServer returns status and maxrss from Splash server
//http://localhost:8050/_ping  endpoint
func (p *pingSplashServerResponse) pingSplashServer(url string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	//req.SetBasicAuth(userName, userPass)
	resp, err := client.Do(req)
	// Check Error
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &p)
	if err != nil {
		return err
	}
	return nil
}

func CheckServices() (status map[string]string) {
	status = make(map[string]string)
	services := []healthChecker{
		redisConn{
			protocol: viper.GetString("redis.protocol"),
			addr:     viper.GetString("redis.address")},
		splashConn{
			url:      fmt.Sprintf("%s%s", viper.GetString("splash.base-url"), viper.GetString("splash.ping-url")),
			userName: viper.GetString("splash.username"),
			userPass: viper.GetString("splash.userpass")}}

	for _, srv := range services {
		err := srv.isAlive()
		if err != nil {
			status[srv.serviceName()] =
				fmt.Sprintf("%s: %s", srv.serviceName(), err.Error())
		} else {
			status[srv.serviceName()] = "Ok"
		}
	}
	return status
}
