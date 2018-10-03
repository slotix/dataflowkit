package main

import (
	"fmt"
)

//VERSION represents the current version of the service
var VERSION = "0.8"
var buildTime = "No buildstamp"

func main() {
	version := fmt.Sprintf("%s\nBuild time: %s\n", VERSION, buildTime)
	Execute(fmt.Sprintf(version))
}
