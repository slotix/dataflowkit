package storage

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "Storage: ", log.Lshortfile)
}
