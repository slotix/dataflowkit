package fetch

import (
	"github.com/slotix/dataflowkit/log"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = log.NewLogger()
}

// import (
// 	 "log"
// 	"os"
// )

// var logger *log.Logger

// func init() {
// 	logger = log.New(os.Stdout, "Fetcher: ", log.Lshortfile)
// }
