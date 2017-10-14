package parse

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "Parser: ", log.Lshortfile)
}
