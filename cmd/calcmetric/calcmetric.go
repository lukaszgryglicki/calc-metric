package main

import (
	"time"

	lib "github.com/lukaszgryglicki/calcmetric"
)

func main() {
	dtStart := time.Now()
	dtEnd := time.Now()
	lib.Logf("Time: %v\n", dtEnd.Sub(dtStart))
}
