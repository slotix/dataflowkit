package parse

import (
	"github.com/slotix/dataflowkit/log"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = log.NewLogger()
}

/*
var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "Parser: ", log.Lshortfile)
}
*/
