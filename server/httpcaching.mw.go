package server

import (
	"errors"
	"fmt"

	"github.com/slotix/dfk-parser/cache"
)

var errRedisSet = errors.New("Redis. Set Value failed")

func httpCachingMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return httpcachemw{next}
	}
}

type httpcachemw struct {
	ParseService
}

func (mw httpcachemw) Download(url string) (output []byte, err error) {
	redisURL := "localhost:6379"
	redisPassword := ""
	redis := cache.NewRedisConn(redisURL, redisPassword, "", 0)
	output, err = redis.GetValue(url)
	if err == nil {
		return output, nil
	}
	output, err = mw.ParseService.Download(url)
	if err == nil {
		err1 := redis.SetValue(url, output)
		if err1 != nil {
			fmt.Printf("%s: %s", errRedisSet, err1.Error())
		}
	//	return content, nil
	}
	return
}
