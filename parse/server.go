package parse

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

// Start func launches Parsing service at DFKParse address
func Start(DFKParse string) {
	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		//logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "ts", time.Now().Format("Jan _2 15:04:05"))
		//logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var svc Service
	svc = ParseService{}
	//svc = StatsMiddleware("18")(svc)
	if !viper.GetBool("SKIP_STORAGE_MW") {
		var err error
		storageType, err = storage.TypeString(viper.GetString("STORAGE_TYPE"))
		if err != nil {
			errChan <- fmt.Errorf("%s", err)
		}
		svc = StorageMiddleware(storage.NewStore(storageType))(svc)
	}
	svc = LoggingMiddleware(logger)(svc)

	endpoints := Endpoints{
		ParseEndpoint: MakeParseEndpoint(svc),
	}

	r := MakeHttpHandler(ctx, endpoints, logger)

	// HTTP transport
	go func() {
		handler := r
		errChan <- http.ListenAndServe(DFKParse, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
