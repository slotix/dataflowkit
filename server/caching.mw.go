package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"fmt"

	"github.com/slotix/dataflowkit/cache"
	"github.com/slotix/dataflowkit/parser"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
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

func (mw cachemw) Fetch(req splash.Request) (output io.ReadCloser, err error) {

	redisURL := viper.GetString("redis")
	redisPassword := ""
	redisCon = cache.NewRedisConn(redisURL, redisPassword, "", 0)
	//if something in a cache return local copy
	redisValue, err := redisCon.GetValue(req.URL)
	if err == nil {
		var sResponse splash.Response
		if err := json.Unmarshal(redisValue, &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		//Error responses: a 404 (Not Found) may be cached.
		if sResponse.Response.Status == 404 {
			return nil, fmt.Errorf("Error: 404. NOT FOUND")
		}
		output, err = sResponse.GetContent()
		if err != nil {
			logger.Printf(err.Error())
		}
		return output, err
	}
	//fetch results if there is nothing in a cache
	resp, respErr := mw.ParseService.GetResponse(req)
	if respErr != nil {
		return nil, respErr
	}

	if resp.Cacheable {
		response, err := json.Marshal(resp)
		if err != nil {
			logger.Printf(err.Error())
		}
		err = redisCon.SetValue(req.URL, response)
		if err != nil {
			logger.Println(err.Error())
		}
		err = redisCon.SetExpireAt(req.URL, resp.CacheExpirationTime)
		if err != nil {
			logger.Println(err.Error())
		}
	}

	output, err = resp.GetContent()
	if err != nil {
		return nil, err
	}
	return
}

func (mw cachemw) ParseData(payload []byte) (output io.ReadCloser, err error) {
	redisURL := viper.GetString("redis")
	redisPassword := ""
	redisCon = cache.NewRedisConn(redisURL, redisPassword, "", 0)
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	redisKey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	redisValue, err := redisCon.GetValue(redisKey)
	if err == nil {
		readCloser := ioutil.NopCloser(bytes.NewReader(redisValue))
		return readCloser, nil
	}

	parsed, err := mw.ParseService.ParseData(payload)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(parsed)
	if err != nil {
		logger.Println(err.Error())
	}

	err = redisCon.SetValue(redisKey, buf.Bytes())
	
	if err != nil {
		logger.Println(err.Error())
	}
	err = redisCon.SetExpireIn(redisKey, 3600)
	if err != nil {
		logger.Println(err.Error())
	}
	output = ioutil.NopCloser(buf)
	return
}
