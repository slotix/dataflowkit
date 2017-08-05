package fetch

import (
	"encoding/json"

	"fmt"

	"github.com/slotix/dataflowkit/cache"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

type cachingMiddleware struct {
	Service
}

// implement function to return ServiceMiddleware
func CachingMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return cachingMiddleware{next}
	}
}

var redisCon cache.RedisConn

func (mw cachingMiddleware) Fetch(req splash.Request) (output interface{}, err error) {

	redisURL := viper.GetString("redis")
	redisPassword := ""
	redisCon = cache.NewRedisConn(redisURL, redisPassword, "", 0)
	//if something in a cache return local copy
	redisValue, err := redisCon.GetValue(req.URL)
	if err == nil {
		var sResponse *splash.Response
		if err := json.Unmarshal(redisValue, &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		//Error responses: a 404 (Not Found) may be cached.
		if sResponse.Response.Status == 404 {
			return nil, fmt.Errorf("Error: 404. NOT FOUND")
		}
		//output, err = sResponse.GetContent()
		output = sResponse
		//	if err != nil {
		//		logger.Printf(err.Error())
		//	}
		return output, nil
	}

	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	if sResponse, ok := resp.(*splash.Response); ok {
		if sResponse.Cacheable {
			response, err := json.Marshal(resp)
			if err != nil {
				logger.Printf(err.Error())
			}
			err = redisCon.SetValue(req.URL, response)
			if err != nil {
				logger.Println(err.Error())
			}
			err = redisCon.SetExpireAt(req.URL, sResponse.CacheExpirationTime)
			if err != nil {
				logger.Println(err.Error())
			}
		}
		output = sResponse
	}
	//output, err = sResponse.GetContent()
	return
}
