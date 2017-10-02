package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/slotix/dataflowkit/server"
	"github.com/slotix/dataflowkit/splash"
)


func Fetch(req splash.Request) (string, error) {
	//viper.Set("SPLASH", "107.22.94.252:8050")
	//viper.Set("SPLASH_TIMEOUT", "20")
	//viper.Set("SPLASH_RESOURCE_TIMEOUT", "30")
	//viper.Set("SPLASH_WAIT", "0,5")
	log.Printf("Fetching %s\n", req.URL)
	fetcher, err := server.NewSplashFetcher()
	if err != nil {
		log.Println(err)
	}

	response, err := fetcher.Fetch(req)
	if err != nil {
		return "", err
	}
	sResponse := response.(*splash.Response)
	content, err := sResponse.GetContent()
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(content)

	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Handle(evt json.RawMessage, ctx *runtime.Context) (string, error) {
	//log.Printf("%T", evt)
	var request splash.Request
	err := json.Unmarshal(evt, &request)
	if err != nil {
		return "", err
	}
	//request := evt.(splash.Request)
	result, err := Fetch(request)
	if err != nil {
		return "", err
	}
	return result, nil
}
