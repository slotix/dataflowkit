package fetch

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
)

var storageType storage.Type

func Start(DFKFetch string) {
	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", time.Now().Format("Jan _2 15:04:05"))
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	//creating storage for cache html content and parsed results
	sType := viper.GetString("STORAGE_TYPE")
	switch sType {
	case "S3":
		storageType = storage.S3
	case "Redis":
		storageType = storage.Redis
	case "Diskv":
		storageType = storage.Diskv
	default:
		panic("Storage type is undefined")
	}
	storage := storage.NewStore(storageType)
	var svc Service
	svc = FetchService{}
	//svc = StatsMiddleware("18")(svc)
	svc = RobotsTxtMiddleware()(svc)
	svc = StorageMiddleware(storage)(svc) //possible values are Diskv, S3, Redis
	svc = LoggingMiddleware(logger)(svc)

	endpoints := Endpoints{
		SplashFetchEndpoint:    MakeSplashFetchEndpoint(svc),
		SplashResponseEndpoint: MakeSplashResponseEndpoint(svc),
		BaseFetchEndpoint:      MakeBaseFetchEndpoint(svc),
		BaseResponseEndpoint:   MakeBaseResponseEndpoint(svc),
	}

	r := MakeHttpHandler(ctx, endpoints, logger)

	// HTTP transport
	go func() {
		handler := r
		errChan <- http.ListenAndServe(DFKFetch, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
