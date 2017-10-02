package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"github.com/slotix/dataflowkit/server"
	"github.com/slotix/dataflowkit/splash"
)

var Handler apigatewayproxy.Handler

func init() {
	Handler = NewHandler()
}

func NewHandler() apigatewayproxy.Handler {
	ln := net.Listen()

	handle := apigatewayproxy.New(ln, nil).Handle

	http.HandleFunc("/fetch", handleFetch)

	go http.Serve(ln, nil)

	return handle
}

func handleFetch(w http.ResponseWriter, r *http.Request) {
	var request splash.Request
	req, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(req, &request)
	if err != nil {
		log.Println(err)
	}
	response, err := Fetch(request)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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
