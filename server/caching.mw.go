package server

import (
	"encoding/json"

	"fmt"

	"github.com/slotix/dataflowkit/cache"
	"github.com/slotix/dataflowkit/downloader"
	"github.com/slotix/dataflowkit/parser"
)

func cachingMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return cachemw{next}
	}
}

type cachemw struct {
	ParseService
}

var redisCon cache.RedisConn

func init() {
	redisURL := "localhost:6379"
	redisPassword := ""
	redisCon = cache.NewRedisConn(redisURL, redisPassword, "", 0)
}

func (mw cachemw) Download(url string) (output []byte, err error) {
	debug := true
	redisValue, err := redisCon.GetValue(url)
	if err == nil {
		var sResponse downloader.SplashResponse
		if err := json.Unmarshal(redisValue, &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		//Error responses: a 404 (Not Found) may be cached.
		if sResponse.Response.Status == 404 {
			return nil, fmt.Errorf("Error: 404. NOT FOUND")
		}
		output, err = sResponse.GetHTML()
		if err != nil {
			logger.Printf(err.Error())
		}
		return output, err
	}

	resp, respErr := mw.ParseService.GetResponse(url)
	//Check if it is cacheable 
	rv := cache.Cacheable(resp)
	expTime := rv.OutExpirationTime.Unix()
	if debug {
		if rv.OutErr != nil {
			logger.Println("Errors: ", rv.OutErr)
		}
		if rv.OutReasons != nil {
			logger.Println("Reasons to not cache: ", rv.OutReasons)
		}
		if rv.OutWarnings != nil {
			logger.Println("Warning headers to add: ", rv.OutWarnings)
		}
		logger.Println("Expiration: ", rv.OutExpirationTime.String())
	}
	//write data to cache
	if len(rv.OutReasons) == 0 {
		response, err := json.Marshal(resp)
		if err != nil {
			logger.Printf(err.Error())
		}
		err = redisCon.SetValue(url, response)
		if err != nil {
			logger.Println(err.Error())
		}
		err = redisCon.SetExpireAt(url, expTime)
		if err != nil {
			logger.Println(err.Error())
		}
	}
	if respErr != nil {
		return nil, respErr
	}

	output, err = resp.GetHTML()
	if err != nil {
		return nil, err
	}
	return
}

func (mw cachemw) ParseData(payload []byte) (output []byte, err error) {
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	rediskey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	redisValue, err := redisCon.GetValue(rediskey)
	if err == nil {
		return redisValue, nil
	}

	output, err = mw.ParseService.ParseData(payload)

	err = redisCon.SetValue(rediskey, output)
	if err != nil {
		logger.Println(err.Error())
	}
	err = redisCon.SetExpireIn(rediskey, 3600)
	if err != nil {
		logger.Println(err.Error())
	}
	return
}
