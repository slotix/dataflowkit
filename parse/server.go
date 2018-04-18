package parse

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/slotix/dataflowkit/logger"
	"github.com/slotix/dataflowkit/storage"
	"github.com/spf13/viper"
)

var storageType storage.Type

// Start func launches Parsing service at DFKParse address
func Start(DFKParse string) {
	ctx := context.Background()
	errChan := make(chan error)

	logger := log.NewLogger(false)

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

	r := NewHttpHandler(ctx, endpoints, logger)

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
