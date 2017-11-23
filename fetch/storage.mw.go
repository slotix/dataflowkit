package fetch

import (
	"encoding/base32"
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

//TODO: it is ugly but this two funcs are completely identical except
//resp, err := mw.Service.Fetch(req) and resp, err := mw.Service.Response(req) strings accordingly. Need to invent something better.

//Fetch endpoint storage middleware is called from Fetch endoint 
func (mw storageMiddleware) Fetch(req interface{}) (output interface{}, err error) {
	s := storage.NewStore(mw.StorageType)

	//if something in a cache return local copy
	var sKey string
	if sReq, ok := req.(splash.Request); ok {
		url := sReq.GetURL()
		//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
		sKey = base32.StdEncoding.EncodeToString([]byte(url))
		value, err := s.Read(sKey)

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
		err = s.Write(sKey, response, expTime)
		if err != nil {
			logger.Println(err.Error())
		}
		output = sResponse
	}
	return
} 

//Response endpoint storage middleware is called from Parse endoint 
func (mw storageMiddleware) Response(req interface{}) (output interface{}, err error) {
	s := storage.NewStore(mw.StorageType)

	//if something in a cache return local copy
	var sKey string
	if sReq, ok := req.(splash.Request); ok {
		url := sReq.GetURL()
		//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
		sKey = base32.StdEncoding.EncodeToString([]byte(url))
		value, err := s.Read(sKey)

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

	//Current err value is not passed outside.
	err = nil
	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Response(req)
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
		err = s.Write(sKey, response, expTime)
		if err != nil {
			logger.Println(err.Error())
		}
		output = sResponse
	}
	return
}
