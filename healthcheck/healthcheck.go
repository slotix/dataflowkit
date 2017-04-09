package healthcheck

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
	conn    redis.Conn
	network string
	host    string
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
	r.conn, err = redis.Dial(r.network, r.host)
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
//_ping  endpoint
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
			network: viper.GetString("redis-network"),
			host:    viper.GetString("redis")},
		splashConn{
			//url: fmt.Sprintf("http://%s/_ping", viper.GetString("splash.host"))}}
			url: fmt.Sprintf("http://%s/_ping", viper.GetString("splash"))}}
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
