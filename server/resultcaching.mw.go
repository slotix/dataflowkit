package server

import (
	"fmt"

	"github.com/slotix/dfk-parser/cache"
	"github.com/slotix/dfk-parser/parser"
)

func resultCachingMiddleware() ServiceMiddleware {
	return func(next ParseService) ParseService {
		return resultcachemw{next}
	}
}

type resultcachemw struct {
	ParseService
}

func (mw resultcachemw) ParseData(payload []byte) (output []byte, err error) {
	redisURL := "localhost:6379"
	redisPassword := ""
	redis := cache.NewRedisConn(redisURL, redisPassword, "", 0)
	p, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	rediskey := fmt.Sprintf("%s-%s", p.Format, p.PayloadMD5)
	redisValue, err := redis.GetValue(rediskey)
	if err == nil {
		return redisValue, nil
	}

	output, err = mw.ParseService.ParseData(payload)

	err = redis.SetValue(rediskey, output)
	if err != nil {
		fmt.Printf("%s: %s", errRedisSet, err.Error())
	}
	return
}
