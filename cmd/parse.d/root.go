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
	"strings"

	"github.com/slotix/dataflowkit/healthcheck"
	"github.com/slotix/dataflowkit/parse"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//flag vars
var (
	DFKParse           string
	DFKFetch           string //DFKFetch service address.
	storageType        string
	skipStorageMW      bool
	storageItemExpires int64 //how long in seconds object stay in a cache before expiration.
	diskvBaseDir       string
	resultsDir         string

	cassandraHost string
	mongoHost     string

	maxPages            int
	paginateResults     bool
	fetchDelay          int
	randomizeFetchDelay bool
	ignoreFetchDelay    bool

	fetchChannelSize int
	blockChannelSize int
	fetchWorkerNum   int
	blockWorkerNum   int
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dataflowkit",
	Short: "DataFlow Kit html parser",
	Long:  `Dataflow kit is a web scraping tool for structured data extraction. It follows the specified extractors described in JSON file and returns parsed data as CSV, JSON or XML data.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking services ... ")

		services := []healthcheck.Checker{
			healthcheck.FetchConn{
				//Check if Chrome Fetch service is alive
				Host: viper.GetString("DFK_FETCH"),
			},
		}
		sType := strings.ToLower(storageType)
		if sType == "cassandra" {
			services = append(services, healthcheck.CassandraConn{
				Host: cassandraHost,
			})
		}
		if sType == "mongodb" {
			services = append(services, healthcheck.MongoConn{
				Host: mongoHost,
			})
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
			if skipStorageMW {
				fmt.Printf("Storage %s\n", "None")
			} else {
				fmt.Printf("Storage %s\n", storageType)
			}
			parseServer := viper.GetString("DFK_PARSE")
			serverCfg := parse.Config{
				Host: parseServer, //"localhost:5000",
			}
			htmlServer := parse.Start(serverCfg)
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

	RootCmd.Flags().StringVarP(&DFKParse, "DFK_PARSE", "p", "127.0.0.1:8001", "HTTP listen address")
	RootCmd.Flags().StringVarP(&DFKFetch, "DFK_FETCH", "f", "127.0.0.1:8000", "DFK Fetch service address")
	RootCmd.Flags().StringVarP(&storageType, "STORAGE_TYPE", "", "Diskv", "Storage backend for intermediary data passed to html parser. Types: Diskv, MongoDB, Cassandra")
	RootCmd.Flags().StringVarP(&resultsDir, "RESULTS_DIR", "", "results", "Directory for storing results")
	RootCmd.Flags().Int64VarP(&storageItemExpires, "ITEM_EXPIRE_IN", "", 86400, "Default value for item expiration in seconds")
	RootCmd.Flags().StringVarP(&diskvBaseDir, "DISKV_BASE_DIR", "", "diskv", "diskv base directory for storing fetch results")
	RootCmd.Flags().StringVarP(&cassandraHost, "CASSANDRA", "c", "127.0.0.1", "Cassandra host address")
	RootCmd.Flags().StringVarP(&mongoHost, "MONGO", "", "127.0.0.1", "MongoDB host address")

	RootCmd.Flags().IntVarP(&maxPages, "MAX_PAGES", "", 10, "The maximum number of pages to scrape")
	RootCmd.Flags().BoolVarP(&paginateResults, "PAGINATE_RESULTS", "", false, "Paginated results are returned. Single list of combined results from every block on all pages is returned by default.")
	RootCmd.Flags().IntVarP(&fetchDelay, "FETCH_DELAY", "", 500, "Specifies sleep time in milliseconds for multiple requests for the same domain.")
	RootCmd.Flags().BoolVarP(&randomizeFetchDelay, "RANDOMIZE_FETCH_DELAY", "", true, "RandomizeFetchDelay setting decreases the chance of a crawler being blocked. This way a random delay ranging from 0.5 * FetchDelay to 1.5 * FetchDelay seconds is used between consecutive requests to the same domain. If FetchDelay is zero this option has no effect.")
	RootCmd.Flags().BoolVarP(&ignoreFetchDelay, "IGNORE_FETCH_DELAY", "", false, "Ignores fetchDelay setting intended for debug purpose. Please set it to false in Production")

	RootCmd.Flags().IntVar(&fetchChannelSize, "FETCH_CHANNEL_SIZE", 60, "The size of fetcher pool")
	RootCmd.Flags().IntVar(&fetchWorkerNum, "FETCH_WORKER_NUM", 60, "The number of fetcher workers")
	RootCmd.Flags().IntVar(&blockChannelSize, "BLOCK_CHANNEL_SIZE", 50, "The size of block pool")
	RootCmd.Flags().IntVar(&blockWorkerNum, "BLOCK_WORKER_NUM", 50, "The number of block workers")

	//viper.AutomaticEnv() // read in environment variables that match

	//Environment variable takes precedence over flag value
	if os.Getenv("DFK_PARSE") != "" {
		//viper.BindEnv("DFK_PARSE")
		viper.Set("DFK_PARSE", os.Getenv("DFK_PARSE"))

	} else {
		viper.BindPFlag("DFK_PARSE", RootCmd.Flags().Lookup("DFK_PARSE"))
	}
	if os.Getenv("DFK_FETCH") != "" {
		viper.Set("DFK_FETCH", os.Getenv("DFK_FETCH"))
	} else {
		viper.BindPFlag("DFK_FETCH", RootCmd.Flags().Lookup("DFK_FETCH"))
	}

	if os.Getenv("DISKV_BASE_DIR") != "" {
		//viper.BindEnv("DISKV_BASE_DIR")
		viper.Set("DISKV_BASE_DIR", os.Getenv("DISKV_BASE_DIR"))
	} else {
		viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	}

	if os.Getenv("MONGO") != "" {
		viper.Set("MONGO", os.Getenv("MONGO"))
	} else {
		viper.BindPFlag("MONGO", RootCmd.Flags().Lookup("MONGO"))
	}

	viper.BindPFlag("RESULTS_DIR", RootCmd.Flags().Lookup("RESULTS_DIR"))
	viper.BindPFlag("STORAGE_TYPE", RootCmd.Flags().Lookup("STORAGE_TYPE"))
	viper.BindPFlag("ITEM_EXPIRE_IN", RootCmd.Flags().Lookup("ITEM_EXPIRE_IN"))
	viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	viper.BindPFlag("CASSANDRA", RootCmd.Flags().Lookup("CASSANDRA"))
	viper.BindPFlag("MONGO", RootCmd.Flags().Lookup("MONGO"))

	viper.BindPFlag("MAX_PAGES", RootCmd.Flags().Lookup("MAX_PAGES"))
	viper.BindPFlag("PAGINATE_RESULTS", RootCmd.Flags().Lookup("PAGINATE_RESULTS"))
	viper.BindPFlag("FETCH_DELAY", RootCmd.Flags().Lookup("FETCH_DELAY"))
	viper.BindPFlag("RANDOMIZE_FETCH_DELAY", RootCmd.Flags().Lookup("RANDOMIZE_FETCH_DELAY"))
	viper.BindPFlag("IGNORE_FETCH_DELAY", RootCmd.Flags().Lookup("IGNORE_FETCH_DELAY"))

	viper.BindPFlag("FETCH_CHANNEL_SIZE", RootCmd.Flags().Lookup("FETCH_CHANNEL_SIZE"))
	viper.BindPFlag("FETCH_WORKER_NUM", RootCmd.Flags().Lookup("FETCH_WORKER_NUM"))
	viper.BindPFlag("BLOCK_CHANNEL_SIZE", RootCmd.Flags().Lookup("BLOCK_CHANNEL_SIZE"))
	viper.BindPFlag("BLOCK_WORKER_NUM", RootCmd.Flags().Lookup("BLOCK_WORKER_NUM"))
}
