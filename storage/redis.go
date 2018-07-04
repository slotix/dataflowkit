package storage

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"gopkg.in/redsync.v1"
)

// RedisConn represents a Redis Connection structure
type RedisConn struct {
	host     string
	network  string
	password string
	db       int
	// If set, path to a socket file overrides hostname
	socketPath string
	pool       *redis.Pool
	redsync    *redsync.Redsync
}

// NewRedisConn initializes parameters for new Redis Connection
func NewRedisConn() RedisConn {
	return RedisConn{
		host:       viper.GetString("REDIS"),
		network:    viper.GetString("REDIS_NETWORK"),
		password:   viper.GetString("REDIS_PASSWORD"),
		db:         viper.GetInt("REDIS_DB"),
		socketPath: viper.GetString("REDIS_SOCKET_PATH"),
	}
}

// Returns / creates instance of Redis connection
func (rc *RedisConn) open() redis.Conn {
	if rc.pool == nil {
		rc.pool = rc.newPool()
	}
	if rc.redsync == nil {
		var pools = []redsync.Pool{rc.pool}
		rc.redsync = redsync.New(pools)
	}
	return rc.pool.Get()
}

// Returns a new pool of Redis connections
func (rc *RedisConn) newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			var (
				c    redis.Conn
				err  error
				opts = make([]redis.DialOption, 0)
			)

			if rc.password != "" {
				opts = append(opts, redis.DialPassword(rc.password))
			}

			if rc.socketPath != "" {
				c, err = redis.Dial("unix", rc.socketPath, opts...)
			} else {
				c, err = redis.Dial(rc.network, rc.host, opts...)
			}

			if rc.db != 0 {
				_, err = c.Do("SELECT", rc.db)
			}

			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func val(conn redis.Conn, key string) (value []byte, err error) {
	value, err = redis.Bytes(conn.Do("GET", key))
	return
}

// Read retrieves value from the specified key.
func (c RedisConn) Read(key string, rectType string) (value []byte, err error) {
	//Get a key
	conn := c.open()
	defer conn.Close()
	value, err = val(conn, key)
	return
}

func intVal(conn redis.Conn, key string) (value int64, err error) {
	value, err = redis.Int64(conn.Do("GET", key))
	return
}

//IntValue returns int64 value of specified key
func (b *RedisConn) ReadInt(key string) (value int64, err error) {
	conn := b.open()
	defer conn.Close()
	value, err = intVal(conn, key)
	return

}

func setVal(conn redis.Conn, key string, value interface{}) error {
	_, err := conn.Do("SET", key, value)
	//	if err != nil {
	//		return err
	//	}
	//	if reply.(string) == "OK" {
	//		return nil
	//	}
	return err
}

//SetValue saves key/ value pair to Redis
func (b *RedisConn) SetValue(key string, value interface{}) error {
	conn := b.open()
	defer conn.Close()
	err := setVal(conn, key, value)
	return err
}

// Write saves key/ value pair along with Expiration time to Redis storage.
func (s RedisConn) Write(key string, rec *Record, expTime int64) error {
	err := s.SetValue(key, rec.Value)
	if err != nil {
		return err
	}
	err = s.SetExpireAt(key, expTime)
	if err != nil {
		return err
	}
	return nil
}

func setExpireAt(conn redis.Conn, key string, expiresAt int64) error {
	_, err := conn.Do("EXPIREAT", key, expiresAt)
	return err
}

//ExpireAt sets TTL value of the specified key to expiresAt time
func (b *RedisConn) SetExpireAt(key string, expiresAt int64) error {
	conn := b.open()
	defer conn.Close()
	err := setExpireAt(conn, key, expiresAt)
	return err
}

func setExpireIn(conn redis.Conn, key string, expireIn int64) error {
	_, err := conn.Do("EXPIRE", key, expireIn)
	return err
}

//ExpireIn sets TTL value of the key to current Time + expireIn seconds
func (b *RedisConn) ExpireIn(key string, expireIn int64) error {
	conn := b.open()
	defer conn.Close()
	err := setExpireIn(conn, key, expireIn)
	return err
}

func expired(conn redis.Conn, key string) bool {
	t, err := ttl(conn, key)
	if err != nil {
		logger.Error(err)
		return true
	}
	return t < 0
}

// Expired returns either specified key is expired or not.
func (s RedisConn) Expired(key string) bool {
	conn := s.open()
	defer conn.Close()
	return expired(conn, key)
}

func ttl(conn redis.Conn, key string) (int64, error) {
	value, err := conn.Do("TTL", key)
	return value.(int64), err
}

//TTL returns Time to live value in seconds for the specified key
func (b *RedisConn) TTL(key string) (value int64, err error) {
	conn := b.open()
	defer conn.Close()
	value, err = ttl(conn, key)
	return
}

func setTTL(conn redis.Conn, key string, ttl int) error {
	_, err := conn.Do("EXPIRE", key, ttl)
	return err
}

//SetTTL sets TTL value in seconds of the specified key
func (b *RedisConn) SetTTL(key string, ttl int) error {
	conn := b.open()
	defer conn.Close()
	err := setTTL(conn, key, ttl)
	return err
}

func deleteKey(conn redis.Conn, key string) error {
	_, err := conn.Do("DEL", key)
	return err
}

//Delete deletes an object from Redis storage with specified key
func (b RedisConn) Delete(key string) error {
	conn := b.open()
	defer conn.Close()
	err := deleteKey(conn, key)
	return err
}

func deleteAllKeys(conn redis.Conn) error {
	_, err := conn.Do("FLUSHDB")
	return err
}

//DeleteAll deletes all objects from Redis storage
func (b RedisConn) DeleteAll() error {
	conn := b.open()
	defer conn.Close()
	err := deleteAllKeys(conn)
	return err
}

// Close storage connection
func (rc RedisConn) Close() {
	rc.Close()
}
