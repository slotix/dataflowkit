package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/slotix/dataflowkit/errs"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/parse"
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

type Endpoints struct {
	FetchEndpoint endpoint.Endpoint
	ParseEndpoint endpoint.Endpoint
}

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

	var fSvc fetch.Service
	fSvc = fetch.FetchService{}

	//svc = StatsMiddleware("18")(svc)
	//svc = CachingMiddleware()(svc)
	fSvc = SQSMiddleware()(fSvc)
	fSvc = fetch.LoggingMiddleware(logger)(fSvc)
	fSvc = fetch.RobotsTxtMiddleware()(fSvc)

	var pSvc parse.Service
	pSvc = parse.ParseService{}
	pSvc = parse.LoggingMiddleware(logger)(pSvc)

	endpoints := Endpoints{
		FetchEndpoint: fetch.MakeFetchEndpoint(fSvc),
		ParseEndpoint: parse.MakeParseEndpoint(pSvc),
	}
	//endpoints := fetch.Endpoints{
	//	FetchEndpoint: fetch.MakeFetchEndpoint(svc),
	//ParseEndpoint: server.MakeParseEndpoint(svc),
	//}
	//r := fetch.MakeHttpHandler(ctx, endpoints, logger)
	r := MakeHttpHandler(ctx, endpoints, logger)
	ln := net.Listen()
	handle := apigatewayproxy.New(ln, nil).Handle
	go http.Serve(ln, r)
	return handle
}

// Make Http Handler
func MakeHttpHandler(ctx context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//"/app/parse"
	r.Methods("POST").Path("/fetch").Handler(httptransport.NewServer(
		endpoint.FetchEndpoint,
		fetch.DecodeFetchRequest,
		fetch.EncodeFetchResponse,
		options...,
	))

	r.Methods("POST").Path("/parse").Handler(httptransport.NewServer(
		endpoint.ParseEndpoint,
		parse.DecodeParseRequest,
		parse.EncodeParseResponse,
		options...,
	))

	return r
}

// encode error
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//logger.Println(err)
	//t = err.(type)
	var httpStatus int
	switch err.(type) {
	default:
		httpStatus = http.StatusInternalServerError
	case *errs.BadRequest,
		*errs.InvalidHost,
		*errs.Error:
		//return 400 Status
		httpStatus = http.StatusBadRequest
	case *errs.ForbiddenByRobots,
		*errs.Forbidden:
		//return 403 Status
		httpStatus = http.StatusForbidden
	case *errs.NotFound:
		//return 404 Status
		httpStatus = http.StatusNotFound
	case *errs.GatewayTimeout:
		//return 504 Status
		httpStatus = http.StatusGatewayTimeout
	}

	w.WriteHeader(httpStatus)
	//AWS error payload should looks like
	//{
	//"errorType": "BadRequest",
	//"httpStatus": httpStatus,
	//"requestId" : "<context.awsRequestId>",
	//"message": err.Error(),
	//}
	//according to the information from https://aws.amazon.com/blogs/compute/error-handling-patterns-in-amazon-api-gateway-and-aws-lambda/

	//But it seems enough to w.WriteHeader(httpStatus) and send an error only
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
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
