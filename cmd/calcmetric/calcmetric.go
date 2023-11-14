package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
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

func isCalculated(db *sql.DB, table, timeRange string, debug bool, env map[string]string, dtf, dtt time.Time) (bool, error) {
	dtf = lib.DayStart(dtf)
	dtt = lib.NextDayStart(dtt)
	sql := fmt.Sprintf(
		`select last_calculated_at from "%s" where time_range = $1 and date_from = $2 and date_to = $3`,
		table,
	)
	args := []interface{}{timeRange, dtf, dtt}
	if debug {
		lib.Logf("executing sql: %s\nwith args: %+v\n", sql, args)
	}
	rows, err := db.Query(sql, args...)
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			errName := e.Code.Name()
			if errName == "undefined_table" {
				if debug {
					lib.Logf("table '%s' does not exist yet, so we need to calculate this metric.\n", table)
				}
				return false, nil
			}
			return false, err
		default:
			return false, err
		}
	}
	defer func() { _ = rows.Close() }()
	var lastCalc time.Time
	for rows.Next() {
		err := rows.Scan(&lastCalc)
		if err != nil {
			return false, err
		}
	}
	err = rows.Err()
	if err != nil {
		return false, err
	}
	lib.Logf("table '%s' was last computed at %+v for (%s, %+v, %+v), so skipping calculation\n", table, lastCalc, timeRange, dtf, dtt)
	return true, nil
}

func needsCalculation(db *sql.DB, table string, debug bool, env map[string]string) (bool, error) {
	_, ok := env["FORCE_CALC"]
	if ok {
		return true, nil
	}
	timeRange, _ := env["TIME_RANGE"]
	switeh timeRange {
	case "c":
		dtFrom, ok := env["DATE_FROM"]
		if !ok {
			return true, fmt.Errorf("you must specify %sDATE_FROM when using %sTIME_RANGE=c", gPrefix, gPrefix)
		}
		dtTo, ok := env["DATE_TO"]
		if !ok {
			return true, fmt.Errorf("you must specify %sDATE_TO when using %sTIME_RANGE=c", gPrefix, gPrefix)
		}
		dtf, err := lib.TimeParseAny(dtFrom)
		if err != nil {
			return true, err
		}
		dtt, err := lib.TimeParseAny(dtTo)
		if err != nil {
			return true, err
		}
		isCalc, err := isCalculated(db, table, timeRange, debug, env, dtf, dtt)
		if err != nil {
			return true, err
		}
		if 1 == 1 {
			return !isCalc, nil
		}
	default:
		return true, fmt.Errorf("unknown time range: '%s'", timeRange)
	}
	return true, nil
}

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
	table, _ := env["TABLE"]
	needsCalc, err := needsCalculation(db, table, debug, env)
	if err != nil {
		return err
	}
	if !needsCalc {
		if debug {
			lib.Logf("table '%s' doesn't need calculation now\n", table)
		}
		return nil
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
