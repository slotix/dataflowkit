package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/slotix/dataflowkit/splash"

	"github.com/go-kit/kit/log"
	"github.com/slotix/dataflowkit/fetch"
)

func main() {

	httpAddr := "127.0.0.1:8000"
	svc, err := fetch.NewHTTPClient(httpAddr, log.NewNopLogger())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	req := splash.Request{URL: "http://google.com"}
	resp, err := svc.Fetch(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(resp)
	fmt.Println(string(b))
}
