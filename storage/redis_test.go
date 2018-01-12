package storage

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func init() {
}

func TestRedis(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer server.Close()
	// Optionally set some keys your code expects:
	testKey := "some key"
	testValue := []byte("some value")
	// Run your code and see if it behaves.
	// An example using the redigo library from "github.com/garyburd/redigo/redis":
	conn, err := redis.Dial("tcp", server.Addr())

	//Write some values
	err = setVal(conn, testKey, testValue)
	assert.Nil(t, err, "Expected no error")
	err = setVal(conn, "numeric key", 18)
	assert.Nil(t, err, "Expected no error")

	//Read from Redis
	value, err := val(conn, testKey)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, value, testValue, "Expected value = testValue")
	numeric, err := intVal(conn, "numeric key")
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, numeric, int64(18), "Expected numeric = 18")

	expAt := time.Now().UTC().Add(time.Duration(5 * time.Second))
	err = setExpireAt(conn, testKey, expAt.Unix())
	assert.Nil(t, err, "Expected no error")
	TTL, err := ttl(conn, testKey)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, TTL > 0, true, "Expected TTL >0")

	expIn := int64(5)
	err = setExpireIn(conn, testKey, expIn)
	assert.Nil(t, err, "Expected no error")
	TTL, err = ttl(conn, testKey)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, TTL > 0, true, "Expected TTL >0")
	exp := expired(conn, testKey)
	//logger.Info(TTL)
	assert.Equal(t, exp, false, "Expected Not expired")
	
	err = setTTL(conn, testKey, -1)
	exp = expired(conn, testKey)
	assert.Equal(t, exp, true, "Expected expired value")
}

func TestNewRedisConn(t *testing.T) {
	rc := NewRedisConn()
	t.Log(rc)
	conn := rc.open()
	t.Log(conn)
	defer conn.Close()
	// err := rc.Write("test",[]byte("fff"),0)
	// t.Log(err)
}