package cache

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/redsync.v1"
)

// RedisConn represents a Redis Connection structure
type RedisConn struct {
	//config   *config.Config
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

// Sets expiration timestamp on a stored state
func (b *RedisConn) setExpirationTime(key string) error {
	//expiresIn := b.config.ResultsExpireIn
	expiresIn := 0
	if expiresIn == 0 {
		// // expire results after 1 hour by default
		expiresIn = 3600
	}
	expirationTimestamp := int32(time.Now().Unix() + int64(expiresIn))

	conn := b.open()
	defer conn.Close()

	_, err := conn.Do("EXPIREAT", key, expirationTimestamp)
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
	//	logger.Println(key, err)
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
	//	logger.Println(key, err)
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
		//set 1 hour 3600 before html content key expiration
		//r.conn.Do("EXPIRE", url, viper.GetInt("redis.expire"))
		err = b.setExpirationTime(key)
		if err != nil {
			return err
		}
	}
	return nil
}

