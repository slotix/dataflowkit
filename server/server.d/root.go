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

	"github.com/slotix/dataflowkit/healthcheck"
	"github.com/slotix/dataflowkit/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	//VERSION               string // VERSION is set during build
	port                  string
	proxy                 string
	redisHost             string
	redisExpire           int
	redisNetwork          string
	splashHost            string
	splashTimeout         int
	splashResourceTimeout int
	splashWait            float64
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dataflowkit",
	Short: "DataFlow Kit html parser",
	Long: `DataFlow Kit html parser serves for scraping data from websites according to chosen css selectors.
	Here is an example of payload structure:
	
	{"format":"json",
		"collections": [
				{
				"name": "collection1",
				"url": "http://example1.com",
				"fields": [
					{
						"field_name": "link",
						"css_selector": ".link a"
					},
					{
						"field_name": "Text",
						"css_selector": ".text"
					},
					{
						"field_name": "Image",
						"css_selector": ".foto img"
					}
				]
			}
		]
	}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking services ... ")

		status := healthcheck.CheckServices()
		allAlive := true

		for k, v := range status {
			fmt.Printf("%s: %s\n", k, v)
			if v != "Ok" {
				allAlive = false
			}
		}

		if allAlive {
			fmt.Printf("Starting Server %s\n", port)
			server.Start(port)
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
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.Flags().StringVarP(&port, "listen", "l", ":8000", "HTTP listen address")
	//RootCmd.Flags().StringVarP(&proxy, "proxy", "p", "", "Optional comma-separated list of URLs to proxy uppercase requests")
	RootCmd.Flags().StringVarP(&redisHost, "REDIS", "r", "127.0.0.1:6379", "Redis host address")
	RootCmd.Flags().IntVarP(&redisExpire, "REDIS_EXPIRE", "", 3600, "Default Redis expire value")
	RootCmd.Flags().StringVarP(&redisNetwork, "REDIS_NETWORK", "", "tcp", "Redis Network")
	RootCmd.Flags().StringVarP(&splashHost, "SPLASH", "s", "127.0.0.1:8050", "Splash host address")
	RootCmd.Flags().IntVarP(&splashTimeout, "SPLASH_TIMEOUT", "", 20, "A timeout (in seconds) for the render.")
	RootCmd.Flags().IntVarP(&splashResourceTimeout, "SPLASH_RESOURCE_TIMEOUT", "", 30, "A timeout (in seconds) for individual network requests.")
	RootCmd.Flags().Float64VarP(&splashWait, "SPLASH_WAIT", "", 0.5, "Time in seconds to wait until js scripts loaded.")
	viper.AutomaticEnv() // read in environment variables that match
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))
	viper.BindPFlag("proxy", RootCmd.Flags().Lookup("proxy"))
	viper.BindPFlag("REDIS", RootCmd.Flags().Lookup("REDIS"))
	viper.BindPFlag("REDIS_EXPIRE", RootCmd.Flags().Lookup("REDIS_EXPIRE"))
	viper.BindPFlag("REDIS_NETWORK", RootCmd.Flags().Lookup("REDIS_NETWORK"))
	viper.BindPFlag("SPLASH", RootCmd.Flags().Lookup("SPLASH"))
	viper.BindPFlag("SPLASH_TIMEOUT", RootCmd.Flags().Lookup("SPLASH_TIMEOUT"))
	viper.BindPFlag("SPLASH_RESOURCE_TIMEOUT", RootCmd.Flags().Lookup("SPLASH_RESOURCE_TIMEOUT"))
	viper.BindPFlag("SPLASH_WAIT", RootCmd.Flags().Lookup("SPLASH_WAIT"))

	//RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is the working directory)")
}

// initConfig reads in config file and ENV variables if set.
/*
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".dataflowkit") // name of config file (without extension)
	viper.AddConfigPath(".")            // look for config in the working directory
	viper.AddConfigPath("$HOME")        // adding home directory to search path
	viper.AddConfigPath("/etc/dataflowkit/")   // path to look for the config file in
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
		// Handle errors reading the config file
	}
}
*/
