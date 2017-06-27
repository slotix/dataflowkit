package main

import (
	"flag"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"fmt"
)

// Define flags
var (
	//configPath    = flag.String("c", "config.yml", "Path to a configuration file")
	broker        = flag.String("b", "redis://127.0.0.1:6379", "Broker URL")
	resultBackend = flag.String("r", "redis://127.0.0.1:6379", "Result backend")

	cnf    config.Config
	server *machinery.Server
	worker *machinery.Worker
)

func init() {
	// Parse the flags
	flag.Parse()

	cnf = config.Config{
		Broker:             *broker,
		ResultBackend:      *resultBackend,
		MaxWorkerInstances: 5,
	}

	// Parse the config
	// NOTE: If a config file is present, it has priority over flags
	//data, err := config.ReadFromFile(*configPath)
	//if err == nil {
	//	err = config.ParseYAMLConfig(&data, &cnf)
	//	errors.Fail(err, "Could not parse config file")
	//}

	server, err := machinery.NewServer(&cnf)
	if err != nil {
		fmt.Println(err, "Could not initialize server")
	}
	
	// Register tasks
	tasks := map[string]interface{}{
		"GetHTML":      GetResponse,
	//	"Parse": Parse,
	}
	server.RegisterTasks(tasks)

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker = server.NewWorker("worker")
}

func main() {
	err := worker.Launch()
	if err != nil {
		fmt.Println(err, "Could not launch worker")
	}
}
