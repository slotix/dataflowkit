package parse

import (
	"io"

	"github.com/slotix/dataflowkit/scrape"
)

// Define service interface
type Service interface {
	Parse(scrape.Payload) (io.ReadCloser, error)
}

// Implement service with empty struct
type ParseService struct {
}

type ServiceMiddleware func(Service) Service

//Parse calls Fetcher to download web page for parsing
func (ps ParseService) Parse(p scrape.Payload) (io.ReadCloser, error) {
	r, err := scrape.Parse(p)
	if err != nil {
		return nil, err
	}
	return r, nil
}
