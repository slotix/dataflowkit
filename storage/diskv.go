package storage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/peterbourgon/diskv"
	"github.com/spf13/viper"
)

//DiskvConn stores connection parameters for Diskv storage
type DiskvConn struct {
	diskv *diskv.Diskv
}

//newDiskvConn creates new connection to Diskv storage initialized with Base directory and Cache Maximum Size parameters
func newDiskvConn(baseDir string, cacheSizeMax uint64) DiskvConn {
	// Simplest transform function: put all the data files into the base dir.
	flatTransform := func(s string) []string { return []string{} }
	// Initialize a new diskv store, rooted at "my-data-dir", with a 1MB cache.
	d := diskv.New(diskv.Options{
		BasePath:     baseDir,
		Transform:    flatTransform,
		CacheSizeMax: cacheSizeMax,
	})

	return DiskvConn{diskv: d}
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
func (d DiskvConn) Expired(key string) bool {
	//pwd
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	//filename
	fullPath := exPath + "/" + d.diskv.BasePath + "/" + key
	//file last modification time
	mTime, err := mTime(fullPath)
	if err != nil {
		logger.Error(err)
	}
	currentTime := time.Now().UTC()
	//calculate expiration time
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := mTime.Add(exp)
	diff := expiry.Sub(currentTime)
	logger.Info("cache lifespan is %+v\n", diff)
	//Expired?
	return diff > 0
}

//Erase deletes specified key from Diskv storage.
func (d DiskvConn) Erase(key string) error {
	err := d.Erase(key)
	if err != nil {
		return err
	}
	return nil
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