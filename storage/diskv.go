package storage

import (
	"os"
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
func (d DiskvConn) Read(rec Record) (value []byte, err error) {
	value, err = d.diskv.Read(rec.Key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Write stores key/ value pair along with Expiration time to DiskV KV storage.
func (d DiskvConn) Write(rec Record) error {
	err := d.diskv.Write(rec.Key, rec.Value)
	if err != nil {
		return err
	}
	return nil
}

// Expired returns Expired value of specified key from DiskV.
func (d DiskvConn) Expired(rec Record) bool {
	//pwd
	//ex, err := os.Executable()
	//if err != nil {
	//		panic(err)
	//	}
	//	exPath := filepath.Dir(ex)
	//filename
	//	fullPath := exPath + "/" + d.diskv.BasePath + "/" + key
	fullPath := d.diskv.BasePath + "/" + rec.Key
	//file last modification time
	fStat, err := os.Stat(fullPath)
	if err != nil {
		logger.Error(err.Error())
		return true
	}
	mTime := fStat.ModTime().UTC()
	//mTime, err := mTime(fullPath)
	//if err != nil {
	//	logger.Error(err)
	//}
	currentTime := time.Now().UTC()
	//calculate expiration time
	exp := time.Duration(viper.GetInt64("ITEM_EXPIRE_IN")) * time.Second
	//exp := time.Duration(3600) * time.Second
	//logger.Info(exp)
	//logger.Info(viper.GetInt64("ITEM_EXPIRE_IN"))
	expiry := mTime.Add(exp)
	diff := expiry.Sub(currentTime)
	//logger.Infof("cache lifespan is %+v", diff)
	//Expired?
	return diff < 0
}

// IsExists checkes whether specified record exists
func (d DiskvConn) IsExists(key string) bool {
	return d.diskv.Has(key)
}

//Delete deletes specified key from Diskv storage.
func (d DiskvConn) Delete(rec Record) error {
	err := d.diskv.Erase(rec.Key)
	if err != nil {
		return err
	}
	return nil
}

//DeleteAll deletes everything from Diskv storage.
func (d DiskvConn) DeleteAll() error {
	err := d.diskv.EraseAll()
	if err != nil {
		return err
	}
	return nil
}

// Close storage connection
func (d DiskvConn) Close() {
}
