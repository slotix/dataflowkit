package fetch

import (
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

func (mw statsMiddleware) Fetch(req interface{}) (output interface{}, err error) {
	mw.incrementCount()
	output, err = mw.Service.Fetch(req)
	return
}


func (mw statsMiddleware) Response(req interface{}) (output interface{}, err error) {
	mw.incrementCount()
	output, err = mw.Service.Response(req)
	return
}


//writing to redis
func (mw statsMiddleware) incrementCount() {
	redisURL := viper.GetString("redis")
	//logger.Println(redisURL)
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
