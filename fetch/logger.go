package fetch

import (
	 "log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "Fetcher: ", log.Lshortfile)
}
