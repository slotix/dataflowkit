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
		ChromeConn{Host: "http://localhost:9222"},
		CassandraConn{Host: "127.0.0.1"},
	}
	status := CheckServices(checkers...)
	t.Log(status)
	eq := reflect.DeepEqual(map[string]string{"DFK Parse Service": "Ok", "DFK Fetch Service": "Ok", "Headless Chrome": "Ok", "Cassandra":"Ok"}, status)
	assert.Equal(t, eq, true)

	checkers = []Checker{
		ParseConn{Host: invalidhost},
		FetchConn{Host: invalidhost},
		ChromeConn{Host: invalidhost},
	}
	status1 := CheckServices(checkers...)
	for _, v := range status1 {
		assert.NotEqual(t, "Ok", v)
	}
}
