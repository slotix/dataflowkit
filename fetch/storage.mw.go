package fetch

import (
	"encoding/base32"
	"encoding/json"
	"time"

	"github.com/slotix/dataflowkit/errs"

	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/storage"
)

type storageMiddleware struct {
	//StorageType storage.Type
	storage storage.Store
	Service
}

// implement function to return ServiceMiddleware
func StorageMiddleware(storage storage.Store) ServiceMiddleware {
	return func(next Service) Service {
		return storageMiddleware{storage, next}
	}
}


func (mw storageMiddleware) get(req FetchRequester) (resp FetchResponser, err error) {
	//s := storage.NewStore(mw.StorageType)
	var fetchResponse FetchResponser
	url := req.GetURL()
	
	switch req.(type) {
	case BaseFetcherRequest:
		fetchResponse = &BaseFetcherResponse{}
	case splash.Request:
		fetchResponse = &splash.Response{}
	default:
		panic("invalid fetcher request")
	}
	//Base32 encoded values are 100% safe for file/uri usage without replacing any characters and guarantees 1-to-1 mapping
	sKey := base32.StdEncoding.EncodeToString([]byte(url))
	value, err := mw.storage.Read(sKey)
	if err == nil {
		if err := json.Unmarshal(value, &fetchResponse); err != nil {
			logger.Println(err)
		}
		//Error responses: a 404 (Not Found) may be cached.
		//if sResponse.Response.Status == 404 {
		//	return nil, &errs.NotFound{URL: url}
		//}
		//check if item is expired.
		expired := fetchResponse.GetExpires()
		logger.Println(expired)
		diff := expired.Sub(time.Now().UTC())
		logger.Printf("%s: cache lifespan is %+v\n", url, diff)

		if diff > 0 { //if cached value is not expired return it
			//output = fetchResponse
			return fetchResponse, nil
		}

		err = &errs.ExpiredItemOrNotCacheable{}
	}
	return nil, err
}


func (mw storageMiddleware) put(req FetchRequester, resp FetchResponser) error {
	url := req.GetURL()
	sKey := base32.StdEncoding.EncodeToString([]byte(url))
	logger.Println("Cachable? ", resp.GetCacheable())
	expired := resp.GetExpires()

	logger.Println(expired)

	r, err := json.Marshal(resp)
	if err != nil {
		logger.Printf(err.Error())
		return err
	}
	//calculate expiration time. This is actual for Redis only.
	//logger.Println(fetchResponse.Expires())
	expTime := expired.Unix()
	err = mw.storage.Write(sKey, r, expTime)
	if err != nil {
		logger.Println(err.Error())
		return err
	}
	return nil
}

//Fetch endpoint storage middleware is called from Fetch endoint
func (mw storageMiddleware) Fetch(req FetchRequester) (FetchResponser, error) {

	fromStorage, err := mw.get(req)
	if err == nil {
		return fromStorage, nil
	}
	logger.Println(err)
	//Current err value should be nilled.
	err = nil
	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	
	var fetchResponse FetchResponser
	switch req.(type) {
	case BaseFetcherRequest:
		fetchResponse = resp.(*BaseFetcherResponse)
	case splash.Request:
		fetchResponse = resp.(*splash.Response)
	default:
		panic("invalid fetcher request")
	}
	err = mw.put(req, fetchResponse)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//Response endpoint storage middleware is called from Parse endoint
func (mw storageMiddleware) Response(req FetchRequester) (FetchResponser, error) {
	fromStorage, err := mw.get(req)
	if err == nil {
		return fromStorage, nil
	}
	logger.Println(err)
	//Current err value should be nilled.
	err = nil
	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Response(req)
	if err != nil {
		return nil, err
	}
	var fetchResponse FetchResponser
	switch req.(type) {
	case BaseFetcherRequest:
		fetchResponse = resp.(*BaseFetcherResponse)
	case splash.Request:
		fetchResponse = resp.(*splash.Response)
	default:
		panic("invalid fetcher request")
	}
	err = mw.put(req, fetchResponse)
	if err != nil {
		return nil, err
	}
	return resp, nil
}