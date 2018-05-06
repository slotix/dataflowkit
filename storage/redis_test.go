package storage

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	viper.Set("REDIS", "127.0.0.1:6379")
	viper.Set("REDIS_NETWORK", "tcp")
	viper.Set("REDIS_PASSWORD", "")
	viper.Set("REDIS_DB", "Redis DB")
	viper.Set("REDIS_SOCKET_PATH", "")
	//viper.Set("REDIS_EXPIRE", 3600)

	redisCon := NewRedisConn()
	conn := redisCon.open()
	//redisCon.conn = &conn
	defer conn.Close()

	//Delete all values from redis if any
	err := redisCon.DeleteAll()

	// Optionally set some keys your code expects:
	testKey := "some key"
	testValue := []byte("some value")

	//Write some values to Redis
	err = redisCon.SetValue(testKey, testValue)
	assert.Nil(t, err, "Expected no error")
	err = redisCon.SetValue("numeric key", 18)
	assert.Nil(t, err, "Expected no error")
	
	//Write
	//expAt := time.Now().UTC().Add(time.Duration(24 * time.Hour))
	err = redisCon.Write("One more key", []byte("111"), 3600)
	assert.Nil(t, err, "Expected no error")

	//Read from Redis
	value, err := redisCon.Read(testKey)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, value, testValue, "Expected value = testValue")
	numeric, err := redisCon.ReadInt("numeric key")
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, numeric, int64(18), "Expected numeric = 18")

	//Expirations
	expIn := int64(50)
	ttl, err := redisCon.TTL(testKey)
	t.Log(ttl, err)
	err = redisCon.ExpireIn(testKey, expIn)
	assert.Nil(t, err, "Expected no error")
	ttl, err = redisCon.TTL(testKey)
	t.Log(ttl, err)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, true, ttl > 0, "Expected TTL >0")

	//EXPIREAT
	expAt := time.Now().UTC().Add(time.Duration(24 * time.Hour))
	t.Log(expAt)
	err = redisCon.SetExpireAt(testKey, expAt.Unix())
	assert.Nil(t, err, "Expected no error")
	ttl, err = redisCon.TTL(testKey)
	t.Log(ttl, err)
	assert.Nil(t, err, "Expected no error")
	assert.Equal(t, true, ttl > 0, "Expected TTL >0")

	exp := redisCon.Expired(testKey)
	//logger.Info(TTL)
	assert.Equal(t, exp, false, "Expected Not expired")

	err = redisCon.SetTTL(testKey, -1)
	exp = redisCon.Expired(testKey)
	assert.Equal(t, exp, true, "Expected expired value")

	err = redisCon.Delete("numeric key")
	assert.Nil(t, err, "Expected no error")
	err = redisCon.DeleteAll()
	assert.Nil(t, err, "Expected no error")
	//err := redisCon.Write("test", []byte("fff"), 0)
	//t.Log(err)
}
