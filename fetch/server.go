package fetch

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config provides basic configuration
type Config struct {
	Host    string
	Version string
}

// HTMLServer represents the web service that serves up HTML
type HTMLServer struct {
	server *http.Server
	wg     sync.WaitGroup
}

// Start func launches Parsing service
func Start(cfg Config) *HTMLServer {
	// Setup Context
	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "fetcher",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), os.Stdout, zapcore.DebugLevel)
	logger := zap.New(core)

	defer logger.Sync() // flushes buffer, if any

	var svc Service
	svc = FetchService{}

	//svc = RobotsTxtMiddleware()(svc)
	svc = LoggingMiddleware(logger)(svc)

	endpoints := endpoints{
		fetchEndpoint: makeFetchEndpoint(svc),
	}

	r := newHttpHandler(ctx, endpoints)

	// Create the HTML Server
	htmlServer := HTMLServer{
		server: &http.Server{
			Addr:           cfg.Host,
			Handler:        r,
			MaxHeaderBytes: 1 << 20,
		},
	}

	// Add to the WaitGroup for the listener goroutine
	htmlServer.wg.Add(1)

	go func() {
		fmt.Printf("\n%s\nStarting ...%s",
			cfg.Version,
			htmlServer.server.Addr,
		)
		htmlServer.server.ListenAndServe()
		htmlServer.wg.Done()
	}()
	return &htmlServer
}

// Stop turns off the HTML Server
func (htmlServer *HTMLServer) Stop() error {
	// Create a context to attempt a graceful 5 second shutdown.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("\nFetch Server : Service stopping\n")

	// Attempt the graceful shutdown by closing the listener
	// and completing all inflight requests
	if err := htmlServer.server.Shutdown(ctx); err != nil {
		// Looks like we timed out on the graceful shutdown. Force close.
		if err := htmlServer.server.Close(); err != nil {
			fmt.Printf("\nFetch Server : Service stopping : Error=%v\n", err)
			return err
		}
	}
	// Wait for the listener to report that it is closed.
	htmlServer.wg.Wait()
	fmt.Printf("\nFetch Server : Stopped\n")
	return nil
}
