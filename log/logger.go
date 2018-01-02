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
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			break
		}
	}
	return nil
}

//NewLogger creates New Logger instance.
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.AddHook(ContextHook{})
	return logger
}
