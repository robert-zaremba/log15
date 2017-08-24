package log15

import (
	"fmt"
	"os"

	"github.com/go-stack/stack"
)

// CallerCtx is log15 ctx element type to signify the CallerFileHandler output
type CallerCtx string

func findCaller(ctx []interface{}) CallerCtx {
	for i := len(ctx) - 1; i >= 0; i-- { // most probably caller is at the end
		if s, ok := ctx[i].(CallerCtx); ok {
			return s
		}
	}
	return ""
}

// CallerFileHandler returns a Handler that adds the line number and filename of
// the calling function to the context with key "caller".
// If short then caller filename is shortened to the parent directory where the calling
// function is placed. Otherwise absolute path is used.
func CallerFileHandler(h Handler, short bool) FuncHandlerT {
	return func(r *Record) error {
		caller := fmt.Sprintf("%#v", r.Call)
		if short {
			caller = shortFilename(caller)
		}
		r.Ctx = append(r.Ctx, CallerCtx(caller))
		return h.Log(r)
	}
}

// returns 'basename(dirname(filename)'/basename(filename)
func shortFilename(filename string) string {
	var prev, last = -1, -1
	for i := 0; i < len(filename); i++ {
		if os.IsPathSeparator(filename[i]) {
			prev, last = last, i
		}
	}
	if last < 0 {
		return filename
	}
	return filename[prev+1:]
}

// CallerFuncHandler returns a Handler that adds the calling function name to
// the context with key "fn".
func CallerFuncHandler(h Handler) Handler {
	return FuncHandler(func(r *Record) error {
		r.Ctx = append(r.Ctx, "fn", fmt.Sprintf("%+n", r.Call))
		return h.Log(r)
	})
}

// CallerStackHandler returns a Handler that adds a stack trace to the context
// with key "stack". The stack trace is formated as a space separated list of
// call sites inside matching []'s. The most recent call site is listed first.
// Each call site is formatted according to format. See the documentation of
// package github.com/go-stack/stack for the list of supported formats.
func CallerStackHandler(format string, h Handler) Handler {
	return FuncHandler(func(r *Record) error {
		s := stack.Trace().TrimBelow(r.Call).TrimRuntime()
		if len(s) > 0 {
			r.Ctx = append(r.Ctx, "stack", fmt.Sprintf(format, s))
		}
		return h.Log(r)
	})
}
