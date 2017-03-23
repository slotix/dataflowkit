package server

import "github.com/slotix/dfk-parser/cache"
import "fmt"

func statsMiddleware(userID string) ServiceMiddleware {
	return func(next ParseService) ParseService {
		return statsmw{userID, next}
	}
}

type statsmw struct {
	userID string
	ParseService
}

func (mw statsmw) ParseData(payload []byte) (output []byte, err error) {
	mw.incrementCount()
	output, err = mw.ParseService.ParseData(payload)
	return
}

func (mw statsmw) Download(url string) (output []byte, err error) {
	mw.incrementCount()
	output, err = mw.ParseService.Download(url)
	return
}

//temporarily writing to redis
func (mw statsmw) incrementCount() {
	redisURL := "localhost:6379"
	redisPassword := ""
	redis := cache.NewRedisConn(redisURL, redisPassword, "", 0)
	count, err := redis.GetIntValue(mw.userID)
	if count == 0 {
		err = redis.SetValue(mw.userID, 1)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	count++
	err = redis.SetValue(mw.userID, count)
	if err != nil {
		fmt.Println(err)
	}
}
