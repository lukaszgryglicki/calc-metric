package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	snc "sync"

	"github.com/lib/pq"
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
	gSlugsMap    map[string][]string
	gMtx         *snc.Mutex
	gProcessing  map[int]map[string]string
	gTaskIndices map[string]map[int]struct{}
)

// Metrics contain all metrics to calculate
type Metrics struct {
	Metrics map[string]Metric `yaml:"metrics"`
}

// Metric contains details about how given metric shoudl be calculated
// More details in README.md
type Metric struct {
	Metrics []string `yaml:"metrics"` // Maps to V3_METRIC - array of strings - there can be > 1 metric to be calculated for this
	Table   string   `yaml:"table"`   // Maps to V3_TABLE
	// Can be overwritten with V3_PROJECT_SLUGS env variable
	// Can also use "all" which connects to DB and gets all slugs using built-in SQL command
	// Can also use "top:N", for example "top:5" - it will return top 5 slugs by number of contributions for all time then.
	ProjectSlugs string `yaml:"project_slugs"` // Comma separated list of V3_PROJECT_SLUG values, can also be SQL like `"sql:select distinct project_slug from mv_subprojects"`
	// Can be overwritten with V3_TIME_RANGES env variable
	TimeRanges  string            `yaml:"time_ranges"`  // Comma separated list of time ranges (V3_TIME_RANGE) to calculate or "all" which means all supported time ranges
	ExtraParams map[string]string `yaml:"extra_params"` // map k:v with `V3_PARAM_` prefix skipped in keys, for example: tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'", is_bot='!= true'
	ExtraEnv    map[string]string `yaml:"extra_env"`    // map k:v with `V3_` prefix skipped in keys, for example: DEBUG=1 DATE_FROM=2023-10-01 DATE_TO=2023-11-01
	// Specify how often given metric should be run, you can spacify any golang duration for this, for example "48h"
	// it will check if last successful sync was > "48h" ago and only run then.
	MaxFrequency string `yaml:"max_frequency"`
}

func getQuerySlugs(db *sql.DB, debug bool, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	ary, ok := gSlugsMap[query]
	if ok {
		if debug {
			lib.Logf("returing result %+v for '%s' query from cache\n", ary, query)
		}
		return ary, nil
	}
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
	gSlugsMap[query] = slugs
	return slugs, nil
}

func logCommand(cmdAndArgs []string, env map[string]string) {
	lib.Logf("command, arguments, environment:\n%+v\n%+v\n", cmdAndArgs, env)
}

func execCommand(debug bool, cmdAndArgs []string, env map[string]string) (string, bool, error) {
	// Execution time
	dtStart := time.Now()

	// command & arguments
	command := cmdAndArgs[0]
	arguments := cmdAndArgs[1:]
	if debug {
		var args []string
		for _, arg := range cmdAndArgs {
			argLen := len(arg)
			if argLen > 0x200 {
				arg = arg[0:0x100] + "..." + arg[argLen-0x100:argLen]
			}
			if strings.Contains(arg, " ") {
				args = append(args, "'"+arg+"'")
			} else {
				args = append(args, arg)
			}
		}
		lib.Logf("%s\n", strings.Join(args, " "))
	}
	// prepare command
	cmd := exec.Command(command, arguments...)
	// Set its env
	if len(env) > 0 {
		newEnv := os.Environ()
		for key, value := range env {
			newEnv = append(newEnv, key+"="+value)
		}
		cmd.Env = newEnv
	}
	// capture stdout & stderr
	var (
		stdOut bytes.Buffer
		stdErr bytes.Buffer
	)
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut

	// start command
	err := cmd.Start()
	if err != nil {
		logCommand(cmdAndArgs, env)
		return "", false, err
	}
	// wait for command to finish
	skipped := false
	err = cmd.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			rCode := exiterr.ExitCode()
			if rCode == 66 {
				err = nil
				skipped = true
			}
		}
	}
	if err != nil {
		logCommand(cmdAndArgs, env)
		return "stdout:\n" + stdOut.String() + "\nstderr: " + stdErr.String(), skipped, err
	}

	if debug {
		info := strings.Join(cmdAndArgs, " ")
		lenInfo := len(info)
		if lenInfo > 0x280 {
			info = info[0:0x140] + "..." + info[lenInfo-0x140:lenInfo]
		}
		dtEnd := time.Now()
		lib.Logf("%s ... %+v\n", info, dtEnd.Sub(dtStart))
	}
	return "stdout:\n" + stdOut.String() + "\nstderr: " + stdErr.String(), skipped, nil
}

func getThreadsNum(debug bool, env map[string]string) int {
	threads, ok := env["THREADS"]
	if ok && threads != "" {
		nThreads, err := strconv.Atoi(threads)
		if err == nil && nThreads > 0 {
			if debug {
				lib.Logf("using environment specified threads count: %d\n", nThreads)
			}
			return nThreads
		}
		if err != nil {
			lib.Logf("error parsing threads number from '%s': %+v\n", threads, err)
		}
	}
	thrN := runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	if debug {
		lib.Logf("using threads count: %d as reported by golang runtime\n", thrN)
	}
	return thrN
}

func createMetricLastSyncTable(db *sql.DB) error {
	createTable := `create table metric_last_sync(
  metric_name text not null,
  last_synced_at timestamp not null,
  primary key(metric_name)
);
  `
	_, err := db.Exec(createTable)
	if err != nil {
		lib.QueryOut(createTable, []interface{}{}...)
		return err
	}
	return nil
}

func markAsDone(db *sql.DB, task string) {
	// insert into metric_last_sync(metric_name, last_synced_at) values ('metric-name', now()) on conflict(metric_name) do update set last_synced_at = excluded.last_synced_at;
	// metric_name is: key:table:metric (key - calculations.yaml metric key/name, table: given key's entry table, metric: given key's entry one of metrics values.
	sqlQuery := `insert into metric_last_sync(metric_name, last_synced_at) values ($1, now()) on conflict(metric_name) do update set last_synced_at = excluded.last_synced_at`
	args := []interface{}{task}
	_, err := db.Exec(sqlQuery, args...)
	if err != nil {
		lib.Logf("error setting last_synced_at for '%s': %+v\n", task, err)
		lib.QueryOut(sqlQuery, args...)
	}
}

// returns last synced date and whatever we need to do sync now or not
func checkFrequency(db *sql.DB, task string, freq time.Duration, debug bool) (time.Time, bool, error) {
	// metric_last_sync(metric_name, last_synced_at)
	now := time.Now()
	sqlQuery := `select last_synced_at from metric_last_sync where metric_name = $1`
	args := []interface{}{task}
	if debug {
		lib.Logf("executing sql: %s\nwith args: %+v\n", sqlQuery, args)
	}
	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			errName := e.Code.Name()
			if errName == "undefined_table" {
				lib.Logf("table metric_last_sync does not exist yet, creating it and assuming nothing was synced yet.\n")
				err := createMetricLastSyncTable(db)
				return now, true, err
			}
			lib.QueryOut(sqlQuery, args...)
			return now, true, err
		default:
			lib.QueryOut(sqlQuery, args...)
			return now, true, err
		}
	}
	defer func() { _ = rows.Close() }()
	var (
		lastCalc time.Time
		fetched  bool
	)
	for rows.Next() {
		err := rows.Scan(&lastCalc)
		if err != nil {
			return now, true, err
		}
		fetched = true
	}
	err = rows.Err()
	if err != nil {
		return now, true, err
	}
	if fetched {
		age := now.Sub(lastCalc)
		needsRecalc := age > freq
		lib.Logf("last calcualted date for '%s' metric is %+v, this gives %+v age, frequency is %+v, so need recalc is %v\n", task, lastCalc, age, freq, needsRecalc)
		return lastCalc, needsRecalc, nil
	}
	lib.Logf("there is no calculated report for '%s' metric yet, assuming it needs calculations\n", task)
	return now, true, nil
}

func runTasks(db *sql.DB, metrics Metrics, debug bool, env map[string]string) error {
	path, ok := env["BIN_PATH"]
	if !ok {
		path = "./"
	}
	calcBin := path + "calcmetric"
	if debug {
		lib.Logf("will use '%s' binary to calculate metrics\n", calcBin)
	}
	allTasks := []map[string]string{}
	for taskName, taskDef := range metrics.Metrics {
		// Task table
		table := taskDef.Table

		// Metrics to run
		var metrics []string

		// Check frequency from "metric_last_sync" table if defined
		maxFreq := strings.TrimSpace(taskDef.MaxFrequency)
		if maxFreq != "" {
			freq, err := time.ParseDuration(maxFreq)
			if err != nil {
				return err
			}
			for _, metric := range taskDef.Metrics {
				metricName := strings.TrimSpace(metric)
				tName := taskName + ":" + table + ":" + metricName
				lastRun, shouldRun, err := checkFrequency(db, tName, freq, debug)
				if err != nil {
					return err
				}
				if !shouldRun {
					lib.Logf("skipping running '%s' due to frequency check: %s/%+v, last run: %+v\n", taskName, maxFreq, freq, lastRun)
					continue
				}
				metrics = append(metrics, metricName)
			}
		} else {
			metrics = taskDef.Metrics
		}

		// Task
		task := make(map[string]string)

		// Slugs
		slugs := taskDef.ProjectSlugs
		envSlugs, ok := env["PROJECT_SLUGS"]
		if ok && envSlugs != "" {
			slugs = envSlugs
		}
		// handle special 'slugs'
		if slugs == "all" {
			slugs = "sql:select distinct project_slug from mv_subprojects where project_slug is not null and trim(project_slug) != ''"
		} else if strings.HasPrefix(slugs, "top:") {
			top, err := strconv.Atoi(slugs[4:])
			if err != nil {
				return err
			}
			if top <= 0 {
				top = 1
			}
			slugs = fmt.Sprintf(
				`sql:select i.project_slug from (select p.project_slug, count(a.id) as acts from activities a, mv_subprojects p
             where a.segmentId = p.id and a.timestamp >= now() - '3 months'::interval and p.project_slug is not null
             and trim(p.project_slug) != '' group by p.project_slug order by acts desc limit %d) i`,
				top,
			)
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
		} else if ranges == "all-current" {
			ranges = "7d,30d,q,ty,y,2y,a"
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

		// Table
		task[gPrefix+"TABLE"] = table

		// Main loop creating all tasks to execute
		nMetrics := len(metrics)
		nSlugs := len(slugsAry)
		nRanges := len(rangesAry)
		nItems := nMetrics * nSlugs * nRanges
		lib.Logf("entry '%s' has %d metrics, %d project slugs, %d time-ranges ranges: %d tasks\n", taskName, nMetrics, nSlugs, nRanges, nItems)
		for _, metric := range metrics {
			metricName := strings.TrimSpace(metric)
			for _, slug := range slugsAry {
				for _, rng := range rangesAry {
					// Final task to execute
					newTask := make(map[string]string)
					for k, v := range task {
						newTask[k] = v
					}
					newTask[gPrefix+"METRIC"] = metricName
					newTask[gPrefix+"TIME_RANGE"] = rng
					newTask[gPrefix+"PROJECT_SLUG"] = slug
					newTask["TASK_NAME"] = taskName + ":" + table + ":" + metricName
					allTasks = append(allTasks, newTask)
				}
			}
		}
	}
	lib.Logf("%d tasks\n", len(allTasks))
	if debug {
		for _, task := range allTasks {
			lib.Logf("task: %+v\n", task)
		}
	}
	// randomize tasks so they will refer to random different time ranges/slugs at calculations
	// seems to be faster when running using multiple threads
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allTasks), func(i, j int) { allTasks[i], allTasks[j] = allTasks[j], allTasks[i] })

	// handle lists of indices per task name, to know when a given task is fully finished
	gTaskIndices = make(map[string]map[int]struct{})
	for i, task := range allTasks {
		taskName := task["TASK_NAME"]
		_, ok := gTaskIndices[taskName]
		if !ok {
			gTaskIndices[taskName] = make(map[int]struct{})
		}
		gTaskIndices[taskName][i] = struct{}{}
	}
	if debug {
		lib.Logf("task name to index mapping:\n%+v\n", gTaskIndices)
	}

	// Mutex
	gMtx = &snc.Mutex{}
	gProcessing = make(map[int]map[string]string)

	// Retry
	retry := 0
	rs, ok := env["RETRY"]
	if ok && rs != "" {
		r, err := strconv.Atoi(rs)
		if err != nil {
			return err
		}
		if r > 0 {
			retry = r
			lib.Logf("set retry to: %d\n", retry)
		}
	}

	// Heartbeat
	hbi := 0
	hb, ok := env["HEARTBEAT"]
	if ok && hb != "" {
		hb, err := strconv.Atoi(hb)
		if err != nil {
			return err
		}
		if hb > 0 {
			hbi = hb
		}
	}
	if hbi > 0 {
		go func() {
			for true {
				time.Sleep(time.Duration(hbi) * time.Second)
				gMtx.Lock()
				lib.Logf("heartbeat %d tasks processing\n", len(gProcessing))
				for idx, task := range gProcessing {
					lib.Logf("running task #%d\n", idx)
					lib.Logf("%s\n", prettyPrintTask(idx, task))
				}
				lib.Logf("heartbeat ends\n")
				gMtx.Unlock()
			}
		}()
	}

	// Signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-sigs
			gMtx.Lock()
			lib.Logf("signal(%d): %d tasks processing\n", sig, len(gProcessing))
			for idx, task := range gProcessing {
				lib.Logf("running task #%d\n", idx)
				lib.Logf("%s\n", prettyPrintTask(idx, task))
			}
			lib.Logf("signal ends\n")
			gMtx.Unlock()
		}
	}()

	_, dryRun := env["DRY_RUN"]
	if dryRun {
		lib.Logf("running in dry-run mode.\n")
	}

	// process tasks
	thrN := getThreadsNum(debug, env)
	numTasks := len(allTasks)
	if thrN > 1 {
		ch := make(chan error)
		nThreads := 0
		for i := range allTasks {
			if i > 0 && i%50 == 0 {
				lib.Logf("on %d/%d task\n", i, numTasks)
			}
			go processTask(ch, db, i, retry, debug, dryRun, calcBin, allTasks)
			nThreads++
			if nThreads == thrN {
				err := <-ch
				nThreads--
				if err != nil {
					lib.Logf("error: %+v\n", err)
				}
			}
		}
		if debug {
			lib.Logf("Final %d threads join\n", nThreads)
		}
		for nThreads > 0 {
			err := <-ch
			nThreads--
			if debug {
				lib.Logf("%d threads left\n", nThreads)
			}
			if err != nil {
				lib.Logf("error: %+v\n", err)
			}
		}
	} else {
		for i := range allTasks {
			if i > 0 && i%50 == 0 {
				lib.Logf("on %d/%d task\n", i, numTasks)
			}
			err := processTask(nil, db, i, retry, debug, dryRun, calcBin, allTasks)
			if err != nil {
				lib.Logf("error: %+v\n", err)
			}
		}
	}
	return nil
}

func prettyPrintTask(idx int, task map[string]string) string {
	var msg string
	offset := len(gPrefix)
	ks := []string{}
	ti := "#" + strconv.Itoa(idx) + ":\n"
	for k, v := range task {
		if k == "TASK_NAME" {
			msg = ti + v
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		msg += ti + fmt.Sprintf("\t%s: %+v\n", k[offset:], task[k])
	}
	return msg
}

func processTask(ch chan error, db *sql.DB, idx, retry int, debug, dryRun bool, binCmd string, tasks []map[string]string) error {
	var (
		res     string
		skipped bool
		err     error
	)
	task := tasks[idx]
	taskName := task["TASK_NAME"]
	gMtx.Lock()
	gProcessing[idx] = task
	gMtx.Unlock()
	defer func() {
		gMtx.Lock()
		defer func() {
			gMtx.Unlock()
			if ch != nil {
				ch <- err
			}
		}()
		if err != nil {
			lib.Logf("task #%d failed, so not marking it as done\n", idx)
			lib.Logf("%s\n", prettyPrintTask(idx, task))
			return
		}
		delete(gProcessing, idx)
		_, ok := gTaskIndices[taskName]
		if ok {
			delete(gTaskIndices[taskName], idx)
			if len(gTaskIndices[taskName]) == 0 {
				markAsDone(db, taskName)
				lib.Logf("task group '%s' done\n", taskName)
			}
		}
	}()
	if debug {
		lib.Logf("starting task #%d, details:\n", idx)
		lib.Logf("%s\n", prettyPrintTask(idx, task))
	}
	dtStart := time.Now()
	if dryRun {
		res, skipped, err = "dry-run", false, nil
	} else {
		for trial := 0; trial <= retry; trial++ {
			if trial > 0 {
				lib.Logf("retry #%d for task #%d, details:\n", retry, idx)
				lib.Logf("%s\n", prettyPrintTask(idx, task))
			}
			res, skipped, err = execCommand(
				debug,
				[]string{binCmd},
				task,
			)
			if err == nil {
				break
			}
		}
	}
	dtEnd := time.Now()
	took := dtEnd.Sub(dtStart)
	if err != nil {
		msg := fmt.Sprintf("task #%d (%+v) failed (took %v): %+v: %s\n", idx, task, dtEnd.Sub(dtStart), err, res)
		if debug {
			lib.Logf("%s\n", msg)
		}
		err = fmt.Errorf("%s", msg)
	} else {
		lib.Logf("task #%d finished in %v (skipped or no data: %v), details:\n", idx, took, skipped)
		lib.Logf("%s\n", prettyPrintTask(idx, task))
	}
	if debug {
		lib.Logf("task #%d (%+v) executed (skipped or no data: %v), took: %v\noutput/stderr:\n%s\n", idx, task, skipped, dtEnd.Sub(dtStart), res)
	}
	return err
}

func sync() error {
	gSlugsMap = make(map[string][]string)
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
	path, ok := env["YAML_PATH"]
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
