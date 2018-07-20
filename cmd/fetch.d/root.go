// Copyright © 2017 Slotix s.r.o. <dm@slotix.sk>
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
	"fmt"
	"os"
	"os/signal"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/healthcheck"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	//VERSION               string // VERSION is set during build
	//  DFKFetch represents address of DFK Fetch service
	DFKFetch   string //Fetch service address
	fetchProxy string //Proxy address http://username:password@proxy-host:port
	chrome     string

	storageType     string
	ignoreCacheInfo bool
	diskvBaseDir    string

	cassandraHost string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dataflowkit",
	Short: "Dataflow Kit html fetcher",
	Long:  `Dataflow Kit fetch service retrieves html pages from websites and passes content to DFK parser service.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking services ... ")
		services := []healthcheck.Checker{
			healthcheck.ChromeConn{
				Host: chrome,
			},
		}
		if storageType == "Cassandra" {
			services = append(services, healthcheck.CassandraConn{
				Host: cassandraHost})
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
			fmt.Printf("Storage %s\n", storageType)
			fetchServer := viper.GetString("DFK_FETCH")
			serverCfg := fetch.Config{
				Host: fetchServer, //"localhost:5000",
			}
			htmlServer := fetch.Start(serverCfg)
			defer htmlServer.Stop()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			<-sigChan

			fmt.Println("main : shutting down")
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	VERSION = version
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	//flags and configuration settings. They are global for the application.

	RootCmd.Flags().StringVarP(&DFKFetch, "DFK_FETCH", "f", "127.0.0.1:8000", "HTTP listen address")
	RootCmd.Flags().StringVarP(&chrome, "CHROME", "c", "http://127.0.0.1:9222", "Headless Chrome URL address. It is used for fetching JS driven web pages")
	RootCmd.Flags().StringVarP(&fetchProxy, "PROXY", "p", "", "Proxy address http://username:password@proxy-host:port")

	//set here default type of storage
	RootCmd.Flags().StringVarP(&storageType, "STORAGE_TYPE", "", "Diskv", "Storage type. Types: Diskv, Cassandra")
	RootCmd.Flags().StringVarP(&diskvBaseDir, "DISKV_BASE_DIR", "", "diskv", "diskv base directory for storing fetch results")
	RootCmd.Flags().StringVarP(&cassandraHost, "CASSANDRA", "", "127.0.0.1", "Cassandra host address")

	if os.Getenv("DFK_FETCH") != "" {
		viper.Set("DFK_FETCH", os.Getenv("DFK_FETCH"))
	} else {
		viper.BindPFlag("DFK_FETCH", RootCmd.Flags().Lookup("DFK_FETCH"))
		//os.Setenv("DFK_FETCH", DFKFetch)
	}

	if os.Getenv("DISKV_BASE_DIR") != "" {
		//viper.BindEnv("DISKV_BASE_DIR")
		viper.Set("DISKV_BASE_DIR", os.Getenv("DISKV_BASE_DIR"))
	} else {
		viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	}

	viper.BindPFlag("PROXY", RootCmd.Flags().Lookup("PROXY"))

	viper.BindPFlag("STORAGE_TYPE", RootCmd.Flags().Lookup("STORAGE_TYPE"))
	viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	viper.BindPFlag("CASSANDRA", RootCmd.Flags().Lookup("CASSANDRA"))
}
