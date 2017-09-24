package main

import (
	"testing"
	"github.com/slotix/dataflowkit/splash"
)

func TestFetch(t *testing.T) {
	request := splash.Request{URL:"http://google.com"}
	result, err := Fetch(request)
	if err != nil {
		logger.Println(err) 
	}
	logger.Println(result)
}

func TestHandle(t *testing.T) {
	//request := splash.Request{URL:"http://google.com"}
	//evt := make(map[string]interface{})
	//evt["url"] = "http://google.com"
	evt := []byte(`{"url":"http://google.com"}`)
	result, err := Handle(evt, nil)
	if err != nil {
		logger.Println(err) 
	}
	logger.Println(result)
}