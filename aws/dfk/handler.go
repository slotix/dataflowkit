package main

import (
	"context"
	"net/http"

	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"github.com/slotix/dataflowkit/server"
)

var Handler apigatewayproxy.Handler

func init() {
	Handler = NewHandler()
}

/*
func NewHandler() apigatewayproxy.Handler {
	ln := net.Listen()
	handle := apigatewayproxy.New(ln, nil).Handle
	http.HandleFunc("/fetch", handleFetch)
	go http.Serve(ln, nil)
	return handle
}
*/

func NewHandler() apigatewayproxy.Handler {
	ctx := context.Background()
	
	var svc server.Service
	svc = server.ParseService{}

	//svc = StatsMiddleware("18")(svc)
	//svc = CachingMiddleware()(svc)
	//svc = LoggingMiddleware(logger)(svc)
	svc = server.RobotsTxtMiddleware()(svc)

	endpoints := server.Endpoints{
		FetchEndpoint: server.MakeFetchEndpoint(svc),
		ParseEndpoint: server.MakeParseEndpoint(svc),
	}
	r := server.MakeHttpHandler(ctx, endpoints, nil)
	ln := net.Listen()
	handle := apigatewayproxy.New(ln, nil).Handle
	go http.Serve(ln, r)
	return handle
}

/*
func handleFetch(w http.ResponseWriter, r *http.Request) {
	var request splash.Request
	req, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(req, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := Fetch(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func Fetch(req splash.Request) (string, error) {
	fetcher, err := server.NewSplashFetcher()
	if err != nil {
		return "", err
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
*/
