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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/slotix/dataflowkit/errs"
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
	Params   string
	Cookies  string
	LUA      string
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
				Host: viper.GetString("DFK_FETCH"),
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
			if URL == "" {
				fmt.Fprintf(os.Stderr, "error: %v\n", &errs.BadRequest{errors.New("no remote address specified")})
				os.Exit(1)
			}
			cx, cancel := context.WithCancel(context.Background())
			ch := make(chan error)

			go func() {
				svc, err := fetch.NewHTTPClient(viper.GetString("DFK_FETCH"))
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}

				var req fetch.FetchRequester
				switch strings.ToLower(fetcher) {
				case "splash":
					req = splash.Request{
						URL:     URL,
					//	FormData:  Params,
					//	Cookies: Cookies,
					//	LUA:     LUA,
					}
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
				select {
				case <-cx.Done():
					// Already timedout
				default:
					ch <- err
				}
			}()
			// Simulating user cancel request
			go func() {
				time.Sleep(10000 * time.Millisecond)
				cancel()
			}()
			select {
			case err := <-ch:
				if err != nil {
					// HTTP error
					panic(err)
				}
				print("no error")
			case <-cx.Done():
				fmt.Println(cx.Err())
			}
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
	RootCmd.Flags().StringVarP(&Params, "PARAMS", "", "", "Params is a string value for passing formdata parameters.")
	RootCmd.Flags().StringVarP(&Cookies, "COOKIES", "", "", "Cookies contain cookies to be added to request  before sending it to browser.")
	RootCmd.Flags().StringVarP(&LUA, "LUA", "", "", "LUA Splash custom script")

	//viper.AutomaticEnv() // read in environment variables that match
	//Environment variable takes precedence over flag value
	if os.Getenv("DFK_FETCH") != "" {
		viper.BindEnv("DFK_FETCH")
	} else {
		viper.BindPFlag("DFK_FETCH", RootCmd.Flags().Lookup("DFK_FETCH"))
	}
	
	viper.BindPFlag("FETCHER_TYPE", RootCmd.Flags().Lookup("FETCHER_TYPE"))
	viper.BindPFlag("URL", RootCmd.Flags().Lookup("URL"))
	viper.BindPFlag("PARAMS", RootCmd.Flags().Lookup("PARAMS"))
	viper.BindPFlag("COOKIES", RootCmd.Flags().Lookup("COOKIES"))
	viper.BindPFlag("LUA", RootCmd.Flags().Lookup("LUA"))

}
