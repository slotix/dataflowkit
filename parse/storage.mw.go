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

type storageMiddleware struct {
	StorageType storage.Type
	Service
}

// implement function to return ServiceMiddleware
func StorageMiddleware(storageType storage.Type) ServiceMiddleware {
	return func(next Service) Service {
		return storageMiddleware{storageType, next}
	}
}

func (mw storageMiddleware) ParseData(p scrape.Payload) (output io.ReadCloser, err error) {
	s := storage.NewStore(mw.StorageType)
	//if something in a cache return local copy
	//storageKey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	storageKey := string(p.PayloadMD5)
	//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
	sKey := base32.StdEncoding.EncodeToString([]byte(storageKey))

	value, err := s.Read(sKey)
	//check if item is expired.
	if err == nil && !s.Expired(sKey) {
		readCloser := ioutil.NopCloser(bytes.NewReader(value))
		return readCloser, nil
	}
	//Parse if there is nothing in a cache
	parsed, err := mw.Service.ParseData(p)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(parsed)
	if err != nil {
		logger.Println(err.Error())
	}

	//calculate expiration time. It is actual for Redis storage.
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := time.Now().UTC().Add(exp)
	logger.Printf("Cache lifespan is %+v\n", expiry.Sub(time.Now().UTC()))

	err = s.Write(sKey, buf.Bytes(), expiry.Unix())
	if err != nil {
		logger.Println(err.Error())
	}
	output = ioutil.NopCloser(buf)
	return
}
