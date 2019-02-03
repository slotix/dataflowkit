package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/template"
	"github.com/gorilla/mux"
)

const PERSONS_PER_PAGE = 6

var (
	personsTpl      *template.Template
	indexTpl        *template.Template
	personDetailTpl *template.Template
	personTableTpl  *template.Template
	loginFormTpl    *template.Template
	citiesTpl       *template.Template
	countriesTpl    *template.Template
	personsV2       *template.Template
	errorTpl        *template.Template
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

// Person represents person info for test pages
type Person struct {
	Name    string          `json:"Name"`
	Phone   json.RawMessage `json:"Phone"`
	Email   string          `json:"Email"`
	Company string          `json:"Company"`
	Counter string          `json:"Counter"`
	Bio     string          `json:"Bio"`
	Country string          `json:"Country"`
	City    string          `json:"City"`
}

func init() {
	flag.Parse()
	// personsTpl = template.Must(template.ParseFiles(*baseDir+"/persons.html", *baseDir+"/base.html"))
	// personsTpl = personsTpl.Funcs(funcMap)
	personsTpl = template.Must(template.New(*baseDir+"/base.html").Funcs(funcMap).ParseFiles(*baseDir+"/persons.html", *baseDir+"/base.html"))
	personTableTpl = template.Must(template.New(*baseDir+"/base.html").Funcs(funcMap).ParseFiles(*baseDir+"/persons-table.html", *baseDir+"/base.html"))
	indexTpl = template.Must(template.ParseFiles(*baseDir+"/index.html", *baseDir+"/base.html"))
	personDetailTpl = template.Must(template.New(*baseDir+"/base.html").Funcs(funcMap).ParseFiles(*baseDir+"/p_detail.html", *baseDir+"/base.html"))
	loginFormTpl = template.Must(template.ParseFiles(*baseDir+"/base.html", *baseDir+"/login_form.html"))
	citiesTpl = template.Must(template.ParseFiles(*baseDir+"/base.html", *baseDir+"/cities.html"))
	countriesTpl = template.Must(template.ParseFiles(*baseDir+"/base.html", *baseDir+"/countries.html"))
	personsV2 = template.Must(template.New(*baseDir+"/base.html").Funcs(funcMap).ParseFiles(*baseDir+"/base.html", *baseDir+"/persons_v2.html"))
	errorTpl = template.Must(template.ParseFiles(*baseDir+"/base.html", *baseDir+"/error.html"))
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
		var user string
		identifyUser(&w, r, &user)
		w.Header().Set("Content-Type", "text/html")
		vars := map[string]interface{}{
			"Header": "Web Scraping Playground",
			"User":   user,
		}
		render(w, r, indexTpl, "base", vars)
	})
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var user string
		identifyUser(&w, r, &user)
		w.Header().Set("Content-Type", "text/html")
		vars := map[string]interface{}{
			"User": user,
		}
		if user != "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		if r.Method == "POST" {
			username := strings.Trim(r.FormValue("username"), " ")
			if username != "" {
				registerUser(&w, r, username)
			}
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		render(w, r, loginFormTpl, "base", vars)
	})
	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("sessionId"); err == nil {
			cookie.MaxAge = -1
			http.SetCookie(w, cookie)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`))
	})

	r.HandleFunc("/countries", countriesHandler)
	r.HandleFunc("/country/{country}", countryHandler)
	r.HandleFunc("/country/{country}/city/{city}", personsV2PagedHandler)
	r.HandleFunc("/country/{country}/city/{city}/page/{page}", personsV2PagedHandler)
	r.HandleFunc("/persons/page-{page}", personsHandler)
	r.HandleFunc("/persons/{id}", personDetailsHandler)
	r.HandleFunc("/persons-table", personTableHandler)
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
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

// Returns Person slice of required page with per_page elements
func getPagedPersons(p []Person, page int, per_page int) (result []Person, err error) {
	if page < 1 {
		return nil, fmt.Errorf("Page doesn't exist")
	}
	length := len(p)
	low := (page - 1) * per_page
	if length <= low {
		return nil, fmt.Errorf("Page doesn't exist")
	}
	high := low + per_page
	if high > length {
		high = length
	}
	return p[low:high], nil
}

func identifyUser(w *http.ResponseWriter, r *http.Request, user *string) {
	if cookie, err := r.Cookie("sessionId"); err != nil {
		return
	} else {
		if bUser, err := base64.StdEncoding.DecodeString(cookie.Value); err != nil {
			cookie.MaxAge = -1
			http.SetCookie(*w, cookie)
			return
		} else {
			*user = string(bUser)
		}
	}
}

func registerUser(w *http.ResponseWriter, r *http.Request, user string) {
	cookie := http.Cookie{
		Name:     "sessionId",
		Value:    base64.StdEncoding.EncodeToString([]byte(user)),
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
	}
	http.SetCookie(*w, &cookie)
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
	"inc":      Inc,
	"dec":      Dec,
	"mult":     Mult,
	"divide":   Div,
	"seqPages": SeqPages,
	"toString": ToStringSlice,
}

func SeqPages(totalPages int) (result []int) {
	for i := 1; i <= totalPages; i++ {
		result = append(result, i)
	}
	return result
}

// Dec decrease input value and return result as a string
func Dec(a int) string {
	return strconv.Itoa(a - 1)
}

// Inc increase input value and return result as a string
func Inc(a int) string {
	return strconv.Itoa(a + 1)
}

// Mult multiplies input values and return result as a string
func Mult(a, b int) string {
	return strconv.Itoa(a * b)
}

// Div divides a/b and return result as a string
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
	var user string
	identifyUser(&w, r, &user)
	w.Header().Set("Content-Type", "text/html")
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
		"User":         user,
	}
	render(w, r, personsTpl, "base", vars)
}

func personTableHandler(w http.ResponseWriter, r *http.Request) {
	var user string
	identifyUser(&w, r, &user)
	w.Header().Set("Content-Type", "text/html")
	vars := map[string]interface{}{
		"Header":  "Persons Table",
		"User":    user,
		"Persons": persons,
	}
	render(w, r, personTableTpl, "base", vars)
}

// ToStringSlice convert input data to string slice
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
	var user string
	identifyUser(&w, r, &user)
	w.Header().Set("Content-Type", "text/html")
	v := mux.Vars(r)
	id, exists := v["id"]
	if exists != true {
		fmt.Println("id is empty")
	}
	personIndex := -1
	for index, person := range persons {
		if person.Counter == id {
			personIndex = index
			break
		}
	}
	if personIndex != -1 {
		vars := map[string]interface{}{
			"Header":       "Person Details: " + persons[personIndex].Name,
			"Person":       persons[personIndex],
			"Phones":       ToStringSlice(persons[personIndex].Phone),
			"PersonsCount": personCount,
			"ItemsPerPage": "10",
			"User":         user,
		}
		render(w, r, personDetailTpl, "base", vars)
	} else {
		vars := map[string]interface{}{
			"Error": "Person not found!",
			"User":  user,
		}
		render(w, r, errorTpl, "base", vars)
	}
}

func countriesHandler(w http.ResponseWriter, r *http.Request) {
	var user string
	identifyUser(&w, r, &user)
	w.Header().Set("Content-Type", "text/html")
	countriesSet := make(map[string]interface{})
	for _, person := range persons {
		countriesSet[person.Country] = struct{}{}
	}
	countries := make([]string, 0, len(countriesSet))
	for country := range countriesSet {
		countries = append(countries, country)
	}
	//sort.Strings(countries)
	vars := map[string]interface{}{
		"User":      user,
		"Countries": countries,
	}
	render(w, r, countriesTpl, "base", vars)
}

func countryHandler(w http.ResponseWriter, r *http.Request) {
	var user string
	identifyUser(&w, r, &user)
	w.Header().Set("Content-Type", "text/html")
	v := mux.Vars(r)
	country := v["country"]
	citiesSet := make(map[string]interface{})
	for _, person := range persons {
		if person.Country == country {
			citiesSet[person.City] = struct{}{}
		}
	}
	cities := make([]string, 0)
	for city := range citiesSet {
		cities = append(cities, city)
	}
	if len(cities) != 0 {
		vars := map[string]interface{}{
			"User":    user,
			"Country": country,
			"Cities":  cities,
		}
		render(w, r, citiesTpl, "base", vars)
	} else {
		vars := map[string]interface{}{
			"User":  user,
			"Error": "No persons for the country " + country,
		}
		render(w, r, errorTpl, "base", vars)
	}
}

func personsV2PagedHandler(w http.ResponseWriter, r *http.Request) {
	var user string
	identifyUser(&w, r, &user)
	w.Header().Set("Content-Type", "text/html")
	v := mux.Vars(r)
	country := v["country"]
	city := v["city"]
	currentPage := 1
	if value, exists := v["page"]; exists {
		var err error
		currentPage, err = strconv.Atoi(value)
		if err != nil {
			vars := map[string]interface{}{
				"User":  user,
				"Error": "Wrong page index",
			}
			render(w, r, errorTpl, "base", vars)
			return
		}
	}
	personsSet := make([]Person, 0)
	for _, person := range persons {
		if person.City == city && person.Country == country {
			personsSet = append(personsSet, person)
		}
	}
	totalPages := int(math.Ceil(float64(len(personsSet)) / PERSONS_PER_PAGE))
	pagedPersons, err := getPagedPersons(personsSet, currentPage, PERSONS_PER_PAGE)
	if err != nil {
		vars := map[string]interface{}{
			"User":  user,
			"Error": err.Error(),
		}
		render(w, r, errorTpl, "base", vars)
		return
	}
	vars := map[string]interface{}{
		"User":         user,
		"Country":      country,
		"City":         city,
		"Persons":      pagedPersons,
		"TotalPages":   totalPages,
		"CurrentPage":  currentPage,
		"TotalPersons": len(personsSet),
	}
	render(w, r, personsV2, "base", vars)
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
