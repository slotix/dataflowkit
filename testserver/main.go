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
	indexContent = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<style type="text/css">
	body { margin: 10px; }
	table, th, td, li, dl { font-family: "lucida grande", arial; font-size: 14pt; }
	dt { font-weight: bold; }
	table { background-color: #efefef; border: 2px solid #dddddd; width: 100%; }
	th { background-color: #efefef; }
	td { background-color: #ffffff; }
	</style>
</head>
<body>
<h1>Persons</h1>
<p>Warning! This is a demo website for web scraping purposes. The data on this page has been randomly generated.</p>
<table cellspacing="0" cellpadding="1">
<tr>
	<th>Name</th>
	<th>Phone</th>
	<th>Email</th>
	<th>Company</th>
</tr>
<tr>
	<td>Eagan C. Higgins</td>
	<td>158-9502</td>
	<td>sed.pede@sapien.ca</td>
	<td>Commodo At Company</td>
</tr>
<tr>
	<td>Ethan Wong</td>
	<td>740-7719</td>
	<td>at@et.edu</td>
	<td>Metus Inc.</td>
</tr>
<tr>
	<td>Quinn Haynes</td>
	<td>372-4289</td>
	<td>Sed.nulla@metusfacilisis.net</td>
	<td>Enim LLP</td>
</tr>
<tr>
	<td>Steel Frederick</td>
	<td>1-260-805-4413</td>
	<td>luctus@idnunc.co.uk</td>
	<td>Ante Nunc Mauris LLP</td>
</tr>
<tr>
	<td>Kasper Anthony</td>
	<td>611-8201</td>
	<td>sit.amet.nulla@non.edu</td>
	<td>Mus Limited</td>
</tr>
<tr>
	<td>Tallulah Nieves</td>
	<td>165-3303</td>
	<td>nascetur@inceptoshymenaeosMauris.net</td>
	<td>Duis Associates</td>
</tr>
<tr>
	<td>Lydia Whitfield</td>
	<td>1-249-695-8401</td>
	<td>sit.amet.orci@semperduilectus.ca</td>
	<td>Praesent Consulting</td>
</tr>
<tr>
	<td>Raven C. Gaines</td>
	<td>100-9381</td>
	<td>Pellentesque@egestasadui.com</td>
	<td>Aliquet Sem Associates</td>
</tr>
<tr>
	<td>Julie Zimmerman</td>
	<td>1-380-382-8144</td>
	<td>lectus.justo@Integer.org</td>
	<td>Eu Consulting</td>
</tr>
<tr>
	<td>Moses D. Hubbard</td>
	<td>1-474-770-2793</td>
	<td>sagittis.semper.Nam@cursusluctus.net</td>
	<td>Vivamus Corp.</td>
</tr>

</table>

</body>
</html>`
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
		w.Write([]byte(indexContent))
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

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alive": true}`))
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
