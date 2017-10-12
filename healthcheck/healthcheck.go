package healthcheck

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/slotix/dataflowkit/splash"
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
	Host            string
	User            string
	Password        string
	Timeout         int
	ResourceTimeout int
	LUAScript       string
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
	resp, err := splash.Ping(s.Host)
	if err != nil {
		return err
	}
	if resp.Status == "ok" {
		return nil
	}
	return err
}

func CheckServices() (status map[string]string) {
	status = make(map[string]string)
	services := []healthChecker{
		redisConn{
			network: viper.GetString("REDIS_NETWORK"),
			host:    viper.GetString("REDIS")},
		splashConn{
			//conn: splash.Connection{
				
				Host: viper.GetString("SPLASH"),
		},
	}
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
