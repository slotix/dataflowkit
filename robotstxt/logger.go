package robotstxt

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "RobotsTxt: ", log.Lshortfile)
}
