package main

import (
	"fmt"

	"github.com/slotix/dfk-parser/cmd"


)

var VERSION = "0.1"
var buildTime = "No buildstamp"

//var githash = "No githash"

func main() {
	version := fmt.Sprintf("%s\nBuild time: %s\n", VERSION, buildTime)
	cmd.Execute(fmt.Sprintf(version))

}