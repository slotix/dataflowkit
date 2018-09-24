package storage

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

//Storage types
const (
	CACHE        = "Cache"
	COOKIES      = "Cookies"
	INTERMEDIATE = "Intermediate"
)

// Record struct keeps Key/Value and expiration time of specified type
type Record struct {
	Type    string
	Key     string
	Value   []byte
	ExpTime int64
}

//Store is the key interface of storage. All other structs implement methods wchich satisfy that interface.
type Store interface {
	//Reads value from storage by specified key
	Read(rec Record) (value []byte, err error)
	//Writes specified pair key value to storage.
	//expTime value sets TTL for Redis storage.
	//expTime set Metadata Expires value for S3Storage
	Write(rec Record) error
	// IsExists checkes whether specified by key record exists
	IsExists(key string) bool
	//Is key expired ? It checks if parse results storage item is expired. Set up  Expiration as "ITEM_EXPIRE_IN" environment variable.
	//html pages cache stores this info in sResponse.Expires . It is not used for fetch endpoint.
	Expired(rec Record) bool
	//Delete deletes specified item from the store
	Delete(rec Record) error
	//DeleteAll erases all items from the store
	DeleteAll() error
	// Close storage connection
	Close()
}

// NewStore creates New initialized Store instance with predefined parameters
// Storage Types: S3, Spaces, Redis, Diskv, Cassandra
func NewStore(sType string) Store {
	switch strings.ToLower(sType) {
	case "diskv":
		baseDir := viper.GetString("DISKV_BASE_DIR")
		//return newDiskvStorage(baseDir, 1024*1024)
		var cacheSizeMax uint64
		cacheSizeMax = 1024 * 1024
		return newDiskvConn(baseDir, cacheSizeMax)
	case "cassandra":
		cassandraHost := viper.GetString("CASSANDRA")
		return newCassandra(cassandraHost)
	default:
		panic(errors.New("no storage type specified"))
		// case "s3": //AWS S3
		// 	bucket := viper.GetString("DFK_BUCKET")
		// 	config := &aws.Config{
		// 		Region: aws.String(viper.GetString("S3_REGION")),
		// 	}
		// 	//return newS3Storage(config, bucket)
		// 	return newS3Conn(config, bucket)

		// case "spaces": //Digital Ocean Spaces
		// 	bucket := viper.GetString("DFK_BUCKET")
		// 	config := &aws.Config{
		// 		Credentials: credentials.NewSharedCredentials(viper.GetString("SPACES_CONFIG"), ""), //Load credentials from specified file
		// 		Endpoint:    aws.String(viper.GetString("SPACES_ENDPOINT")),                         //Endpoint is obligatory for DO Spaces
		// 		Region:      aws.String(viper.GetString("S3_REGION")),
		// 		//Region:      aws.String("ams333"),                                                   //Actually for Digital Ocean spaces region parameter may have any value. But it can't be omitted.
		// 	}
		// 	return newS3Conn(config, bucket)
		// 	// return newS3Storage(config, bucket)

		// case "redis":
		// 	host := viper.GetString("REDIS")
		// 	network := viper.GetString("REDIS_NETWORK")
		// 	password := viper.GetString("REDIS_PASSWORD")
		// 	db := viper.GetInt("REDIS_DB")
		// 	return NewRedisConn(host, network, password, db)
	}
}
