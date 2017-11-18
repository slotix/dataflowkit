package fetch

import (
	"strconv"

	"github.com/slotix/dataflowkit/storage"
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
	s := storage.NewStore(storageType)

	buf, err := s.Read(mw.userID)
	if err != nil {
		logger.Println(err)
	}
	strCount := string(buf)
	count, err := strconv.Atoi(strCount)
	if err != nil {
		logger.Println(err)
	}
	
	count++
	err = s.Write(mw.userID, []byte(strconv.Itoa(count)),0)
	if err != nil {
		logger.Println(err)
	}
}
