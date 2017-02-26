package server

import (
	"github.com/go-kit/kit/log"
	"github.com/slotix/dfk-parser/parser"
)

var logger log.Logger

// ParseService provides operations on strings.
type ParseService interface {
	GetHTML(string) ([]byte, error)
	MarshalData(payload []byte) ([]byte, error)
	CheckServices() (status map[string]string)
	//	Save(payload []byte) (string, error)
}

type parseService struct{}

func (parseService) GetHTML(url string) ([]byte, error) {
	content, err := parser.GetHTML(url)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (parseService) MarshalData(payload []byte) ([]byte, error) {
	res, err := parser.MarshalData(payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (parseService) CheckServices() (status map[string]string) {
	return CheckServices() //, allAlive
}

// ServiceMiddleware is a chainable behavior modifier for ParseService.
type ServiceMiddleware func(ParseService) ParseService