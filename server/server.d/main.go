package main

import (
	"fmt"
)

var VERSION = "0.1"
var buildTime = "No buildstamp"

func main() {
	version := fmt.Sprintf("%s\nBuild time: %s\n", VERSION, buildTime)
	Execute(fmt.Sprintf(version))
	/*
		viper.Set("SPLASH", "127.0.0.1:8050")
		viper.Set("SPLASH_TIMEOUT", "20")
		viper.Set("SPLASH_RESOURCE_TIMEOUT", "30")
		viper.Set("SPLASH_WAIT", "0,5")
		viper.Set("redis", "127.0.0.1:6379")
		viper.Set("redis-expire", "3600")
		viper.Set("redis-network", "tcp")
	*/

}
