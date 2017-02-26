package server

//TODO https://github.com/happierall/l - ? logger
//
import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"context"
)



func Init(addr string, proxy string) {
	
	/*
		var (
			httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
			consulAddr   = flag.String("consul.addr", "", "Consul agent address")
			retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
			retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
		)*/
	
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("listen", addr).With("caller", log.DefaultCaller)

	ctx := context.Background()

	/*
		fieldKeys := []string{"method", "error"}

			requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "dfk_group",
				Subsystem: "parse_service",
				Name:      "request_count",
				Help:      "Number of requests received.",
			}, fieldKeys)
			requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "dfk_group",
				Subsystem: "parse_service",
				Name:      "request_latency_microseconds",
				Help:      "Total duration of requests in microseconds.",
			}, fieldKeys)
			countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "dfk_group",
				Subsystem: "parse_service",
				Name:      "count_result",
				Help:      "The result of each count method.",
			}, []string{})
	*/

	var svc ParseService
	svc = parseService{}
	svc = proxyingMiddleware(ctx, proxy, logger)(svc)
	svc = loggingMiddleware(logger)(svc)
	//svc = instrumentingMiddleware(requestCount, requestLatency, countResult)(svc)

	getHTMLHandler := httptransport.NewServer(
		ctx,
		makeGetHTMLEndpoint(svc),
		decodeGetHTMLRequest,
		encodeResponse,
	)

	marshalDataHandler := httptransport.NewServer(
		ctx,
		makeMarshalDataEndpoint(svc),
		decodeMarshalDataRequest,
		encodeResponse,
	)

	checkServicesHandler := httptransport.NewServer(
		ctx,
		makeCheckServicesEndpoint(svc),
		decodeCheckServicesRequest,
		encodeCheckServicesResponse,
	)

	
	router := httprouter.New()
	router.Handler("POST", "/app/gethtml", getHTMLHandler)
	router.Handler("POST", "/app/marshaldata", marshalDataHandler)
	router.Handler("POST", "/app/chkservices", checkServicesHandler)
	router.Handler("GET", "/", http.FileServer(http.Dir("web")))
	router.HandlerFunc("GET", "/test1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TESTNAH")
	})
	router.HandlerFunc("GET", "/test2", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index1.html")
	})
	

	router.Handler("GET", "/metrics", stdprometheus.Handler())
	

	logger.Log("msg", "HTTP", "addr", addr)
	logger.Log("err", http.ListenAndServe(addr, router))
}
