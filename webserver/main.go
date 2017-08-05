package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"

	"gopkg.in/gin-gonic/gin.v1"
	//"github.com/slotix/dataflowkit/cmd"
)

var (
	port          = flag.String("p", ":8080", "HTTP listen address")
	fetcherPort   = flag.String("f", ":8000", "Fetcher port")
	dfkParserPort = flag.String("d", ":8001", "DFK Parser port")
	baseDir       = flag.String("b", "web", "HTML files location.")
)

//var VERSION = "0.1"
//var buildTime = "No buildstamp"

//var githash = "No githash"
/*
func init() {
	viper.Set("splash", "127.0.0.1:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")
	viper.Set("redis", "127.0.0.1:6379")
	viper.Set("redis-expire", "3600")
	viper.Set("redis-network", "tcp")
}
*/

func main() {
	//version := fmt.Sprintf("%s\nBuild time: %s\n", VERSION, buildTime)
	//cmd.Execute(fmt.Sprintf(version))

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

	r.POST("/app/fetch", ReverseProxy(fetcherPort))
	r.POST("/app/parse", ReverseProxy(dfkParserPort))
	r.Run(*port)
}

func ReverseProxy(t *string) gin.HandlerFunc {

	target := fmt.Sprintf("localhost%s", *t)

	return func(c *gin.Context) {
		director := func(req *http.Request) {
			r := c.Request
			req = r
			req.URL.Scheme = "http"
			req.URL.Host = target
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
