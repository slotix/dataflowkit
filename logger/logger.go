// Dataflow kit - log
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

// Package log of the Dataflow kit implements modified sirupsen/logrus logger enabling to show Log filename and line number.
//
// see more info at https://github.com/sirupsen/logrus/issues/63
package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type ContextHook struct{}

//Levels return Levels of ContextHook
func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

//Fire the hook.
func (hook ContextHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = path.Base(file)
			//entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			break
		}
	}
	return nil
}

//NewLogger creates New Logger instance.
func NewLogger(withContext bool) *logrus.Logger {
	logger := logrus.New()
	logrus.SetOutput(os.Stdout)
	if withContext {
		logger.AddHook(ContextHook{})
	}
	return logger
}

func NewFileLogger(withContext bool, fileName string) *logrus.Logger {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		fmt.Printf("Failed to create %s file", fileName)
	}
	logger := logrus.New()
	logger.Out = file

	if withContext {
		logger.AddHook(ContextHook{})
	}
	return logger
}
