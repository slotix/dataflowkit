package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/alecthomas/template"
	"github.com/gorilla/mux"
)

var (
	personsTpl      *template.Template
	indexTpl        *template.Template
	personDetailTpl *template.Template
	personTableTpl  *template.Template
	personsJSON     string
	persons         []Person
	port            = flag.String("p", ":12345", "HTTP(S) listen address")
	https           = flag.Bool("s", false, "expect HTTPS connections")
	certFile        = flag.String("c", "", "certificate file path")
	keyFile         = flag.String("k", "", "private key file path")
	fetcherPort     = flag.String("f", ":8000", "DFK Fetcher port")
	baseDir         = flag.String("b", "./web", "HTML files location.")
	personCount     int
)

type Person struct {
	Name    string          `json:"Name"`
	Phone   json.RawMessage `json:"Phone"`
	Email   string          `json:"Email"`
	Company string          `json:"Company"`
	Counter string          `json:"Counter"`
	Bio     string          `json:"Bio"`
}

func init() {
	flag.Parse()
	// personsTpl = template.Must(template.ParseFiles(*baseDir+"/persons.html", *baseDir+"/base.html"))
	// personsTpl = personsTpl.Funcs(funcMap)
	personsTpl = template.Must(template.New(*baseDir+"/base.html").Funcs(funcMap).ParseFiles(*baseDir+"/persons.html", *baseDir+"/base.html"))
	personTableTpl = template.Must(template.ParseFiles(*baseDir+"/persons-table.html", *baseDir+"/base.html"))
	indexTpl = template.Must(template.ParseFiles(*baseDir+"/index.html", *baseDir+"/base.html"))
	personDetailTpl = template.Must(template.New(*baseDir+"/base.html").Funcs(funcMap).ParseFiles(*baseDir+"/p_detail.html", *baseDir+"/base.html"))

	dat, err := ioutil.ReadFile(*baseDir + "/data/persons.json")
	if err != nil {
		fmt.Println(err)
	}
	personsJSON = string(dat)

	if err := json.Unmarshal(dat, &persons); err != nil {
		panic(err)
	}
	personCount = len(persons)
}

// Config provides basic configuration
type Config struct {
	Host     string
	HTTPS    bool //expect HTTPS connections
	KeyFile  string
	CertFile string
	// ReadTimeout  time.Duration
	// WriteTimeout time.Duration
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
		vars := map[string]interface{}{
			"Header": "Web Scraping Playground",
		}
		render(w, r, indexTpl, "base", vars)
	})
	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`))
	})
	r.HandleFunc("/persons/page-{page}", personsHandler)
	r.HandleFunc("/persons/{id}", personDetailsHandler)
	r.HandleFunc("/persons-table", personTableHandler)
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Conent-Type", "text/html")
		w.Write([]byte("\n\t\tUser-agent: *\n\t\tAllow: /allowed\n\t\tDisallow: /disallowed\n\t\tDisallow: /redirect\n\t\t"))
	})
	r.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("allowed"))
	})
	r.HandleFunc("/disallowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("disallowed"))
	})

	//handle redirects
	r.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirected", http.StatusMovedPermanently)
	})

	r.HandleFunc("/redirected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Redirected"))
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
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(*baseDir+"/static"))))

	// Create the HTML Server
	htmlServer := HTMLServer{
		server: &http.Server{
			Addr:    cfg.Host,
			Handler: r,
			// ReadTimeout:    cfg.ReadTimeout,
			// WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}

	// Add to the WaitGroup for the listener goroutine
	htmlServer.wg.Add(1)

	if !cfg.HTTPS {
		// Start HTTP listener
		go func() {
			fmt.Printf("\nTest Server : Service started : Host=%v\n", htmlServer.server.Addr)
			htmlServer.server.ListenAndServe()
			htmlServer.wg.Done()
		}()
	} else {
		// Start HTTPS listener
		go func() {
			fmt.Printf("\nTest Server : Service started : Host=%v\n", htmlServer.server.Addr)
			htmlServer.server.ListenAndServeTLS(
				*certFile,
				*keyFile,
			)
			htmlServer.wg.Done()
		}()

		//redirect all requests from http to https
		go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		}))
	}
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
		Host: *port, //"localhost:12345",
		// ReadTimeout:  60 * time.Second,
		// WriteTimeout: 60 * time.Second,
		HTTPS: *https,
	}
	if serverCfg.HTTPS {
		serverCfg.Host = ":443"
		if len(*certFile) == 0 || len(*keyFile) == 0 {
			log.Fatal(errors.New("No certificate file or private key file specified"))
		}

	}
	htmlServer := Start(serverCfg)
	defer htmlServer.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("main : shutting down")
}

var funcMap = template.FuncMap{
	"inc":    Inc,
	"dec":    Dec,
	"mult":   Mult,
	"divide": Div,
}

func Dec(a int) string {
	return strconv.Itoa(a - 1)
}

func Inc(a int) string {
	return strconv.Itoa(a + 1)
}

func Mult(a, b int) string {
	return strconv.Itoa(a * b)
}

func Div(a, b string) string {
	intA, err := strconv.Atoi(a)
	if err != nil {
		return ""
	}
	intB, err := strconv.Atoi(b)
	if err != nil {
		return ""
	}

	return strconv.Itoa(intA / intB)
}

func personsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Conent-Type", "text/html")
	v := mux.Vars(r)
	page, err := strconv.Atoi(v["page"])
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Print(string(dat))
	vars := map[string]interface{}{
		"Header":       "Person Cards",
		"Data":         string(personsJSON),
		"Page":         page,
		"PersonsCount": personCount,
		"ItemsPerPage": 10,
	}
	render(w, r, personsTpl, "base", vars)
}

func personTableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Conent-Type", "text/html")
	//v := mux.Vars(r)
	// page, err := strconv.Atoi(v["page"])
	// if err != nil {
	// 	fmt.Println(err)
	// }
	vars := map[string]interface{}{
		"Header": "Persons Table",
		//"Data":         string(personsJSON),
		//"Page":         page,
		//"PersonsCount": personCount,
		//"ItemsPerPage": 10,
	}
	render(w, r, personTableTpl, "base", vars)
}

func ToStringSlice(data []byte) []string {
	var v []string
	err := json.Unmarshal(data, &v)
	if err != nil {
		var s string
		err := json.Unmarshal(data, &s)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		v = append(v, s)
	}
	return v
}

func personDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Conent-Type", "text/html")
	v := mux.Vars(r)
	id, err := strconv.Atoi(v["id"])
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Print(string(dat))
	vars := map[string]interface{}{
		"Header":       "Person Details: " + persons[id].Name,
		"Person":       persons[id],
		"Phones":       ToStringSlice(persons[id].Phone),
		"PersonsCount": personCount,
		"ItemsPerPage": "10",
	}
	render(w, r, personDetailTpl, "base", vars)
}

// Render a template, or server error.
func render(w http.ResponseWriter, r *http.Request, tpl *template.Template, name string, data interface{}) {
	buf := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Write(buf.Bytes())
}
