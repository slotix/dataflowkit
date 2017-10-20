package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"github.com/go-kit/kit/log"
	"github.com/slotix/dataflowkit/fetch"
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
	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		//logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "ts", time.Now().Format("Jan _2 15:04:05"))
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var svc fetch.Service
	svc = fetch.FetchService{}

	//svc = StatsMiddleware("18")(svc)
	//svc = CachingMiddleware()(svc)
	svc = fetch.SQSMiddleware()(svc)
	svc = fetch.LoggingMiddleware(logger)(svc)
	svc = fetch.RobotsTxtMiddleware()(svc)

	endpoints := fetch.Endpoints{
		FetchEndpoint: fetch.MakeFetchEndpoint(svc),
	//	ParseEndpoint: server.MakeParseEndpoint(svc),
	}
	r := fetch.MakeHttpHandler(ctx, endpoints, logger)
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
