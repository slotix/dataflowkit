package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	port        = flag.String("p", ":12345", "HTTP listen address")
	fetcherPort = flag.String("f", ":8000", "DFK Fetcher port")
)

func init() {
	flag.Parse()
}

// Config provides basic configuration
type Config struct {
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// HTMLServer represents the web service that serves up HTML
type HTMLServer struct {
	server *http.Server
	wg     sync.WaitGroup
}

// Start launches the HTML Server
func Start(cfg Config) *HTMLServer {
	flag.Parse()
	// Setup Context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Handlers
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`))
	})
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write([]byte("\n\t\tUser-agent: *\n\t\tAllow: /allowed\n\t\tDisallow: /disallowed\n\t\t"))
	})
	r.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("allowed"))
	})
	r.HandleFunc("/disallowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("disallowed"))
	})

	r.HandleFunc("/status/{status}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		st, err := strconv.Atoi(vars["status"])
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(st)
		w.Write([]byte(vars["status"]))
	})

	// Create the HTML Server
	htmlServer := HTMLServer{
		server: &http.Server{
			Addr:           cfg.Host,
			Handler:        r,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}

	// Add to the WaitGroup for the listener goroutine
	htmlServer.wg.Add(1)

	// Start the listener
	go func() {
		// fmt.Printf("\nProxy Server : Service started : Host=%v\n", htmlServer.server.Addr)
		// htmlServer.server.ListenAndServeTLS(
		// 	"/etc/letsencrypt/live/dataflowkit.org/fullchain.pem",
		// 	"/etc/letsencrypt/live/dataflowkit.org/privkey.pem",
		// )
		fmt.Printf("\nProxy Server : Service started : Host=%v\n", htmlServer.server.Addr)
		htmlServer.server.ListenAndServe()
		htmlServer.wg.Done()
	}()
	//redirect all requests from http to https
	// go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	// }))
	return &htmlServer
}

// Stop turns off the HTML Server
func (htmlServer *HTMLServer) Stop() error {
	// Create a context to attempt a graceful 5 second shutdown.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("\nTest Server : Service stopping\n")

	// Attempt the graceful shutdown by closing the listener
	// and completing all inflight requests
	if err := htmlServer.server.Shutdown(ctx); err != nil {
		// Looks like we timed out on the graceful shutdown. Force close.
		if err := htmlServer.server.Close(); err != nil {
			fmt.Printf("\nTest Server : Service stopping : Error=%v\n", err)
			return err
		}
	}
	// Wait for the listener to report that it is closed.
	htmlServer.wg.Wait()
	fmt.Printf("\nTest Server : Stopped\n")
	return nil
}

func main() {
	serverCfg := Config{
		Host:         *port, //"localhost:5000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	htmlServer := Start(serverCfg)
	defer htmlServer.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("main : shutting down")
}
