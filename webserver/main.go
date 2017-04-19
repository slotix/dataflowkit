package main

import (
	"flag"
	"fmt"
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"
)

var (
	port    = flag.String("p", ":8080", "HTTP listen address")
	baseDir = flag.String("d", "/web", "HTML files location.")
)

func main() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
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
	r.StaticFile("/favicon.ico", fmt.Sprintf("%s/favicon.ico",staticStr))
	
	r.LoadHTMLFiles(fmt.Sprintf("%s/index.html", *baseDir))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.Run(*port)
}
