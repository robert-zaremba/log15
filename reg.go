package log15

import (
	"errors"
	"sync"
)

// reg is the registry of loggers.
// It's used to share loggers among libraries
var reg = map[string]Logger{}
var regMutex = sync.Mutex{}

// Get returns a logger from the registry. If it doesnt' exsist it creates and registers
// the new one.
func Get(name string) Logger {
	regMutex.Lock()
	defer regMutex.Unlock()
	l, ok := reg[name]
	if ok {
		return l
	}
	l = New()
	reg[name] = l
	return l
}

// Set puts a new logger into registry if it's not yet there,
// otherwise copy into the old logger. It requires that the Logger type is *logger
func Set(name string, l Logger) error {
	regMutex.Lock()
	defer regMutex.Unlock()
	lOld, ok := reg[name]
	if !ok {
		reg[name] = l
		return nil
	}
	ll, ok := l.(*logger)
	if !ok {
		return errors.New("unsupported logger type to overwrite already esisting logger")
	}
	llOld, ok := lOld.(*logger)
	if !ok {
		return errors.New("unsupported logger type to overwrite already esisting logger")
	}
	*llOld = *ll
	return nil
}
