package main

import "fmt"

var (
	// BuildTime is a time label of the moment when the binary was built
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built
	Commit = "unset"
	// Release is a semantic version of current build
	Release = "unset"
	Version = "unset"
)

func main() {
	Version = fmt.Sprintf("Dataflow Kit fetcher\n Release: %s\n Commit: %s\n Build time: %s", Release, Commit, BuildTime)
	Execute()
}
