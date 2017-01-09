package log15

import "unsafe"

// Writer implements io.Writer interface for Logger functions
// This allows using logger as io.Writer. Usage:
//
//    writer := Writer{Log: logger.Debug}
//    function_taking_IOWriter(writer)
type Writer struct {
	Log func(msg string, ctx ...interface{})
}

func (w Writer) Write(p []byte) (n int, err error) {
	str := ""
	if p != nil {
		str = *(*string)(unsafe.Pointer(&p))
	}
	w.Log(str)
	return len(p), nil
}
