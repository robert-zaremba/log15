package rollbar

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/robert-zaremba/log15"
	"github.com/stvp/rollbar"
)

// Config wraps Rollbar parameters
type Config struct {
	Token   string `yaml:"token"`
	Env     string `yaml:"-"`
	Version string `yaml:"-"`
}

// Check validates the config content
func (rc Config) Check() error {
	if rc.Token == "" {
		return errors.New("Wrong `token` value. Can't be empty")
	}
	if rc.Env == "" {
		return errors.New("Wrong `env` value. Can't be empty")
	}
	if rc.Version == "" {
		return errors.New("Wrong `version` value. Can't be empty")
	}
	return nil
}

func toRollbarLevel(l log15.Lvl) string {
	switch l {
	case log15.LvlCrit:
		return rollbar.CRIT
	case log15.LvlError:
		return rollbar.ERR
	case log15.LvlWarn:
		return rollbar.WARN
	case log15.LvlInfo:
		return rollbar.INFO
	default:
		return rollbar.DEBUG
	}
}

// LogInternalErrors logs rollbar post errors
func LogInternalErrors(l log15.Logger) {
	for err := range rollbar.PostErrors() {
		l.Error("Rollbar post error", err)
	}
}

// MkHandler creates a handler for Rollbar service
func MkHandler(c Config) log15.FuncHandlerT {
	rollbar.Token = c.Token
	rollbar.Environment = c.Env
	rollbar.CodeVersion = c.Version
	return rollbarHandler
}

func rollbarHandler(r *log15.Record) error {
	if r.Lvl > log15.LvlWarn {
		return nil
	}

	var k, v, caller string
	var err error
	var fields = []*rollbar.Field{{Name: "message", Data: r.Msg}}
	var erri, spewi int
	for i := 0; i < len(r.Ctx); i++ {
		switch vt := r.Ctx[i].(type) {
		case error:
			if err == nil {
				err = vt
				continue
			}
			k = "error" + strconv.Itoa(erri)
			v = fmt.Sprint(err)
			erri++
		case log15.SpewWrapper:
			if vt.Msg == "" {
				k = "spew" + strconv.Itoa(spewi)
				spewi++
			} else {
				k = vt.Msg
			}
			v = spew.Sdump(vt.Obj)
		case string:
			k = vt
			i++
			if i >= len(r.Ctx) {
				v = "MALFORMED_LOGFMT: no value for last key"
			} else {
				v = log15.FormatLogfmtValue(r.Ctx[i])
			}
		case log15.CallerCtx:
			caller = string(vt)
			continue
		default:
			k = "MALFORMED_LOGFMT_KEY"
			v = log15.FormatLogfmtValue(vt)
		}
		fields = append(fields, &rollbar.Field{Name: k, Data: v})
	}
	fields, stack, fmtError := mkStacktrace(caller, err, fields)
	lvl := toRollbarLevel(r.Lvl)
	if err == nil {
		err = errors.New(r.Msg)
	}

	if stack != nil {
		rollbar.ErrorWithStack(lvl, err, stack, fields...)
	} else {
		rollbar.ErrorWithStackSkip(lvl, err, 11, fields...)
	}
	return fmtError
}

func mkStacktrace(caller string, err error, fields []*rollbar.Field) ([]*rollbar.Field, rollbar.Stack, error) {
	var stack rollbar.Stack
	var fmtError error
	if errStack, ok := err.(log15.FancyError); ok {
		stacks := errStack.Stacktrace().Stacks()
		if len(stacks) > 0 {
			for _, f := range stacks[0] {
				stack = append(stack,
					rollbar.Frame{Filename: f.File, Method: f.Name, Line: f.Line})
			}
		}
	}
	if caller != "" {
		if stack != nil {
			i := strings.LastIndex(caller, ":")
			frame := rollbar.Frame{Filename: caller[:i]}
			frame.Line, fmtError = strconv.Atoi(caller[i+1:])
			stack = rollbar.Stack{frame}
		} else {
			fields = append(fields, &rollbar.Field{Name: "caller", Data: caller})
		}
	}
	return fields, stack, fmtError
}

// ReporterLogger is a reduced logger interface for critical messages
type ReporterLogger interface {
	Crit(msg string, ctx ...interface{})
	Debug(msg string, ctx ...interface{})
}

// WaitForRollbar is a panic handler that waits for rollbar if rollbar is configured.
func WaitForRollbar(logger ReporterLogger) {
	if rollbar.Token == "" {
		return
	}
	rollbar.Wait()
	if err := recover(); err != nil {
		if rollbar.Token != "" {
			if _, ok := err.(log15.FatalMessage); !ok { // logger.Fatal are already handled
				logger.Crit("PANIC. ", err)
			}
			rollbar.Wait()
		}
		panic(err)
	}
}
