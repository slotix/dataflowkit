package storage

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/redsync.v1"
	"github.com/spf13/viper"
)

// RedisConn represents a Redis Connection structure
type RedisConn struct {
	host     string
	password string
	db       int
	pool     *redis.Pool
	// If set, path to a socket file overrides hostname
	socketPath string
	redsync    *redsync.Redsync
}

// NewRedisConn creates RedisConn instance
//func NewRedisConn(cnf *config.Config, host, password, socketPath string, db int) RedisConn {
func NewRedisConn(host, password, socketPath string, db int) RedisConn {

	return RedisConn{
		//	config:     cnf,
		host:       host,
		db:         db,
		password:   password,
		socketPath: socketPath,
	}
}

// Returns / creates instance of Redis connection
func (b *RedisConn) open() redis.Conn {
	if b.pool == nil {
		b.pool = b.newPool()
	}
	if b.redsync == nil {
		var pools = []redsync.Pool{b.pool}
		b.redsync = redsync.New(pools)
	}
	return b.pool.Get()
}

// Returns a new pool of Redis connections
func (b *RedisConn) newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			var (
				c    redis.Conn
				err  error
				opts = make([]redis.DialOption, 0)
			)

			if b.password != "" {
				opts = append(opts, redis.DialPassword(b.password))
			}

			if b.socketPath != "" {
				c, err = redis.Dial("unix", b.socketPath, opts...)
			} else {
				c, err = redis.Dial("tcp", b.host, opts...)
			}

			if b.db != 0 {
				_, err = c.Do("SELECT", b.db)
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

func (b *RedisConn) SetExpireAt(key string, expiresAt int64) error {
	var expirationTimestamp int32
	if expiresAt == 0 {
		// expire results after 1 hour by default
		expiresAt = viper.GetInt64("REDIS_EXPIRE")
		expirationTimestamp = int32(time.Now().UTC().Unix()+ expiresAt)
	} else {
		expirationTimestamp = int32(expiresAt)
	}
	conn := b.open()
	defer conn.Close()

	_, err := conn.Do("EXPIREAT", key, expirationTimestamp)
	if err != nil {
		return err
	}
	return nil
}

func (b *RedisConn) SetExpireIn(key string, expiresIn int64) error {
	if expiresIn == 0 {
		// expire results after 1 hour by default
		expiresIn = viper.GetInt64("REDIS_EXPIRE")
	}
	conn := b.open()
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, expiresIn)
	if err != nil {
		return err
	}
	return nil
}

//GetValue gets value from Redis
func (b *RedisConn) GetValue(key string) ([]byte, error) {
	//Get a key
	conn := b.open()
	defer conn.Close()
	content, err := redis.Bytes(conn.Do("GET", key))
	if err == nil {
		return content, nil
	}
	return nil, err
}

//GetIntValue gets value from Redis
func (b *RedisConn) GetIntValue(key string) (int64, error) {
	//Get a key
	conn := b.open()
	defer conn.Close()
	int, err := redis.Int64(conn.Do("GET", key))
	if err == nil {
		return int, nil
	}
	return 0, err
}

//SetValue pushes value to Redis
func (b *RedisConn) SetValue(key string, value interface{}) error {
	conn := b.open()
	defer conn.Close()
	reply, err := conn.Do("SET", key, value)
	if err != nil {
		return err
	}
	if reply.(string) == "OK" {
		return nil
	}
	return err

}