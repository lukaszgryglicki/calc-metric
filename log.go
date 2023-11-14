package calcmetric

import (
	"fmt"
	"time"
)

// Logf is a wrapper around Printf(...) that supports logging.
func Logf(format string, args ...interface{}) (int, error) {
	return fmt.Printf("%s: "+format, append([]interface{}{ToYMDHMS(time.Now())}, args...)...)
}
