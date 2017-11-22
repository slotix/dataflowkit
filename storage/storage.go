package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Store interface {
	//Reads value from storage by specified key
	Read(key string) (value []byte, err error)
	//Writes specified pair key value to storage.
	//expTime value sets TTL for Redis storage.
	//expTime set Metadata Expires value for S3Storage
	Write(key string, value []byte, expTime int64) error
	//Is key expired ? It checks if parse results storage item is expired. Set up  Expiration fixed value as "STORAGE_EXPIRE" environment variable.
	//html pages cache stores this info in sResponse.Expires . It is not used for fetch endpoint.
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

func (s S3Conn) Write(key string, value []byte, expTime int64) error {
	err := s.Upload(key, value, expTime)
	if err != nil {
		return err
	}
	return nil
}

func (s S3Conn) Expired(key string) bool {
	obj, err := s.GetObject(key)
	if err != nil {
		panic(err)
	}
	currentTime := time.Now().UTC()
	lastModified := obj.LastModified
	//calculate expiration time
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := lastModified.Add(exp)
	diff := expiry.Sub(currentTime)
	logger.Printf("cache lifespan is %+v\n", diff)
	//Expired?
	if diff > 0 {
		return false
	}
	return true
}

func newDiskvStorage(baseDir string, CacheSizeMax uint64) Store {
	d := newDiskvConn(baseDir, CacheSizeMax)
	return d
}

func (d DiskvConn) Read(key string) (value []byte, err error) {
	value, err = d.diskv.Read(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (d DiskvConn) Write(key string, value []byte, expTime int64) error {
	err := d.diskv.Write(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (s DiskvConn) Expired(key string) bool {
	//pwd
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	//filename
	fullPath := exPath + "/" + s.diskv.BasePath + "/" + key
	//file last modification time
	mTime, err := mTime(fullPath)
	currentTime := time.Now().UTC()
	//calculate expiration time
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := mTime.Add(exp)
	//logger.Println(mTime)
	//logger.Println(currentTime)
	//logger.Println(expiry)
	diff := expiry.Sub(currentTime)
	logger.Printf("cache lifespan is %+v\n", diff)
	//Expired?
	if diff > 0 {
		return false
	}
	return true
}

//mTime returns File Modify Time
//Last modification time shows time of the  last change to file's contents. It does not change with owner or permission changes, and is therefore used for tracking the actual changes to data of the file itself.
func mTime(name string) (mtime time.Time, err error) {
	fi, err := os.Stat(name)
	if err != nil {
		return
	}
	mtime = fi.ModTime().UTC()
	return
}
