package calcmetric

import (
	"fmt"
	"reflect"
	"time"
)

// QueryOut - output query and its arguments
func QueryOut(query string, args ...interface{}) {
	Logf("%s\n", query)
	if len(args) > 0 {
		s := ""
		for vi, vv := range args {
			switch v := vv.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, string, bool, time.Time:
				s += fmt.Sprintf("%d:%+v ", vi+1, v)
			case nil:
				s += fmt.Sprintf("%d:(null) ", vi+1)
			default:
				s += fmt.Sprintf("%d:%+v ", vi+1, reflect.ValueOf(vv))
			}
		}
		Logf("[%s]\n", s)
	}
}

// Logf is a wrapper around Printf(...) that supports logging.
func Logf(format string, args ...interface{}) (int, error) {
	return fmt.Printf("%s: "+format, append([]interface{}{ToYMDHMS(time.Now())}, args...)...)
}
