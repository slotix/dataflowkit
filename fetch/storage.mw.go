package fetch

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/slotix/dataflowkit/errs"

	"github.com/slotix/dataflowkit/splash"
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

func (mw storageMiddleware) Fetch(req interface{}) (output interface{}, err error) {
	s := storage.NewStore(mw.StorageType)

	//if something in a cache return local copy
	var url string
	if sReq, ok := req.(splash.Request); ok {
		url = sReq.GetURL()
		value, err := s.Read(url)

		//if err == nil && !s.Expired(url) {
		if err == nil {
			var sResponse *splash.Response
			if err := json.Unmarshal(value, &sResponse); err != nil {
				logger.Println("Json Unmarshall error", err)
			}
			//Error responses: a 404 (Not Found) may be cached.
			//if sResponse.Response.Status == 404 {
			//	return nil, &errs.NotFound{URL: url}
			//}
			//check if item is expired.
			diff := sResponse.Expires.Sub(time.Now().UTC())
			logger.Printf("%s: cache lifespan is %+v\n", url, diff)

			if diff > 0 { //if cached value is not expired return it
				output = sResponse
				return output, nil
			}
			err = &errs.ExpiredItemOrNotCacheable{}
		}
	} else {
		logger.Println("Bad request")
		return nil, errors.New("Bad request")

	}

	logger.Println(err)
	//Current err value is not passed outside.
	err = nil
	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	if sResponse, ok := resp.(*splash.Response); ok {
		logger.Println("Cachable? ", sResponse.Cacheable)
		response, err := json.Marshal(resp)
		if err != nil {
			logger.Printf(err.Error())
		}
		//calculate expiration time. This is actual for Redis only.
		expTime := sResponse.Expires.Unix()
		err = s.Write(url, response, expTime)
		if err != nil {
			logger.Println(err.Error())
		}
		output = sResponse
	}
	return
}
