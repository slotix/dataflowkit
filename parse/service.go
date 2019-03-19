package parse

import (
	"io"

	"github.com/slotix/dataflowkit/scrape"
)

// Service defines Parse service interface
type Service interface {
	Parse(scrape.Payload) (io.ReadCloser, error)
}

// ParseService implements service with empty struct
type ParseService struct {
}

// ServiceMiddleware defines a middleware for a Parse service
type ServiceMiddleware func(Service) Service

//Parse service processes fetched page following the rules from Payload.
func (ps ParseService) Parse(p scrape.Payload) (io.ReadCloser, error) {
	task := scrape.NewTask()
	r, err := task.Parse(p)
	if err != nil {
		return nil, err
	}
	return r, nil
}
