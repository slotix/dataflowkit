package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"

	"github.com/slotix/dataflowkit/server"
)

func main() {
	viper.Set("splash", "127.0.0.1:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")
	viper.Set("redis", "127.0.0.1:6379")
	viper.Set("redis-expire", "3600")
	viper.Set("redis-network", "tcp")
	var (
		port = flag.String("listen", ":8000", "HTTP listen address")
	)
	flag.Parse()
	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		//logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "ts", time.Now().Format("Jan _2 15:04:05"))
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var svc server.Service
	svc = server.ParseService{}
	svc = server.StatsMiddleware("18")(svc)
	//svc = server.CachingMiddleware()(svc)
	//svc = server.ProxyingMiddleware(ctx, "http://127.0.0.1:8000")(svc)
	//svc = server.ProxyingMiddleware(ctx,"")(svc)

	svc = server.LoggingMiddleware(logger)(svc)
	svc = server.RobotsTxtMiddleware()(svc)

	endpoints := server.Endpoints{
		FetchEndpoint: server.MakeFetchEndpoint(svc),
		ParseEndpoint: server.MakeParseEndpoint(svc),
	}

	r := server.MakeHttpHandler(ctx, endpoints, logger)

	// HTTP transport
	go func() {
		fmt.Printf("Starting server at port %s\n", *port)
		handler := r
		errChan <- http.ListenAndServe(*port, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
