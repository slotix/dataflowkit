package server

import (
	"bufio"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)


func TestGetResponse(t *testing.T) {
	viper.Set("splash", "127.0.0.1:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")
	logger.Println("Load URLs:")
	urls := LoadURLsFromCSV("forum_list.csv")
	for _, url := range urls[9:10] {
		req := splash.Request{URL: url}
		logger.Println(req.URL)
		splashURL, err := splash.NewSplashConn(req)
		response, err := splash.GetResponse(splashURL)
		if err != nil {
			logger.Println(err)
		} else {
			if response.Error != "" {
				logger.Println(response.Error)
			} else {
				logger.Println(response.Response.Ok)
			}
		}
	}
}

//LoadURLsFromCSV loads list of URLs from CSV
func LoadURLsFromCSV(file string) []string {
	// Load a CSV file.
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		urls = append(urls, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return urls

}
