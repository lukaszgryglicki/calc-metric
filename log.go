package calcmetric

import (
	"fmt"
	"time"
)

// Printf is a wrapper around Printf(...) that supports logging.
func Printf(format string, args ...interface{}) (int, error) {
	return fmt.Printf("%s: "+format, append([]interface{}{time.Now()}, args...)...)
}
