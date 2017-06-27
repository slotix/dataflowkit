package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"fmt"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

func init() {
	viper.Set("splash", "127.0.0.1:8050")
	viper.Set("splash-timeout", "20")
	viper.Set("splash-resource-timeout", "30")
	viper.Set("splash-wait", "0,5")
}

func RunTasks(from, to int) {
	//initTasks()
	fmt.Println("Send tasks:")
	urls := LoadURLsFromCSV("forum_list.csv")
	//SendTasksToRedis(urls, from, to)
	SendTasksToRedis(urls, from, to)
}

func GetResponse(url string) (string, error) {
	req := splash.Request{URL: url}
	//logger.Println(req.URL)
	splashURL, err := splash.NewSplashConn(req)
	response, err := splash.GetResponse(splashURL)
	if err != nil {
		return "", err
	}
	if response.Error != "" {
		return fmt.Sprintf("%s : %s", url, response.Error), err
	}
	return fmt.Sprintf("%s : %s", url, "Ok"), err

	/*
		_, err := splash.Fetch(url)
		if err != nil {
			return "", err
		}
		//	time.Sleep(10 * time.Second)
		return "200", nil
	*/
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

func generateGetHTMLTask(url string) *tasks.Signature {
	return &tasks.Signature{
		Name: "GetHTML",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: url,
			},
		},
	}
}

//SendItemsToQuery sends urls to fetchbot query to be parsed
func SendTasksToRedis(urls []string, from, to int) {
	workerStep := 20
	var succeeded, failed int
	start := time.Now()
	i := from
	for i < to {
		if dif := to - i; dif < workerStep {
			workerStep = dif
		}
		var tsks []*tasks.Signature
		fmt.Println("i=", i)
		for j, url := range urls[i : i+workerStep] {
			tsks = append(tsks, generateGetHTMLTask(url))
			fmt.Printf("%d - %s \n", i+j, url)
		}
		group := tasks.NewGroup(tsks...)
		asyncResults, err := server.SendGroup(group)
		if err != nil {
			fmt.Println(err, "Could not send task")
		}

		for _, asyncResult := range asyncResults {
			_, err := asyncResult.Get(time.Duration(time.Millisecond * 5))
			taskState := asyncResult.GetState()
			fmt.Printf("URL: %v Current state of %v task is: %s\n", asyncResult.Signature.Args[0].Value, taskState.TaskUUID, taskState.State)
			if taskState.State == "SUCCESS" {
				succeeded++
			} else {
				failed++
			}
			if err != nil {
				fmt.Println(err)
			}
		}
		i += workerStep
	}
	fmt.Printf("Summary: %.2fm elapsed\n", time.Since(start).Minutes())
	fmt.Printf("succeeded = %d; failed = %d\n", succeeded, failed)

}

/*
func GetHTML1(url string) (string, error) {
	content, err := splash.Fetch(url)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
*/
/*
func MarshalData(payload []byte) (io.ReadCloser, error) {
	parser, err := scrape.NewPayload(payload)
	if err != nil {
		return nil, err
	}
	//res, err := parser.MarshalData()
	//if err != nil {
	//	return nil, err
	//}
	return res, nil
}
*/
