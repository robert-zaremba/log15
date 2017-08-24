package log15

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/robert-zaremba/go-tty"
)

const (
	timeFormat     = "2006-01-02T15:04:05-0700"
	termTimeFormat = " 01-02|15:04:05"
	floatFormat    = 'f'
)

// Format  is the interface implemented by StreamHandler formatters.
type Format interface {
	Format(r *Record) []byte
}

// FormatF is a corresponding function for Format interface
type FormatF func(*Record) []byte

// Format implements Format interface
func (f FormatF) Format(r *Record) []byte {
	return f(r)
}

// FormatFunc returns a new Format object which uses
// the given function to perform record formatting.
func FormatFunc(f func(*Record) []byte) Format {
	return formatFunc(f)
}

type formatFunc func(*Record) []byte

func (f formatFunc) Format(r *Record) []byte {
	return f(r)
}

// LogfmtFormat construct records and prints them in using Logfmt.
func LogfmtFormat() FormatF {
	return func(r *Record) []byte {
		common := []interface{}{r.KeyNames.Time, r.Time, r.KeyNames.Lvl, r.Lvl, r.KeyNames.Msg, r.Msg}
		buf := &bytes.Buffer{}
		Logfmt(buf, append(common, r.Ctx...), 0)
		return buf.Bytes()
	}
}

// Logfmt prints records in logfmt format,
// an easy machine-parseable but human-readable format for key/value pairs.
//
// For more details see: http://godoc.org/github.com/kr/logfmt
func Logfmt(buf *bytes.Buffer, ctx []interface{}, color tty.ECode) {
	var v string
	for i := 0; i < len(ctx); i += 2 {
		if i != 0 {
			buf.WriteByte(' ')
		}
		k, ok := ctx[i].(string)
		if ok {
			v = FormatLogfmtValue(ctx[i+1])
		} else {
			k, v = errorKey, FormatLogfmtValue(k)
		}

		// XXX: we should probably check that all of your key bytes aren't invalid
		if color > 0 {
			fmt.Fprint(buf, tty.AnsiEscapeS(color, k), "=", v)
		} else {
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}

	buf.WriteByte('\n')
}

// JsonFormat formats log records as JSON objects separated by newlines.
// It is the equivalent of JsonFormatEx(false, true).
func JsonFormat() Format {
	return JsonFormatEx(false, true)
}

// JsonFormatEx formats log records as JSON objects. If pretty is true,
// records will be pretty-printed. If lineSeparated is true, records
// will be logged with a new line between each record.
func JsonFormatEx(pretty, lineSeparated bool) FormatF {
	jsonMarshal := json.Marshal
	if pretty {
		jsonMarshal = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		}
	}

	return func(r *Record) []byte {
		props := make(map[string]interface{})

		props[r.KeyNames.Time] = r.Time
		props[r.KeyNames.Lvl] = r.Lvl.String()
		props[r.KeyNames.Msg] = r.Msg

		for i := 0; i < len(r.Ctx); i += 2 {
			k, ok := r.Ctx[i].(string)
			if !ok {
				props[errorKey] = fmt.Sprintf("%+v is not a string key", r.Ctx[i])
			}
			props[k] = formatJSONValue(r.Ctx[i+1])
		}

		b, err := jsonMarshal(props)
		if err != nil {
			b, _ = jsonMarshal(map[string]string{
				errorKey: err.Error(),
			})
			return b
		}

		if lineSeparated {
			b = append(b, '\n')
		}

		return b
	}
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}

func formatJSONValue(value interface{}) interface{} {
	value = formatShared(value)
	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}

// FormatLogfmtValue converts value to string
func FormatLogfmtValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprint(v)
	}
}

var stringBufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func escapeString(s string) string {
	needsQuotes := false
	needsEscape := false
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' {
			needsQuotes = true
		}
		if r == '\\' || r == '"' || r == '\n' || r == '\r' || r == '\t' {
			needsEscape = true
		}
	}
	if !needsEscape && !needsQuotes {
		return s
	}
	e := stringBufPool.Get().(*bytes.Buffer)
	e.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			e.WriteByte('\\')
			e.WriteByte(byte(r))
		case '\n':
			e.WriteString("\\n")
		case '\r':
			e.WriteString("\\r")
		case '\t':
			e.WriteString("\\t")
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
	var ret string
	if needsQuotes {
		ret = e.String()
	} else {
		ret = string(e.Bytes()[1 : e.Len()-1])
	}
	e.Reset()
	stringBufPool.Put(e)
	return ret
}
