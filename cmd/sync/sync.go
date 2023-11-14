package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	lib "github.com/lukaszgryglicki/calcmetric"
	yaml "gopkg.in/yaml.v2"
)

const (
	gPrefix = "V3_"
)

var (
	gRequired = []string{
		"CONN",
	}
)

// Metrics contain all metrics to calculate
type Metrics struct {
	Metrics map[string]Metric `yaml:"metrics"`
}

// Metric contains details about how given metric shoudl be calculated
// More details in README.md
type Metric struct {
	Metric string `yaml:"metric"` // Maps to V3_METRIC
	Table  string `yaml:"table"`  // Maps to V3_TABLE
	// Can be overwritten with V3_PROJECT_SLUGS env variable
	// Can also use "all" which connects to DB and gets all slugs using built-in SQL command
	ProjectSlugs string `yaml:"project_slugs"` // Comma separated list of V3_PROJECT_SLUG values, can also be SQL like `"sql:select distinct project_slug from mv_subprojects"`
	// Can be overwritten with V3_TIME_RANGES env variable
	TimeRanges  string            `yaml:"time_ranges"`  // Comma separated list of time ranges (V3_TIME_RANGE) to calculate or "all" which means all supported time ranges
	ExtraParams map[string]string `yaml:"extra_params"` // map k:v with `V3_PARAM_` prefix skipped in keys, for example: tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'", is_bot='!= true'
	ExtraEnv    map[string]string `yaml:"extra_env"`    // map k:v with `V3_` prefix skipped in keys, for example: DEBG=1 DATE_FROM=2023-10-01 DATE_TO=2023-11-01
}

func getQuerySlugs(db *sql.DB, debug bool, query string) ([]string, error) {
	slug, slugs := "", []string{}
	if debug {
		lib.Logf("executing the following query to get slugs:\n%s\n", query)
	}
	rows, err := db.Query(query)
	if err != nil {
		return slugs, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		err := rows.Scan(&slug)
		if err != nil {
			return slugs, err
		}
		slugs = append(slugs, slug)
	}
	err = rows.Err()
	if err != nil {
		return slugs, err
	}
	return slugs, nil
}

func runTasks(db *sql.DB, metrics Metrics, debug bool, env map[string]string) error {
	path, ok := env["CALC_PATH"]
	if !ok {
		path = "./"
	}
	calcBin := path + "calcmetric"
	if debug {
		lib.Logf("will use '%s' binary to calculate metrics\n", calcBin)
	}
	allTasks := []map[string]string{}
	for _, taskDef := range metrics.Metrics {
		task := make(map[string]string)
		// Basics
		task[gPrefix+"METRIC"] = taskDef.Metric
		task[gPrefix+"TABLE"] = taskDef.Table

		// Slugs
		slugs := taskDef.ProjectSlugs
		envSlugs, ok := env["PROJECT_SLUGS"]
		if ok && envSlugs != "" {
			slugs = envSlugs
		}
		// handle special 'slugs'
		if slugs == "all" {
			slugs = "sql:select distinct project_slug from mv_subprojects where project_slug is not null and trim(project_slug) != ''"
		}
		var slugsAry []string
		if strings.HasPrefix(slugs, "sql:") {
			var err error
			slugsAry, err = getQuerySlugs(db, debug, slugs[4:])
			if err != nil {
				return err
			}
		} else {
			slugsAry = strings.Split(slugs, ",")
		}
		task[gPrefix+"PROJECT_SLUG"] = slugs

		// Ranges
		ranges := taskDef.TimeRanges
		envRanges, ok := env["TIME_RANGES"]
		if ok && envRanges != "" {
			ranges = envRanges
		}
		// handle special 'ranges'
		if ranges == "all" {
			ranges = "7d,30d,q,ty,y,2y,a,7dp,30dp,qp,typ,yp,2yp"
		}
		rangesAry := strings.Split(ranges, ",")

		// Extra params
		for k, v := range taskDef.ExtraParams {
			task[gPrefix+"PARAM_"+k] = v
		}

		// Extra env
		for k, v := range taskDef.ExtraEnv {
			task[gPrefix+k] = v
		}

		for _, slug := range slugsAry {
			for _, rng := range rangesAry {
				// Final task to execute
				newTask := make(map[string]string)
				for k, v := range task {
					newTask[k] = v
				}
				newTask[gPrefix+"TIME_RANGE"] = rng
				newTask[gPrefix+"PROJECT_SLUG"] = slug
				allTasks = append(allTasks, newTask)
			}
		}
	}
	// FIXME
	lib.Logf("%d tasks\n", len(allTasks))
	if debug {
		for _, task := range allTasks {
			lib.Logf("task: %+v\n", task)
		}
	}
	return nil
}

func sync() error {
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
			lib.Logf("env: %s\n", msg)
			err := fmt.Errorf("%s", msg)
			return err
		}
	}
	connStr, _ := env["CONN"]
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer func() { db.Close() }()
	if debug {
		lib.Logf("db: %+v\n", db)
	}
	path, ok := env["SYNC_PATH"]
	if !ok {
		path = "./"
	}
	fn := path + "calculations.yaml"
	contents, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	var metrics Metrics
	err = yaml.Unmarshal(contents, &metrics)
	if err != nil {
		return err
	}
	if debug {
		lib.Logf("metrics: %+v\n", metrics)
	}
	err = runTasks(db, metrics, debug, env)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	dtStart := time.Now()
	err := sync()
	if err != nil {
		lib.Logf("sync error: %+v\n", err)
	}
	dtEnd := time.Now()
	lib.Logf("time: %v\n", dtEnd.Sub(dtStart))
}
