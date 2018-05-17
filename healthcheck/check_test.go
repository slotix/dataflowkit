package healthcheck

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler(t *testing.T) {
	host := "127.0.0.1:12345"
	invalidhost := "invalidhost"

	checkers := []Checker{
		ParseConn{Host: host},
		FetchConn{Host: host},
		RedisConn{
			Network: "tcp",
			Host:    "127.0.0.1:6379",
		},
		SplashConn{Host: "127.0.0.1:8050"},
	}
	status := CheckServices(checkers...)
	eq := reflect.DeepEqual(map[string]string{"DFK Parse Service": "Ok", "DFK Fetch Service": "Ok", "Redis": "Ok", "Splash": "Ok"}, status)
	assert.Equal(t, eq, true)

	checkers = []Checker{
		ParseConn{Host: invalidhost},
		FetchConn{Host: invalidhost},
		RedisConn{
			Network: "tcp",
			Host:    invalidhost + ":12345",
		},
		SplashConn{Host: invalidhost},
	}
	status1 := CheckServices(checkers...)
	for _, v := range status1{
		assert.NotEqual(t, "Ok", v)
	}
}
