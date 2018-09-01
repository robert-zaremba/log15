package log15setup

import (
	"errors"
	"os"
	"strings"

	"github.com/robert-zaremba/log15"
	"github.com/robert-zaremba/log15/rollbar"
)

var root = log15.Root()

// timeFMT is the map of predefined time formats for TerminalFmt
var timeFMT = map[string]string{
	"off":        "",
	"date":       " 01-02 ",
	"date-year":  " 2006-01-02 ",
	"sec":        " 01-02 15:04:05 ",
	"sec-year":   " 2006-01-02 15:04:05 ",
	"musec":      " 01-02 15:04:05.0000 ",
	"musec-year": " 2006-01-02 15:04:05.0000 ",
}

// Config represents  log config
type Config struct {
	Color   bool   `yaml:"color"`
	TimeFmt string `yaml:"timeFmt"` // one of timeFMT values
	Level   string `yaml:"level"`

	lvl log15.Lvl
}

// Check validates the config content
func (c *Config) Check() error {
	if _, ok := timeFMT[c.TimeFmt]; !ok {
		return errors.New("Wrong timeFmt value, should be one of log15setup.timeFMT values")
	}
	var err error
	c.lvl, err = log15.LvlFromString(c.Level)
	return err
}

// New constructs logger and registers it in the log15 repository.
// If the logger `name` is already registered then it's upgraded according to given config.
func New(name string, c Config, rc rollbar.Config) (log15.Logger, error) {
	err := c.Check()
	if err != nil {
		return nil, err
	}

	f := log15.TerminalFormat{WithColor: c.Color, TimeFmt: timeFMT[c.TimeFmt], Name: name}
	h := log15.StreamHandler(os.Stderr, f)
	h = log15.SyncHandler(h)
	stderrHandler := h
	if rc.Token != "" {
		if err = rc.Check(); err != nil {
			return nil, err
		}
		h = log15.MultiHandler(h, rollbar.MkHandler(rc))

		rollbarLogger := log15.Get("rollbar")
		rollbarLogger.SetHandler(stderrHandler)
		go rollbar.LogInternalErrors(rollbarLogger)
	}
	h = log15.CallerFileHandler(h, true)
	h = log15.LvlFilterHandler(c.lvl, h)
	// l := log15.Get(name)
	root.SetHandler(h)
	if rc.Token == "" {
		root.Info("Rollbar token not set. Disabling rollbar integration.")
	}
	return root, nil
}

// MustLogger setups logger. It panics when the provided configuration is malformed.
// envName is the name of running environment; eg: localhost, qa, stagging, prod...
func MustLogger(envName, appname, version, rollbartoken, timeFmt, level string, colored bool) {
	MustAppName(envName, "environment name")
	MustAppName(appname, "application name")
	appname = envName + ":" + appname
	rc := rollbar.Config{
		Version: version,
		Env:     envName,
		Token:   rollbartoken}
	// we don't need to overwrite the global object
	_, err := New(appname,
		Config{Color: colored, TimeFmt: timeFmt, Level: level}, rc)
	if err != nil {
		root.Fatal("Can't initialize logger", err)
	}
	if strings.HasPrefix(envName, "prod") && rollbartoken == "" {
		root.Error("Rollbar token must be set in production environment")
	}
	root.Debug("Logger initialized", "app_version", version)
}
