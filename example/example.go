package main

import (
	"fmt"
	"os"

	"github.com/facebookgo/stack"
	"github.com/robert-zaremba/errstack"
	"github.com/robert-zaremba/log15"
)

func main() {
	log15.Trace("page accessed", "path", "/a/bbb/c")
	log15.Debug("page accessed", "path", "/a/bbb/c")
	log15.Info("page accessed", "path", "/a/bbb/c")
	log15.Warn("page accessed", "path", "/a/bbb/c")
	log15.Error("page accessed", "path", "/a/bbb/c")
	log15.Crit("page accessed", "path", "/a/bbb/c")

	f := log15.TerminalFormat{WithColor: true, TimeFmt: " 01-02 15:04:05 ", Name: "myapp"}
	h := log15.StreamHandler(os.Stderr, f)
	h = log15.SyncHandler(h)
	h = log15.CallerFileHandler(h, true)
	h = log15.CallerFuncHandler(h)
	h = log15.LvlFilterHandler(log15.LvlTrace, h)
	l := log15.Get("example-logger")
	l.SetHandler(h)
	sampleLogs(l)
}

func sampleLogs(l log15.Logger) {
	type MyStruct struct {
		FieldA int
		FieldB string
		FieldC float64
	}
	v := MyStruct{123, "river", 3.14159}

	err := fmt.Errorf("sample error")
	l.Info("Log without arguments")
	l.Info("I'm the message", "value", 3, "dog", 4,
		log15.Alone("t1", 12), log15.Alone("t2", 13))
	l.Debug("Debug with error", err)
	l.Error("Simple error", err, "key1", 123)
	failMe(l)

	errA := errstack.WrapAsDomain(err, "I can't handle this setup")
	l.Error("Here we have an infrastructure error", errA)
	l.Crit("And now Critical goes", log15.Spew(v), fmt.Errorf("this is error %d", 1), err)
}

func handlePanic(l log15.Logger) {
	errMsg := recover()
	if errMsg == nil {
		return
	}
	s := stack.Callers(1)
	l.Crit("Handler crashed", log15.Alone("error", fmt.Sprintf("%+v", errMsg)),
		log15.Alone("stacktrace", s))
}

func failMe(l log15.Logger) {
	defer handlePanic(l)
	panic("ops, panic is here!")
}
