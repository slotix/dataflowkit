package parse

import (
	"bytes"
	"io"
	"io/ioutil"

	"fmt"

	"github.com/slotix/dataflowkit/cache"
	"github.com/slotix/dataflowkit/scrape"
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

func (mw cachingMiddleware) ParseData(p scrape.Payload) (output io.ReadCloser, err error) {
	redisURL := viper.GetString("REDIS")
	redisPassword := ""
	redisCon = cache.NewRedisConn(redisURL, redisPassword, "", 0)
	//	p, err := scrape.NewPayload(payload)
	//	if err != nil {
	//		return nil, err
	//	}
	redisKey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	redisValue, err := redisCon.GetValue(redisKey)
	if err == nil {
		readCloser := ioutil.NopCloser(bytes.NewReader(redisValue))
		return readCloser, nil
	}
	parsed, err := mw.Service.ParseData(p)
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
