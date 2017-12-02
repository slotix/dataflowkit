package main

import (
	"fmt"
)

var VERSION = "0.1"
var buildTime = "No buildstamp"

func main() {
	version := fmt.Sprintf("%s\nBuild time: %s\n", VERSION, buildTime)
	Execute(fmt.Sprintf(version))
}
