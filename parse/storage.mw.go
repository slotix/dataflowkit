package parse

import (
	"bytes"
	"fmt"
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
	storageKey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	value, err := s.Read(storageKey)
	//check if item is expired.
	if err == nil && !s.Expired(storageKey) {	
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

	//check if item is expired.
	exp := time.Duration(viper.GetInt64("STORAGE_EXPIRE")) * time.Second
	expiry := time.Now().UTC().Add(exp)
	logger.Printf("Cache lifespan is %+v\n", expiry)

	err = s.Write(storageKey, buf.Bytes(), expiry.Unix())
	logger.Println(err)
	if err != nil {
		logger.Println(err.Error())
	}
	output = ioutil.NopCloser(buf)
	return
}
