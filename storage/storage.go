package storage

import (
	"encoding/base32"
	"fmt"

	"github.com/spf13/viper"
)

type Store interface {
	//Reads value from storage by specified key
	Read(key string) (value []byte, err error)
	//Writes specified pair key value to storage.
	//expTime value sets TTL for Redis storage.
	//expTime set Metadata Expires value for S3Storage
	Write(key string, value []byte, expTime int64) error
	//Is key expired ?
	Expired(key string) bool
}

type Type string

const (
	S3    Type = "S3"
	Diskv      = "Diskv"
	Redis      = "Redis"
)

func NewStore(t Type) Store {
	switch t {
	case Diskv:
		baseDir := viper.GetString("DISKV_BASE_DIR")
		return newDiskvStorage(baseDir, 1024*1024)
	case S3:
		bucket := viper.GetString("FETCH_BUCKET")
		return newS3Storage(bucket)
	case Redis:
		redisHost := viper.GetString("REDIS")
		redisPassword := ""
		return newRedisStorage(redisHost, redisPassword)
	default:
		return nil
	}
}

func newRedisStorage(redisHost, redisPassword string) Store {
	redisCon := NewRedisConn()
	return redisCon
}

func (s RedisConn) Read(key string) (value []byte, err error) {
	value, err = s.Value(key)
	return
}

func (s RedisConn) Write(key string, value []byte, expTime int64) error {
	err := s.SetValue(key, value)
	if err != nil {
		return err
	}
	err = s.ExpireAt(key, expTime)
	if err != nil {
		return err
	}
	return nil
}

func (s RedisConn) Expired(key string) bool {
	ttl, err := s.TTL(key)
	if err != nil {
		fmt.Println(err)
	}
	if ttl > 0 {
		return false
	}
	return true

}

func newS3Storage(bucket string) Store {
	s3Conn := newS3Conn(bucket)
	return s3Conn
}

func (s S3Conn) Read(key string) (value []byte, err error) {
	value, err = s.Download(key)
	return
}

//TODO: implement expiration functionality
func (s S3Conn) Write(key string, value []byte, expTime int64) error {
	err := s.Upload(key, value, expTime)
	if err != nil {
		return err
	}
	return nil
}

func (s S3Conn) Expired(key string) bool {
	return false
}

func newDiskvStorage(baseDir string, CacheSizeMax uint64) Store {
	d := newDiskvConn(baseDir, CacheSizeMax)
	return d
}

func (d DiskvConn) Read(key string) (value []byte, err error) {

	//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
	sKey := base32.StdEncoding.EncodeToString([]byte(key))
	value, err = d.diskv.Read(sKey)
	if err != nil {
		return nil, err
	}
	return value, nil
}

//TODO: implement expiration functionality
func (d DiskvConn) Write(key string, value []byte, expTime int64) error {
	//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
	sKey := base32.StdEncoding.EncodeToString([]byte(key))
	err := d.diskv.Write(sKey, value)
	if err != nil {
		return err
	}
	return nil
}

func (s DiskvConn) Expired(key string) bool {
	return false
}
