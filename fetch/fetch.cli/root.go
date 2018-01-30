// Copyright © 2018 Slotix s.r.o. <dm@slotix.sk>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/healthcheck"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	DFKFetch string //DFKFetch service address.
	fetcher  string //fetcher type: splash, base
	URL      string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dataflowkit",
	Short: "DataFlow Kit html fetcher CLI",
	Long:  `Dataflow Kit Fetcher CLI downloads HTML content from specified URL`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking services ... ")

		services := []healthcheck.Checker{
			healthcheck.FetchConn{
				//Check if Splash Fetch service is alive
				Host: DFKFetch,
			},
		}

		status := healthcheck.CheckServices(services...)
		allAlive := true

		for k, v := range status {
			fmt.Printf("%s: %s\n", k, v)
			if v != "Ok" {
				allAlive = false
			}
		}
		if allAlive {
			svc, err := fetch.NewHTTPClient(DFKFetch, log.NewNopLogger())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			var req fetch.FetchRequester
			switch strings.ToLower(fetcher) {
			case "splash":
				req = splash.Request{URL: URL}
			case "base":
				req = fetch.BaseFetcherRequest{URL: URL}
			default:
				err := errors.New("invalid fetcher type specified")
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
		//	req := splash.Request{URL: URL}
			resp, err := svc.Fetch(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			b, err := ioutil.ReadAll(resp)
			fmt.Println(string(b))
		}
	},
}

func Execute(version string) {
	VERSION = version

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	//flags and configuration settings. They are global for the application.
	RootCmd.Flags().StringVarP(&DFKFetch, "DFK_FETCH", "f", "127.0.0.1:8000", "DFK Fetch service address")
	RootCmd.Flags().StringVarP(&fetcher, "FETCHER_TYPE", "t", "splash", "DFK Fetcher type: splash, base")
	RootCmd.Flags().StringVarP(&URL, "URL", "u", "", "URL to be fetched")

	viper.AutomaticEnv() // read in environment variables that match
	viper.BindPFlag("DFK_FETCH", RootCmd.Flags().Lookup("DFK_FETCH"))
	viper.BindPFlag("FETCHER_TYPE", RootCmd.Flags().Lookup("FETCHER_TYPE"))
	viper.BindPFlag("URL", RootCmd.Flags().Lookup("URL"))

}
