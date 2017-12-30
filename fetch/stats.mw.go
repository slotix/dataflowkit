package fetch

import (
	"strconv"

	"github.com/slotix/dataflowkit/storage"
)

// StatsMiddleware tracks requests sent to Fetch endpoint.
func StatsMiddleware(userID string) ServiceMiddleware {
	return func(next Service) Service {
		return statsMiddleware{userID, next}
	}
}

type statsMiddleware struct {
	userID string
	Service
}


//Fetch increments requst count before sending it to actual Fetch service handler.
func (mw statsMiddleware) Fetch(req FetchRequester) (response FetchResponser, err error) {
	mw.incrementCount()
	response, err = mw.Service.Fetch(req)
	return
}

//Response increments requst count before sending it to actual Response service handler.
func (mw statsMiddleware) Response(req FetchRequester) (response FetchResponser, err error) {
	mw.incrementCount()
	response, err = mw.Service.Response(req)
	return
}


func (mw statsMiddleware) incrementCount() {
	s := storage.NewStore(storageType)

	buf, err := s.Read(mw.userID)
	if err != nil {
		logger.Error(err)
	}
	strCount := string(buf)
	count, err := strconv.Atoi(strCount)
	if err != nil {
		logger.Error(err)
	}
	
	count++
	err = s.Write(mw.userID, []byte(strconv.Itoa(count)),0)
	if err != nil {
		logger.Error(err)
	}
}
