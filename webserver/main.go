package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"

	"gopkg.in/gin-gonic/gin.v1"
)

var (
	port    = flag.String("p", ":8080", "HTTP listen address")
	dfkPort = flag.String("d", ":8000", "DataFlow kit port")
	baseDir = flag.String("b", "/web", "HTML files location.")
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

	r.LoadHTMLFiles(fmt.Sprintf("%s/index.html", *baseDir))
	r.GET("/", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"index.html",
			nil)
	})
	r.POST("/app/fetch", ReverseProxy())
	r.POST("/app/parse", ReverseProxy())
	r.Run(*port)
}

func ReverseProxy() gin.HandlerFunc {

	target := fmt.Sprintf("localhost%s",*dfkPort)

	//target := "localhost:8000"

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			r := c.Request
			req = r
			req.URL.Scheme = "http"
			req.URL.Host = target
			req.Header["my-header"] = []string{r.Header.Get("my-header")}
			// Golang camelcases headers
			delete(req.Header, "My-Header")
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}


