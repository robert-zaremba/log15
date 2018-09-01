// Here we implement terminal fancy Formater for logger.

package log15

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/facebookgo/stack"
	"github.com/robert-zaremba/go-tty"
)

const termMsgJust = 46

// TerminalFormat Formatter
type TerminalFormat struct {
	WithColor bool
	TimeFmt   string
	Name      string
}

func (tf TerminalFormat) timeStr(r *Record) string {
	if tf.TimeFmt == "" {
		return " "
	}
	return r.Time.Format(tf.TimeFmt)
}

// Format implements Format interface for terminal output.
// It formats log records optimized for human readability on
// a terminal with color-coded level output and terser human friendly timestamp.
// This format should only be used for interactive programs or while developing.
//
//     [TIME] [LEVEL] MESAGE key=value key=value ...
//
// Example:
//
//     [May 16 20:58:45] [DBUG] remove route ns=haproxy addr=127.0.0.1:50002
//
// * structured log attributes are not colored, but bolded when `WithColor = true`
// * support for implicit keys for error attributes - if the error comes without string
//   title, then it's printed separately in the new line.
// * support for verbose attribute printing: `Spew`
func (tf TerminalFormat) Format(r *Record) []byte {
	var color tty.ECode
	if tf.WithColor {
		switch r.Lvl {
		case LvlError, LvlCrit:
			color = tty.RED
		case LvlWarn:
			color = tty.YELLOW
		case LvlInfo:
			color = tty.MAGENTA
		}
	}
	b := &bytes.Buffer{}
	lvl := r.Lvl.StringUP()
	t := tf.timeStr(r)
	caller := findCaller(r.Ctx)
	if color > 0 {
		fmt.Fprint(b, tty.AnsiEscapeS(color, lvl), " ", tf.Name, t, caller, "] ", r.Msg, "  ")
	} else {
		fmt.Fprint(b, lvl, " ", tf.Name, t, caller, "] ", r.Msg)
	}

	// try to justify the log output for short messages
	if len(r.Ctx) > 0 && len(r.Msg) < termMsgJust {
		_, _ = b.Write(bytes.Repeat([]byte{' '}, termMsgJust-len(r.Msg)))
	}
	if tf.WithColor {
		color = tty.BOLD
	} else {
		color = 0
	}
	logfmt(b, r.Ctx, color)
	return b.Bytes()
}

// logfmt prints records in logfmt format,
// an easy machine-parseable but human-readable format for key/value pairs.
// It support implicit error keys. By passing error you don't need to write a key for it.
// For more details see: http://godoc.org/github.com/kr/logfmt
func logfmt(buf *bytes.Buffer, ctx []interface{}, color tty.ECode) {
	var k, v string
	var errs []error
	var spews []SpewWrapper
	var alones []aloneWrapper
	for i := 0; i < len(ctx); i++ {
		if i != 0 {
			_ = buf.WriteByte(' ')
		}
		switch vt := ctx[i].(type) {
		case error:
			errs = append(errs, vt)
			continue
		case SpewWrapper:
			spews = append(spews, vt)
			continue
		case aloneWrapper:
			alones = append(alones, vt)
			continue
		case string:
			k = vt
		case nil, CallerCtx:
			continue
		default:
			k = "MALFORMED_LOGFMT_KEY"
		}
		i++
		if i >= len(ctx) {
			v = "MALFORMED_LOGFMT: no value for last key"
		} else {
			v = FormatLogfmtValue(ctx[i])
		}
		if color > 0 {
			fmt.Fprint(buf, tty.AnsiEscapeS(color, k), "=", v)
		} else {
			fmt.Fprint(buf, k, "=", v)
		}
	}
	for _, s := range alones {
		_, _ = buf.WriteString("\n  * ")
		_, _ = buf.WriteString(s.title)
		_, _ = buf.WriteString(": ")
		_, _ = buf.WriteString(FormatLogfmtValue(s.obj))
	}
	for _, s := range spews {
		if s.Msg == "" {
			s.Msg = "spew"
		}
		_, _ = buf.WriteString("\n-------- ")
		_, _ = buf.WriteString(s.Msg)
		_, _ = buf.WriteString(" --------\n")
		_, _ = buf.WriteString(spew.Sdump(s.Obj))
	}
	if len(spews) == 0 {
		_ = buf.WriteByte('\n')
	}
	var errStr string
	for _, err := range errs {
		errStr = err.Error()
		if e, ok := err.(FancyError); ok && !e.IsReq() {
			_, _ = buf.WriteString(errInfHeader)
			_, _ = buf.WriteString(errStr)
			_, _ = buf.WriteString("\nstacktrace:\n")
			_, _ = buf.WriteString(e.Stacktrace().String())
		} else {
			_, _ = buf.WriteString(errHeader)
			_, _ = buf.WriteString(errStr)
		}
		_ = buf.WriteByte('\n')
	}
}

// FancyError is an enhanced Error type - we can check if it's a (user) request like error
// and call for a stacktrace.
type FancyError interface {
	IsReq() bool
	Stacktrace() *stack.Multi
}

const errHeader = "-------- ERROR --------\n"
const errInfHeader = "-------- ERROR (infrastructure) --------\n"

// SpewWrapper wraps the object to preaty print it using spew package.
type SpewWrapper struct {
	Obj interface{}
	Msg string
}

// Spew is a helper method to wrap an object into SpewWrapper.
// You can optionally add an obj description.
func Spew(obj interface{}, description ...string) interface{} {
	s := SpewWrapper{obj, ""}
	if len(description) > 0 {
		s.Msg = description[0]
	}
	return s
}

type aloneWrapper struct {
	title string
	obj   interface{}
}

// Alone wraps the object to print it in a separate line.
func Alone(title string, obj interface{}) interface{} {
	return aloneWrapper{title, obj}
}
