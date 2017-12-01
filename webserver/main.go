package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"gopkg.in/gin-gonic/gin.v1"
)

var (
	port          = flag.String("p", ":8080", "HTTP listen address")
	fetcherPort   = flag.String("f", ":8000", "Fetcher port")
	dfkParserPort = flag.String("d", ":8001", "DFK Parser port")
	baseDir       = flag.String("b", "web", "HTML files location.")
)


func main() {
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

	r.POST("/fetch/splash", ReverseProxy(fetcherPort))
	r.POST("/parse", ReverseProxy(dfkParserPort))
	r.Run(*port)
}

func ReverseProxy(p *string) gin.HandlerFunc {

	port := fmt.Sprintf("localhost%s", *p)

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			r := c.Request
			req = r
			req.URL.Scheme = "http"
			req.URL.Host = port
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
