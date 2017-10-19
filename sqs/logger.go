package sqs

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "SQS: ", log.Lshortfile)
}
