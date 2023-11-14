package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	lib "github.com/lukaszgryglicki/calcmetric"
)

const (
	gPrefix = "V3_"
)

var (
	gRequired = []string{
    "CONN",
    "METRIC",
    "TABLE",
    "PROJECT_SLUG",
    "TIME_RANGE",
  }
)

func calcMetric() error {
	env := make(map[string]string)
	prefixLen := len(gPrefix)
	for _, pair := range os.Environ() {
		if strings.HasPrefix(pair, gPrefix) {
			ary := strings.Split(pair, "=")
			if len(ary) != 2 {
				continue
			}
			env[ary[0][prefixLen:]] = ary[1]
		}
	}
	_, debug := env["DEBUG"]
	if debug {
		lib.Logf("map: %+v\n", env)
	}
	for _, key := range gRequired {
		_, ok := env[key]
		if !ok {
			msg := fmt.Sprintf("you must define %s%s environment variable to run this", gPrefix, key)
			lib.Logf("%s\n", msg)
			err := fmt.Errorf("%s", msg)
			return err
		}
	}
	return nil
}

func main() {
	dtStart := time.Now()
	err := calcMetric()
	if err != nil {
		lib.Logf("calcMetric error: %+v\n", err)
	}
	dtEnd := time.Now()
	lib.Logf("Time: %v\n", dtEnd.Sub(dtStart))
}
