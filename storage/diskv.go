package storage
import (
	"github.com/peterbourgon/diskv"
	slug "github.com/slotix/slugifyurl"
)

type DiskvConn struct {
	diskv *diskv.Diskv
	//fetched pages can't be saved as-is. Keys are filenames representing urls have to be Slugied to a sanitized string which can be used as a filename.
	options slug.Options
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
	slugOptions := slug.Options{
		SlashChar:    "-",
		MaxLength:    50,
		SkipScheme:   true,
		SkipUserinfo: true,
		UnixOnly:     false,
	}
	return DiskvConn{diskv: d, options: slugOptions}
}

func (d DiskvConn) Erase(key string) error {
	err := d.Erase(key)
	if err != nil {
		return err
	}
	return nil
}
