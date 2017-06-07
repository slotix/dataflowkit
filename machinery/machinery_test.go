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
	"flag"
	"fmt"

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

func TestOut(t *testing.T) {
	//initTasks()
	fmt.Println("Send tasks:")
	//urls := LoadURLsFromCSV("forum_list.csv")
	//SendTasksToRedis(urls, from, to)
	//SendTasksToRedis(urls, 0, 11)
	RunTasks(1000, 2000)
}
