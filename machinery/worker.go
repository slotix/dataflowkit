package main

import "github.com/slotix/dataflowkit/parser"
//import "time"

func GetHTML(url string) (string, error) {
	_, err := parser.GetHTML(url)
	if err != nil {
		return "", err
	}
//	time.Sleep(10 * time.Second)
	return "200", nil
}

func GetHTML1(url string) (string, error) {
	content, err := parser.GetHTML(url)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func MarshalData(payload []byte) ([]byte, error) {
	res, err := parser.MarshalData(payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}
