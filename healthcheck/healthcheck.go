package healthcheck

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/slotix/dataflowkit/splash"
)

type Checker interface {
	isAlive() error
	serviceName() string
}

type RedisConn struct {
	Conn    redis.Conn
	Network string
	Host    string
}

type SplashConn struct {
	Host string
	//	User            string
	//	Password        string
}

type FetchConn struct {
	Host string
}

type ParseConn struct {
	Host string
}

func (FetchConn) serviceName() string {
	return "DFK Fetch Service"
}

func (ParseConn) serviceName() string {
	return "DFK Parse Service"
}

func (RedisConn) serviceName() string {
	return "Redis"
}

func (SplashConn) serviceName() string {
	return "Splash"
}

func (p ParseConn) isAlive() error {
	//reader := bytes.NewReader(b)
	addr := "http://" + p.Host + "/ping"
	request, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(request)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		panic(err)
	}
	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if string(resp) != `{"alive": true}` {
		return errors.New("Parse Service is dead")
	}
	return nil
}

func (f FetchConn) isAlive() error {
	//reader := bytes.NewReader(b)
	addr := "http://" + f.Host + "/ping"
	request, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	r, err := client.Do(request)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		panic(err)
	}
	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if string(resp) != `{"alive": true}` {
		return errors.New("Parse Service is dead")
	}
	return nil
}

func (r RedisConn) isAlive() error {
	var err error
	r.Conn, err = redis.Dial(r.Network, r.Host)
	if err != nil {
		return err
	}
	defer r.Conn.Close()
	res, err := r.Conn.Do("PING")
	if err != nil {
		return err
	}
	if res == "PONG" {
		return nil
	}
	return err
}

func (s SplashConn) isAlive() error {
	resp, err := splash.Ping(s.Host)
	if err != nil {
		return err
	}
	if resp.Status == "ok" {
		return nil
	}
	return err
}

func CheckServices(hc ...Checker) (status map[string]string) {
	status = make(map[string]string)
	for _, srv := range hc {
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
