package storage
import (
	"github.com/peterbourgon/diskv"
)

//DiskvConn stores connection parameters for Diskv storage
type DiskvConn struct {
	diskv *diskv.Diskv
}

//newDiskvConn creates new connection to Diskv storage initialized with Base directory and Cache Maximum Size parameters
func newDiskvConn(baseDir string, cacheSizeMax uint64) DiskvConn{
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

//Erase deletes specified key from Diskv storage.
func (d DiskvConn) Erase(key string) error {
	err := d.Erase(key)
	if err != nil {
		return err
	}
	return nil
}
