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
	"time"

	"github.com/slotix/dataflowkit/healthcheck"
	"github.com/slotix/dataflowkit/parse"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	DFKParse           string //DFKParse service address.
	DFKFetch           string //DFKFetch service address.
	fetcherType        string //fetcher type: splash, base
	storageType        string
	skipStorageMW      bool
	storageItemExpires int64 //how long in seconds object stay in a cache before expiration.
	diskvBaseDir       string

	spacesConfig   string //Digital Ocean spaces configuration file
	spacesEndpoint string //Digital Ocean spaces endpoint address
	DFKBucket      string //Bucket name for AWS S3 or DO Spaces

	redisHost       string
	redisExpire     int
	redisNetwork    string
	redisPassword   string
	redisDB         int
	redisSocketPath string

	maxPages            int
	format              string
	paginateResults     bool
	fetchDelay          int
	randomizeFetchDelay bool
	ignoreFetchDelay    bool
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
				//Check if Splash Fetch service is alive
				Host: viper.GetString("DFK_FETCH"),
			},
		}
		if storageType == "Redis" {
			services = append(services, healthcheck.RedisConn{
				Network: redisNetwork,
				Host:    redisHost})
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
				Host:         parseServer, //"localhost:5000",
				ReadTimeout:  60 * time.Second,
				WriteTimeout: 60 * time.Second,
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
	RootCmd.Flags().StringVarP(&fetcherType, "FETCHER_TYPE", "t", "base", "DFK Fetcher type: splash, base")
	//default type of storage
	RootCmd.Flags().BoolVarP(&skipStorageMW, "SKIP_STORAGE_MW", "", true, "If true no parsed data will be saved to storage. This flag forces parser to bypass storage middleware.")
	RootCmd.Flags().StringVarP(&storageType, "STORAGE_TYPE", "", "Diskv", "Storage backend for intermediary data passed to html parser. Types: S3, Spaces, Redis, Diskv")
	RootCmd.Flags().Int64VarP(&storageItemExpires, "ITEM_EXPIRE_IN", "", 3600, "Default value for item expiration in seconds")
	RootCmd.Flags().StringVarP(&diskvBaseDir, "DISKV_BASE_DIR", "", "diskv", "diskv base directory for storing fetch results")
	//RootCmd.Flags().StringVarP(&spacesConfig, "SPACES_CONFIG", "", os.Getenv("HOME")+"/"+".spaces/credentials", "Digital Ocean Spaces Configuration file")
	RootCmd.Flags().StringVarP(&spacesConfig, "SPACES_CONFIG", "",
		"$HOME/.spaces/credentials", "Digital Ocean Spaces Configuration file")
	RootCmd.Flags().StringVarP(&spacesEndpoint, "SPACES_ENDPOINT", "", "https://ams3.digitaloceanspaces.com", "Digital Ocean Spaces Endpoint Address")
	RootCmd.Flags().StringVarP(&DFKBucket, "DFK_BUCKET", "", "dfk-storage", "AWS S3 or Digital Ocean Spaces bucket name for storing parsed results")

	RootCmd.Flags().StringVarP(&redisHost, "REDIS", "r", "127.0.0.1:6379", "Redis host address")
	RootCmd.Flags().IntVarP(&redisExpire, "REDIS_EXPIRE", "", 3600, "Default Redis expire value in seconds")
	RootCmd.Flags().StringVarP(&redisNetwork, "REDIS_NETWORK", "", "tcp", "Redis Network")
	RootCmd.Flags().StringVarP(&redisPassword, "REDIS_PASSWORD", "", "", "Redis Password")
	RootCmd.Flags().IntVarP(&redisDB, "REDIS_DB", "", 0, "Redis DB")
	RootCmd.Flags().StringVarP(&redisSocketPath, "REDIS_SOCKET_PATH", "", "", "Redis Socket Path")

	RootCmd.Flags().IntVarP(&maxPages, "MAX_PAGES", "", 1, "The maximum number of pages to scrape")
	RootCmd.Flags().StringVarP(&format, "FORMAT", "", "json", "Output format (CSV, JSON, XML)")
	RootCmd.Flags().BoolVarP(&paginateResults, "PAGINATE_RESULTS", "", false, "Paginated results are returned. Single list of combined results from every block on all pages is returned by default.")
	RootCmd.Flags().IntVarP(&fetchDelay, "FETCH_DELAY", "", 500, "Specifies sleep time in milliseconds for multiple requests for the same domain.")
	RootCmd.Flags().BoolVarP(&randomizeFetchDelay, "RANDOMIZE_FETCH_DELAY", "", true, "RandomizeFetchDelay setting decreases the chance of a crawler being blocked. This way a random delay ranging from 0.5 * FetchDelay to 1.5 * FetchDelay seconds is used between consecutive requests to the same domain. If FetchDelay is zero this option has no effect.")
	RootCmd.Flags().BoolVarP(&ignoreFetchDelay, "IGNORE_FETCH_DELAY", "", true, "Ignores fetchDelay setting intended for debug purpose. Please set it to false in Production")

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

	if os.Getenv("FETCHER_TYPE") != "" {
		viper.Set("FETCHER_TYPE", os.Getenv("FETCHER_TYPE"))
	} else {
		viper.BindPFlag("FETCHER_TYPE", RootCmd.Flags().Lookup("FETCHER_TYPE"))
	}

	if os.Getenv("DISKV_BASE_DIR") != "" {
		//viper.BindEnv("DISKV_BASE_DIR")
		viper.Set("DISKV_BASE_DIR", os.Getenv("DISKV_BASE_DIR"))
	} else {
		viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	}
	//viper.BindPFlag("FETCHER_TYPE", RootCmd.Flags().Lookup("FETCHER_TYPE"))
	viper.BindPFlag("SKIP_STORAGE_MW", RootCmd.Flags().Lookup("SKIP_STORAGE_MW"))
	viper.BindPFlag("STORAGE_TYPE", RootCmd.Flags().Lookup("STORAGE_TYPE"))
	viper.BindPFlag("ITEM_EXPIRE_IN", RootCmd.Flags().Lookup("ITEM_EXPIRE_IN"))
	viper.BindPFlag("SPACES_CONFIG", RootCmd.Flags().Lookup("SPACES_CONFIG"))
	viper.BindPFlag("SPACES_ENDPOINT", RootCmd.Flags().Lookup("SPACES_ENDPOINT"))
	viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	viper.BindPFlag("DFK_BUCKET", RootCmd.Flags().Lookup("DFK_BUCKET"))
	viper.BindPFlag("REDIS", RootCmd.Flags().Lookup("REDIS"))
	viper.BindPFlag("REDIS_EXPIRE", RootCmd.Flags().Lookup("REDIS_EXPIRE"))
	viper.BindPFlag("REDIS_NETWORK", RootCmd.Flags().Lookup("REDIS_NETWORK"))
	viper.BindPFlag("REDIS_PASSWORD", RootCmd.Flags().Lookup("REDIS_PASSWORD"))
	viper.BindPFlag("REDIS_DB", RootCmd.Flags().Lookup("REDIS_DB"))
	viper.BindPFlag("REDIS_SOCKET_PATH", RootCmd.Flags().Lookup("REDIS_SOCKET_PATH"))

	viper.BindPFlag("MAX_PAGES", RootCmd.Flags().Lookup("MAX_PAGES"))
	viper.BindPFlag("FORMAT", RootCmd.Flags().Lookup("FORMAT"))
	viper.BindPFlag("PAGINATE_RESULTS", RootCmd.Flags().Lookup("PAGINATE_RESULTS"))
	viper.BindPFlag("FETCH_DELAY", RootCmd.Flags().Lookup("FETCH_DELAY"))
	viper.BindPFlag("RANDOMIZE_FETCH_DELAY", RootCmd.Flags().Lookup("RANDOMIZE_FETCH_DELAY"))
	viper.BindPFlag("IGNORE_FETCH_DELAY", RootCmd.Flags().Lookup("IGNORE_FETCH_DELAY"))

}
