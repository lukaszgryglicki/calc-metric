package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq" // As suggested by lib/pq driver
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
			if len(ary) < 2 {
				continue
			}
			key := ary[0]
			val := strings.Join(ary[1:], "=")
			env[key[prefixLen:]] = val
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
	connStr, _ := env["CONN"]
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if debug {
		lib.Logf("db: %+v\n", db)
	}
	metric, _ := env["METRIC"]
	path, ok := env["PATH"]
	if !ok {
		path = "./sql/"
	}
	fn := path + metric + ".sql"
	contents, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	sql := string(contents)
	projectSlug, _ := env["PROJECT_SLUG"]
	sql = strings.Replace(sql, "{{project_slug}}", projectSlug, -1)
	limit, _ := env["LIMIT"]
	sql = strings.Replace(sql, "{{limit}}", limit, -1)
	offset, _ := env["OFFSET"]
	sql = strings.Replace(sql, "{{offset}}", offset, -1)
	for k, v := range env {
		if strings.HasPrefix(k, "PARAM_") {
			n := k[6:]
			sql = strings.Replace(sql, "{{"+n+"}}", v, -1)
		}
	}
	if debug {
		lib.Logf("generated SQL:\n%s\n", sql)
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
