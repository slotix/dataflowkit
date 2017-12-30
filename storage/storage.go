package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/sirupsen/logrus"
	"github.com/slotix/dataflowkit/log"
	"github.com/spf13/viper"
)

var logger *logrus.Logger

func init() {
	logger = log.NewLogger()
}

//Store is the key interface of storage. All other structs implement methods wchich satisfy that interface.
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

//Type represent available storage types
type Type string

const (
	//Amazon S3 storage
	S3 Type = "S3"
	//Digital Ocean Spaces
	Spaces = "Spaces"
	//diskv key/value storage "github.com/peterbourgon/diskv"
	Diskv = "Diskv"
	//Redis
	Redis = "Redis"
)

// ParseType takes a string representing storage type and returns the Storage Type constant.
func ParseType(t string) (Type, error) {
	switch strings.ToLower(t) {
	case "s3":
		return S3, nil
	case "spaces":
		return Spaces, nil
	case "diskv":
		return Diskv, nil
	case "redis":
		return Redis, nil
	}
	var tp Type
	return tp, fmt.Errorf("not a valid Storage Type: %q", tp)
}

// NewStore creates New initialized Store instance with predefined parameters
func NewStore(t Type) Store {
	switch t {
	case Diskv:
		baseDir := viper.GetString("DISKV_BASE_DIR")
		return newDiskvStorage(baseDir, 1024*1024)
	case S3: //AWS S3
		bucket := viper.GetString("DFK_BUCKET")
		config := &aws.Config{
			Region: aws.String("us-east-1"),
		}
		return newS3Storage(config, bucket)

	case Spaces: //Digital Ocean Spaces
		bucket := viper.GetString("DFK_BUCKET")
		config := &aws.Config{
			Credentials: credentials.NewSharedCredentials(viper.GetString("SPACES_CONFIG"), ""), //Load credentials from specified file
			Endpoint:    aws.String(viper.GetString("SPACES_ENDPOINT")),                         //Endpoint is obligatory for DO Spaces
			Region:      aws.String("ams333"),                                                   //Actually for Digital Ocean spaces region parameter may have any value. But it can't be omited.
		}
		return newS3Storage(config, bucket)

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

// Read retrieves value according to the specified key from Redis.
func (s RedisConn) Read(key string) (value []byte, err error) {
	value, err = s.Value(key)
	return
}

// Write pushes key/ value pair along with Expiration time to Redis storage.
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

// Expired returns Expired value of specified key from Redis.
func (s RedisConn) Expired(key string) bool {
	ttl, err := s.TTL(key)
	if err != nil {
		fmt.Println(err)
	}
	return ttl > 0
	//if ttl > 0 {
	//	return false
	//}
	//return true

}

func newS3Storage(config *aws.Config, bucket string) Store {
	s3Conn := newS3Conn(config, bucket)
	return s3Conn
}

// Read retrieves value according to the specified key from AWS S3 Storage/ Digital Ocean Spaces.
func (s S3Conn) Read(key string) (value []byte, err error) {
	value, err = s.download(key)
	return
}

// Write uploads key/ value pair along with Expiration time to  AWS S3 Storage/ Digital Ocean Spaces.
func (s S3Conn) Write(key string, value []byte, expTime int64) error {
	err := s.upload(key, value, expTime)
	if err != nil {
		return err
	}
	return nil
}

// Expired returns Expired value of specified key from AWS S3 Storage/ Digital Ocean Spaces.
func (s S3Conn) Expired(key string) bool {
	obj, err := s.getObject(key)
	if err != nil {
		panic(err)
	}
	currentTime := time.Now().UTC()
	lastModified := obj.LastModified
	//calculate expiration time
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := lastModified.Add(exp)
	diff := expiry.Sub(currentTime)
	logger.Info("cache lifespan is %+v\n", diff)
	//Expired?
	return diff > 0
	//if diff > 0 {
	//	return false
	//}
	//return true
}

func newDiskvStorage(baseDir string, CacheSizeMax uint64) Store {
	d := newDiskvConn(baseDir, CacheSizeMax)
	return d
}

// Read loads value according to the specified key from DiskV KV storage.
func (d DiskvConn) Read(key string) (value []byte, err error) {
	value, err = d.diskv.Read(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Write stores key/ value pair along with Expiration time to DiskV KV storage.
func (d DiskvConn) Write(key string, value []byte, expTime int64) error {
	err := d.diskv.Write(key, value)
	if err != nil {
		return err
	}
	return nil
}

// Expired returns Expired value of specified key from DiskV.
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
	diff := expiry.Sub(currentTime)
	logger.Info("cache lifespan is %+v\n", diff)
	//Expired?
	return diff > 0
	// if diff > 0 {
	// 	return false
	// }
	// return true
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
