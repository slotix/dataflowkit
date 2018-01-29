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
)

var storageType string

// Start func launches Parsing service at DFKFetch address
func Start(DFKFetch string) {
	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", time.Now().Format("Jan _2 15:04:05"))
		//logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var svc Service
	svc = FetchService{}
	//svc = StatsMiddleware("18")(svc)
	//	storageType = viper.GetString("STORAGE_TYPE")
	//	svc = StorageMiddleware(storage.NewStore(storageType))(svc)
	svc = RobotsTxtMiddleware()(svc)
	svc = LoggingMiddleware(logger)(svc)

	endpoints := Endpoints{
		SplashFetchEndpoint:    MakeSplashFetchEndpoint(svc),
		SplashResponseEndpoint: MakeSplashResponseEndpoint(svc),
		BaseFetchEndpoint:      MakeBaseFetchEndpoint(svc),
		BaseResponseEndpoint:   MakeBaseResponseEndpoint(svc),
	}

	r := NewHttpHandler(ctx, endpoints, logger)

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
