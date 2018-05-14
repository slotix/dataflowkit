package healthcheck

import (
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
	assert.Equal(t, map[string]string{"DFK Parse Service":"Ok", "DFK Fetch Service":"Ok", "Redis":"Ok", "Splash":"Ok"}, status)

	checkers = []Checker{
		ParseConn{Host: invalidhost},
		FetchConn{Host: invalidhost},
		RedisConn{
			Network: "tcp",
			Host:    invalidhost,
		},
		SplashConn{Host: invalidhost},
	}
	status1 := CheckServices(checkers...)
	assert.Equal(t, map[string]string{"DFK Parse Service": "Get http://invalidhost/ping: dial tcp: lookup invalidhost: no such host", "DFK Fetch Service": "Get http://invalidhost/ping: dial tcp: lookup invalidhost: no such host", "Redis": "dial tcp: address invalidhost: missing port in address", "Splash": "Get http://invalidhost/_ping: dial tcp: lookup invalidhost: no such host"}, status1)
}
