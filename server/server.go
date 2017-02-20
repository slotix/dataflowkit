package server

//TODO https://github.com/happierall/l - ? logger
//	"github.com/julienschmidt/httprouter"

import (
	"os"
	"net/http"
	"fmt"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

func Init(port int) {
	//config
	/*
		viper.SetConfigName("config")
		viper.AddConfigPath(".././config")
		viper.AddConfigPath("config")
		viper.AddConfigPath("$GOPATH/src/dataflowkit/config")
		viper.AddConfigPath("$GOPATH/bin/server")

		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	*/
	/*
		var (
			httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
			consulAddr   = flag.String("consul.addr", "", "Consul agent address")
			retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
			retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
		)*/
	var (
	//	listen = flag.String("listen", viper.GetString("server.port"), "HTTP listen address")
	//listen = flag.String("listen", , "HTTP listen address")

	//	proxy  = flag.String("proxy", "", "Optional comma-separated list of URLs to proxy parsing requests")
	)
	//flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	//logger = log.NewContext(logger).With("listen", *listen).With("caller", log.DefaultCaller)
	logger = log.NewContext(logger).With("listen", port).With("caller", log.DefaultCaller)

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
	//	svc = proxyingMiddleware(*proxy, ctx, logger)(svc)
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
		encodeResponse,
	)

	//router := mux.NewRouter().StrictSlash(true)
	//router.HandleFunc("/", heartbeat)
	//logger.Log(http.ListenAndServe(":8080", router))
//	router := httprouter.New()
//	router.GET("/", indexHandler)
	http.Handle("/gethtml", getHTMLHandler)
	http.Handle("/marshaldata", marshalDataHandler)
	http.Handle("/chkservices", checkServicesHandler)
	http.Handle("/", http.FileServer(http.Dir("./")))
	//http.Handle("/", )

	http.Handle("/metrics", stdprometheus.Handler())
	//logger.Log("msg", "HTTP", "addr", *listen)
	//logger.Log("err", http.ListenAndServe(*listen, nil))
	logger.Log("msg", "HTTP", "addr", port)
	logger.Log("err", http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
