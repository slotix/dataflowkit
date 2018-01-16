package parse

import (
	"io"
	"strconv"

	"github.com/slotix/dataflowkit/scrape"
	"github.com/slotix/dataflowkit/storage"
)

// StatsMiddleware tracks requests sent to Parse endpoint.
func StatsMiddleware(userID string) ServiceMiddleware {
	return func(next Service) Service {
		return statsMiddleware{userID, next}
	}
}

type statsMiddleware struct {
	userID string
	Service
}

//Parse increments counter before sending it to actual Parse service handler.
func (mw statsMiddleware) Parse(payload scrape.Payload) (output io.ReadCloser, err error) {
	mw.incrementCount()
	output, err = mw.Service.Parse(payload)
	return
}

//writing to redis
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
	err = s.Write(mw.userID, []byte(strconv.Itoa(count)), 0)
	if err != nil {
		logger.Error(err)
	}
}
