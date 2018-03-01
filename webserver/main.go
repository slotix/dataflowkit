package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	port        = flag.String("p", ":8080", "HTTP listen address")
	fetcherPort = flag.String("f", ":8000", "DFK Fetcher port")
	parserPort  = flag.String("d", ":8001", "DFK Parser port")
	baseDir     = flag.String("b", "../web", "HTML files location.")
)

func helpHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./../web/get_started.html")
}

func reverseProxy(p *string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		host := "localhost" + *p
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   host,
		})
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	flag.Parse()
	//http.HandleFunc("/favicon.ico", faviconHandler)
	http.Handle("/", http.FileServer(http.Dir("./../web")))
	http.HandleFunc("/get_started", helpHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./../web/static"))))

	// http.HandleFunc("/fetch/splash", func(w http.ResponseWriter, r *http.Request) {
	// 	host := "localhost" + *fetcherPort
	// 	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
	// 		Scheme: "http",
	// 		//Host:   "localhost:8000",
	// 		Host: host,
	// 	})
	// 	proxy.ServeHTTP(w, r)
	// })
	http.HandleFunc("/fetch/splash", reverseProxy(fetcherPort))
	http.HandleFunc("/fetch/base", reverseProxy(fetcherPort))
	http.HandleFunc("/parse", reverseProxy(parserPort))
	fmt.Println("starting at port ", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}

/* func main() {
	flag.Parse()
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	staticStr := fmt.Sprintf("%s/static", *baseDir)
	r.Static("/static", staticStr)
	r.StaticFile("/favicon.ico", fmt.Sprintf("%s/favicon.ico", staticStr))

	r.LoadHTMLFiles(fmt.Sprintf("%s/index.html", *baseDir),
		fmt.Sprintf("%s/get_started.html", *baseDir))
	r.GET("/", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"index.html",
			nil)
	})
	r.GET("/get_started", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"get_started.html",
			nil)
	})

	r.POST("/fetch/splash", reverseProxy(fetcherPort))
	r.POST("/fetch/base", reverseProxy(fetcherPort))
	r.POST("/parse", reverseProxy(dfkParserPort))
	r.Run(*port)
} */

/* func reverseProxy(p *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := fmt.Sprintf("localhost%s", *p)
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			//Host:   "localhost:8000",
			Host: host,
		})
		proxy.ServeHTTP(c.Writer, c.Request)
	}
} */

/* func ReverseProxy(p *string) gin.HandlerFunc {

	host := fmt.Sprintf("localhost%s", *p)

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			r := c.Request
			req = r
			req.URL.Scheme = "http"
			req.URL.Host = host

		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
} */
