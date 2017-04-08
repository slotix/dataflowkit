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
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"testing"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/errors"
	"github.com/RichardKnop/machinery/v1/logger"
	"github.com/RichardKnop/machinery/v1/signatures"
)

// Define flagss
var (
	//	configPath    = flag.String("c", "config.yml", "Path to a configuration file")
	//	broker        = flag.String("b", "redis://0.0.0.0:6379", "Broker URL")
	//	resultBackend = flag.String("r", "redis://0.0.0.0:6379", "Result backend")

	//	cnf                                             config.Config
	//	server                                          *machinery.Server
	task0, task1, task2, task3, task4, task5, task6 signatures.TaskSignature

	urlsFile = "forum_list.csv"
	from     = 0
	to       = 0
)

func init() {
	flag.StringVar(&urlsFile, "f", urlsFile, "url list file")
	flag.IntVar(&from, "from", from, "starting from record number")
	flag.IntVar(&to, "to", to, "ending to record number")

	// Parse the flags
	flag.Parse()

	cnf = config.Config{
		Broker:        *broker,
		ResultBackend: *resultBackend,
	}

	// Parse the config
	// NOTE: If a config file is present, it has priority over flags

	data, err := config.ReadFromFile(*configPath)
	if err == nil {
		err = config.ParseYAMLConfig(&data, &cnf)
		errors.Fail(err, "Could not parse config file")
	}

	server, err = machinery.NewServer(&cnf)
	logger.Get().Println(cnf.ResultBackend)
	errors.Fail(err, "Could not initialize server")
}

func initTasks() {
	task0 = signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
			{
				Type:  "string",
				Value: "http://skincrafter.com",
			},
		},
	}

	task1 = signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
			{
				Type:  "string",
				Value: "http://google.com",
			},
		},
	}

	task2 = signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
			{
				Type:  "string",
				Value: "http://dbconvert.com",
			},
		},
	}

	task3 = signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
			{
				Type:  "string",
				Value: "http://yahoo.com",
			},
		},
	}

	task4 = signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
			{
				Type:  "string",
				Value: "http://diesel.elcat.kg",
			},
		},
	}

	task5 = signatures.TaskSignature{
		Name: "GetHTML",
		Args: []signatures.TaskArg{
			{
				Type:  "string",
				Value: "http://dataflowkittt.org",
			},
		},
	}
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
	workerStep := 5
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

func TestOut(t *testing.T) {
	//initTasks()
	fmt.Println("Send tasks:")
	urls := LoadURLsFromCSV("forum_list.csv")
	//SendTasksToRedis(urls, from, to)
	SendTasksToRedis(urls, 150, 175)
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
