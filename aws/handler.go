package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/slotix/dataflowkit/server"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "fetch: ", log.Lshortfile)
}

func Fetch(req splash.Request) (string, error) {
	viper.Set("splash", "107.22.94.252:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")

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
