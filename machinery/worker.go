package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"fmt"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/errors"
	"github.com/RichardKnop/machinery/v1/signatures"
	"github.com/slotix/dataflowkit/parser"
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
	if response.Reason != "" {
		return fmt.Sprintf("%s : %s", url, response.Reason), err
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

func generateGetHTMLTask(url string) *signatures.TaskSignature {
	return &signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
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
		var tasks []*signatures.TaskSignature
		fmt.Println("i=", i)
		for j, url := range urls[i : i+workerStep] {
			tasks = append(tasks, generateGetHTMLTask(url))
			fmt.Printf("%d - %s \n", i+j, url)
		}
		group := machinery.NewGroup(tasks...)
		asyncResults, err := server.SendGroup(group)
		errors.Fail(err, "Could not send task")
		for _, asyncResult := range asyncResults {
			_, err := asyncResult.Get()
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

func MarshalData(payload []byte) (io.ReadCloser, error) {
	parser, err := parser.NewParser(payload)
	if err != nil {
		return nil, err
	}
	res, err := parser.MarshalData()
	if err != nil {
		return nil, err
	}
	return res, nil
}
