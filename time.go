package calcmetric

import (
	"fmt"
	"time"
)

// ToYMDHMS - return time formatted as YYYY-MM-DD HH:MI:SS
func ToYMDHMS(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
}
