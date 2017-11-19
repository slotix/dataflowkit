package storage
import (
	"github.com/peterbourgon/diskv"
)

type DiskvConn struct {
	diskv *diskv.Diskv
}

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

func (d DiskvConn) Erase(key string) error {
	err := d.Erase(key)
	if err != nil {
		return err
	}
	return nil
}
