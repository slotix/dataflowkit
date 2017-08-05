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

	"github.com/slotix/dataflowkit/parse"
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

	var svc parse.Service
	svc = parse.ParseService{}
	svc = parse.CachingMiddleware()(svc)
	svc = parse.LoggingMiddleware(logger)(svc)

	endpoint := parse.Endpoints{
		ParseEndpoint: parse.MakeParseEndpoint(svc),
	}

	r := parse.MakeHttpHandler(ctx, endpoint, logger)

	// HTTP transport
	go func() {
		/*fmt.Println("Checking services ... ")
		status := healthcheck.CheckServices()
		allAlive := true
		for k, v := range status {
			fmt.Printf("%s: %s\n", k, v)
			if v != "Ok" {
				allAlive = false
			}
		}
		if allAlive {
			fmt.Println("Starting server at port 8001")
			handler := r
			errChan <- http.ListenAndServe(":8001", handler)
		}
		*/
		fmt.Println("Starting server at port 8001")
		handler := r
		errChan <- http.ListenAndServe(":8001", handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
