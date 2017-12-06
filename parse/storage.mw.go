package parse

import (
	"bytes"
	"encoding/base32"
	"io"
	"io/ioutil"
	"time"

	"github.com/slotix/dataflowkit/scrape"
	"github.com/spf13/viper"

	"github.com/slotix/dataflowkit/storage"
)

//storageMiddleware caches Parsed results in storage.
type storageMiddleware struct {
	//storage instance puts fetching results to a cache
	storage storage.Store
	Service
}

// implement function to return ServiceMiddleware
func StorageMiddleware(storage storage.Store) ServiceMiddleware {
	return func(next Service) Service {
		return storageMiddleware{storage, next}
	}
}

//Parse returns parsed results either from storage or directly from scraped web pages.
func (mw storageMiddleware) Parse(p scrape.Payload) (output io.ReadCloser, err error) {
	//if parsed result is in storage return local copy
	//storageKey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	storageKey := string(p.PayloadMD5)
	//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
	sKey := base32.StdEncoding.EncodeToString([]byte(storageKey))

	value, err := mw.storage.Read(sKey)
	//check if item is expired.
	if err == nil && !mw.storage.Expired(sKey) {
		readCloser := ioutil.NopCloser(bytes.NewReader(value))
		return readCloser, nil
	}
	//Parse if there is nothing in storage
	parsed, err := mw.Service.Parse(p)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(parsed)
	if err != nil {
		logger.Println(err.Error())
	}

	//save parsed results after scraping to storage
	//calculate expiration time. It is actual for Redis storage.
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := time.Now().UTC().Add(exp)
	logger.Printf("Cache lifespan is %+v\n", expiry.Sub(time.Now().UTC()))

	err = mw.storage.Write(sKey, buf.Bytes(), expiry.Unix())
	if err != nil {
		logger.Println(err.Error())
	}
	output = ioutil.NopCloser(buf)
	return
}
