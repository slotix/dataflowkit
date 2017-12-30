package storage

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"gopkg.in/redsync.v1"
)

// Options struct inclued parameters for Redis Connection 
type Options struct {
	host     string
	network  string
	password string
	db       int
	// If set, path to a socket file overrides hostname
	socketPath string
}

// Option represent parameters used for connection to Redis server 
type Option func(*Options)

func host(h string) Option {
	return func(args *Options) {
		args.host = h
	}
}

func network(n string) Option {
	return func(args *Options) {
		args.network = n
	}
}

func password(p string) Option {
	return func(args *Options) {
		args.password = p
	}
}

func db(d int) Option {
	return func(args *Options) {
		args.db = d
	}
}

func socketPath(s string) Option {
	return func(args *Options) {
		args.socketPath = s
	}
}


// RedisConn represents a Redis Connection structure
type RedisConn struct {
	//host     string
	//password string
	//db       int
	opts *Options
	pool *redis.Pool
	// If set, path to a socket file overrides hostname
	//socketPath string
	redsync *redsync.Redsync
}

// NewRedisConn initializes parameters for new Redis Connection
func NewRedisConn(setters ...Option) RedisConn {
	args := &Options{
		host:       viper.GetString("REDIS"),
		network:    viper.GetString("REDIS_NETWORK"),
		password:   viper.GetString("REDIS_PASSWORD"),
		db:         viper.GetInt("REDIS_DB"),
		socketPath: viper.GetString("REDIS_SOCKET_PATH"),
	}

	return RedisConn{
		opts: args,
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

			if b.opts.password != "" {
				opts = append(opts, redis.DialPassword(b.opts.password))
			}

			if b.opts.socketPath != "" {
				c, err = redis.Dial("unix", b.opts.socketPath, opts...)
			} else {
				c, err = redis.Dial(b.opts.network, b.opts.host, opts...)
			}

			if b.opts.db != 0 {
				_, err = c.Do("SELECT", b.opts.db)
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

//Value returns value of specified key
func (b *RedisConn) Value(key string) ([]byte, error) {
	//Get a key
	conn := b.open()
	defer conn.Close()
	content, err := redis.Bytes(conn.Do("GET", key))
	if err == nil {
		return content, nil
	}
	return nil, err
}

//IntValue returns int64 value of specified key 
func (b *RedisConn) IntValue(key string) (int64, error) {
	//Get a key
	conn := b.open()
	defer conn.Close()
	int, err := redis.Int64(conn.Do("GET", key))
	if err == nil {
		return int, nil
	}
	return 0, err
}

//SetValue saves key/ value pair to Redis
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

//ExpireAt sets TTL value of the specified key to expiresAt time
func (b *RedisConn) ExpireAt(key string, expiresAt int64) error {
	conn := b.open()
	defer conn.Close()
	_, err := conn.Do("EXPIREAT", key, expiresAt)
	if err != nil {
		return err
	}
	return nil
}

//ExpireIn sets TTL value of the key to current Time + expireIn seconds
func (b *RedisConn) ExpireIn(key string, expireIn int64) error {
	conn := b.open()
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, expireIn)
	if err != nil {
		return err
	}
	return nil
}

//TTL returns Time to live value in seconds for the specified key 
func (b *RedisConn) TTL(key string) (int64, error) {
	conn := b.open()
	defer conn.Close()
	reply, err := conn.Do("TTL", key)
	if err != nil {
		return 0, err
	}
	return reply.(int64), nil
}
