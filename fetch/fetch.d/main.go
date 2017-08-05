package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"

	"github.com/slotix/dataflowkit/fetch"
)

func main() {
	viper.Set("splash", "127.0.0.1:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")
	viper.Set("redis", "127.0.0.1:6379")
	viper.Set("redis-expire", "3600")
	viper.Set("redis-network", "tcp")

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

	var svc fetch.Service
	svc = fetch.FetchService{}
	svc = fetch.StatsMiddleware("18")(svc)
	svc = fetch.CachingMiddleware()(svc)
	svc = fetch.LoggingMiddleware(logger)(svc)
	//svc = fetch.RobotsTxtMiddleware()(svc)
	
	endpoint := fetch.Endpoints{
		FetchEndpoint: fetch.MakeFetchEndpoint(svc),
	}

	r := fetch.MakeHttpHandler(ctx, endpoint, logger)

	// HTTP transport
	go func() {
		fmt.Println("Starting server at port 8000")
		handler := r
		errChan <- http.ListenAndServe(":8000", handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
