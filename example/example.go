package main

import (
	"gopkg.in/inconshreveable/log15.v2"
)

func main() {
	log15.Trace("page accessed", "path", "/a/bbb/c")
	log15.Debug("page accessed", "path", "/a/bbb/c")
	log15.Info("page accessed", "path", "/a/bbb/c")
	log15.Warn("page accessed", "path", "/a/bbb/c")
	log15.Error("page accessed", "path", "/a/bbb/c")
	log15.Crit("page accessed", "path", "/a/bbb/c")
}
