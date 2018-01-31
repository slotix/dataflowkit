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
	"log"
	"os"
	"os/user"

	"github.com/slotix/dataflowkit/fetch"
	"github.com/slotix/dataflowkit/healthcheck"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	//VERSION               string // VERSION is set during build
	//  DFKFetch represents address of DFK Fetch service
	DFKFetch string //Fetch service address

	splashHost            string
	splashTimeout         int
	splashResourceTimeout int
	splashWait            float64

	storageType     string
	skipStorageMW   bool
	ignoreCacheInfo bool
	diskvBaseDir    string

	//Digital Ocean spaces configuration file
	spacesConfig string
	//Digital Ocean spaces endpoint address
	spacesEndpoint string
	s3Region       string
	//Bucket name for AWS S3 or DO Spaces
	DFKBucket string

	redisHost       string
	redisExpire     int
	redisNetwork    string
	redisPassword   string
	redisDB         int
	redisSocketPath string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dataflowkit",
	Short: "Dataflow Kit html fetcher",
	Long:  `Dataflow Kit fetch service retrieves html pages from websites and passes content to DFK parser service.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking services ... ")
		services := []healthcheck.Checker{
			healthcheck.SplashConn{
				Host: splashHost,
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
			fmt.Printf("Starting Server %s\n", DFKFetch)
			fetch.Start(DFKFetch)
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

	RootCmd.Flags().StringVarP(&DFKFetch, "DFK_FETCH", "a", "127.0.0.1:8000", "HTTP listen address")

	RootCmd.Flags().StringVarP(&splashHost, "SPLASH", "s", "127.0.0.1:8050", "Splash host address")
	RootCmd.Flags().IntVarP(&splashTimeout, "SPLASH_TIMEOUT", "", 20, "Timeout (in seconds) for the render.")
	RootCmd.Flags().IntVarP(&splashResourceTimeout, "SPLASH_RESOURCE_TIMEOUT", "", 30, "A timeout (in seconds) for individual network requests.")
	RootCmd.Flags().Float64VarP(&splashWait, "SPLASH_WAIT", "", 0.5, "Time in seconds to wait until js scripts loaded.")

	//set here default type of storage
	RootCmd.Flags().StringVarP(&storageType, "STORAGE_TYPE", "", "Diskv", "Storage backend for intermediary data passed to html parser. Types: S3, Spaces, Redis, Diskv")
	//RootCmd.Flags().Int64VarP(&storageItemExpires, "ITEM_EXPIRE_IN", "", 3600, "Default value for item expiration in seconds")
	RootCmd.Flags().BoolVarP(&skipStorageMW, "SKIP_STORAGE_MW", "", false, "If true no data will be saved to storage. This flag forces fetcher to bypass storage middleware.")
	RootCmd.Flags().BoolVarP(&ignoreCacheInfo, "IGNORE_CACHE_INFO", "", true, "If a website is not cachable by some reason, ignore this and use cached copy if any. Please don't set it to true in production")
	RootCmd.Flags().StringVarP(&diskvBaseDir, "DISKV_BASE_DIR", "", "diskv", "diskv base directory for storing fetch results")
	RootCmd.Flags().StringVarP(&spacesConfig, "SPACES_CONFIG", "", homeDir()+".spaces/credentials", "Digital Ocean Spaces Configuration file")
	RootCmd.Flags().StringVarP(&spacesEndpoint, "SPACES_ENDPOINT", "", "https://ams3.digitaloceanspaces.com", "Digital Ocean Spaces Endpoint Address")
	RootCmd.Flags().StringVarP(&s3Region, "S3_REGION", "", "us-east-1", "AWS S3 or Digital Ocean Spaces region")
	RootCmd.Flags().StringVarP(&DFKBucket, "DFK_BUCKET", "", "dfk-storage", "AWS S3 or Digital Ocean Spaces bucket name for storing fetch results")

	RootCmd.Flags().StringVarP(&redisHost, "REDIS", "r", "127.0.0.1:6379", "Redis host address")
	RootCmd.Flags().IntVarP(&redisExpire, "REDIS_EXPIRE", "", 3600, "Default Redis expire value in seconds")
	RootCmd.Flags().StringVarP(&redisNetwork, "REDIS_NETWORK", "", "tcp", "Redis Network")
	RootCmd.Flags().StringVarP(&redisPassword, "REDIS_PASSWORD", "", "", "Redis Password")
	RootCmd.Flags().IntVarP(&redisDB, "REDIS_DB", "", 0, "Redis DB")
	RootCmd.Flags().StringVarP(&redisSocketPath, "REDIS_SOCKET_PATH", "", "", "Redis Socket Path")

	viper.AutomaticEnv() // read in environment variables that match
	viper.BindPFlag("DFK_FETCH", RootCmd.Flags().Lookup("DFK_FETCH"))

	viper.BindPFlag("SPLASH", RootCmd.Flags().Lookup("SPLASH"))
	viper.BindPFlag("SPLASH_TIMEOUT", RootCmd.Flags().Lookup("SPLASH_TIMEOUT"))
	viper.BindPFlag("SPLASH_RESOURCE_TIMEOUT", RootCmd.Flags().Lookup("SPLASH_RESOURCE_TIMEOUT"))
	viper.BindPFlag("SPLASH_WAIT", RootCmd.Flags().Lookup("SPLASH_WAIT"))

	viper.BindPFlag("STORAGE_TYPE", RootCmd.Flags().Lookup("STORAGE_TYPE"))
	//viper.BindPFlag("ITEM_EXPIRE_IN", RootCmd.Flags().Lookup("ITEM_EXPIRE_IN"))
	viper.BindPFlag("SKIP_STORAGE_MW", RootCmd.Flags().Lookup("SKIP_STORAGE_MW"))
	viper.BindPFlag("IGNORE_CACHE_INFO", RootCmd.Flags().Lookup("IGNORE_CACHE_INFO"))
	viper.BindPFlag("SPACES_CONFIG", RootCmd.Flags().Lookup("SPACES_CONFIG"))
	viper.BindPFlag("SPACES_ENDPOINT", RootCmd.Flags().Lookup("SPACES_ENDPOINT"))
	viper.BindPFlag("DISKV_BASE_DIR", RootCmd.Flags().Lookup("DISKV_BASE_DIR"))
	viper.BindPFlag("S3_REGION", RootCmd.Flags().Lookup("S3_REGION"))
	viper.BindPFlag("DFK_BUCKET", RootCmd.Flags().Lookup("DFK_BUCKET"))
	viper.BindPFlag("REDIS", RootCmd.Flags().Lookup("REDIS"))
	viper.BindPFlag("REDIS_EXPIRE", RootCmd.Flags().Lookup("REDIS_EXPIRE"))
	viper.BindPFlag("REDIS_NETWORK", RootCmd.Flags().Lookup("REDIS_NETWORK"))
	viper.BindPFlag("REDIS_PASSWORD", RootCmd.Flags().Lookup("REDIS_PASSWORD"))
	viper.BindPFlag("REDIS_DB", RootCmd.Flags().Lookup("REDIS_DB"))
	viper.BindPFlag("REDIS_SOCKET_PATH", RootCmd.Flags().Lookup("REDIS_SOCKET_PATH"))

}

//homeDir returns user's $HOME directory
func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir + "/"
}
