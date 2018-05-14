// Dataflow kit - healthcheck
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package healthcheck of the Dataflow kit checks if specified services are alive.
//
package healthcheck

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/slotix/dataflowkit/splash"
)

//Checker is the key interface of healthChecker. All other structs implement methods wchich satisfy that interface.
type Checker interface {
	//if server is alive
	isAlive() error
	//String returns service name
	String() string
}

// RedisConn struct implements methods satisfying Checker interface
type RedisConn struct {
	Conn    redis.Conn
	Network string
	Host    string
}

// SplashConn struct implements methods satisfying Checker interface
type SplashConn struct {
	Host string
	//	User            string
	//	Password        string
}

// FetchConn struct implements methods satisfying Checker interface
type FetchConn struct {
	Host string
}

// ParseConn struct implements methods satisfying Checker interface
type ParseConn struct {
	Host string
}

func (FetchConn) String() string {
	return "DFK Fetch Service"
}

func (ParseConn) String() string {
	return "DFK Parse Service"
}

func (RedisConn) String() string {
	return "Redis"
}

func (SplashConn) String() string {
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
		return err
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
		return err
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

// CheckServices checks if services passed as parameters are alive. It returns the map containing serviceName and appropriate status
func CheckServices(hc ...Checker) (status map[string]string) {
	status = make(map[string]string)
	for _, srv := range hc {
		err := srv.isAlive()
		if err != nil {
			status[srv.String()] =
				fmt.Sprintf("%s", err.Error())
		} else {
			status[srv.String()] = "Ok"
		}
	}
	return status
}
