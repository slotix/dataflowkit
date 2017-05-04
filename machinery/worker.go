package main

import (
	"github.com/slotix/dataflowkit/splash"
	"github.com/slotix/dataflowkit/parser"
)

//import "time"

func GetHTML(url string) (string, error) {
	_, err := splash.Fetch(url)
	if err != nil {
		return "", err
	}
	//	time.Sleep(10 * time.Second)
	return "200", nil
}

/*
func GetHTML1(url string) (string, error) {
	content, err := splash.Fetch(url)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
*/

func MarshalData(payload []byte) ([]byte, error) {
	parser, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	res, err := parser.MarshalData()
	if err != nil {
		return nil, err
	}
	return res, nil
}
