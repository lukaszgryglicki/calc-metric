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
	ProjectSlugs string `yaml:"project_slugs"` // Comma separated list of V3_PROJECT_SLUG values, can also be SQL like `sql: "select distinct project_slug from mv_subprojects"`
	// Can be overwritten with V3_TIME_RANGES env variable
	TimeRanges  string            `yaml:"time_ranges"`  // Comma separated list of time ranges (V3_TIME_RANGE) to calculate or "all" which means all supporte dtime ranges
	ExtraParams map[string]string `yaml:"extra_params"` // map k:v with `V3_PARAM_` prefix skipped in keys, for example: tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'", is_bot='!= true'
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
