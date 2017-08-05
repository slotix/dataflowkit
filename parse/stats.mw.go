package parse

import (
	"io"

	"github.com/slotix/dataflowkit/cache"
	"github.com/spf13/viper"
)

func StatsMiddleware(userID string) ServiceMiddleware {
	return func(next Service) Service {
		return statsMiddleware{userID, next}
	}
}

type statsMiddleware struct {
	userID string
	Service
}

func (mw statsMiddleware) ParseData(payload []byte) (output io.ReadCloser, err error) {
	mw.incrementCount()
	output, err = mw.Service.ParseData(payload)
	return
}

//writing to redis
func (mw statsMiddleware) incrementCount() {
	redisURL := viper.GetString("redis")
	redisPassword := ""
	redis := cache.NewRedisConn(redisURL, redisPassword, "", 0)
	count, err := redis.GetIntValue(mw.userID)
	if count == 0 {
		err = redis.SetValue(mw.userID, 1)
		if err != nil {
			logger.Println(err)
		}
		return
	}
	count++
	err = redis.SetValue(mw.userID, count)
	if err != nil {
		logger.Println(err)
	}
}
