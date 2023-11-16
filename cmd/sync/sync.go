package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	snc "sync"

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
	gSlugsMap   map[string][]string
	gMtx        *snc.Mutex
	gProcessing map[int]map[string]string
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
	// Can also use "top:N", for example "top:5" - it will return top 5 slugs by number of contributions for all time then.
	ProjectSlugs string `yaml:"project_slugs"` // Comma separated list of V3_PROJECT_SLUG values, can also be SQL like `"sql:select distinct project_slug from mv_subprojects"`
	// Can be overwritten with V3_TIME_RANGES env variable
	TimeRanges  string            `yaml:"time_ranges"`  // Comma separated list of time ranges (V3_TIME_RANGE) to calculate or "all" which means all supported time ranges
	ExtraParams map[string]string `yaml:"extra_params"` // map k:v with `V3_PARAM_` prefix skipped in keys, for example: tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'", is_bot='!= true'
	ExtraEnv    map[string]string `yaml:"extra_env"`    // map k:v with `V3_` prefix skipped in keys, for example: DEBUG=1 DATE_FROM=2023-10-01 DATE_TO=2023-11-01
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

	// Mutex
	gMtx = &snc.Mutex{}
	gProcessing = make(map[int]map[string]string)

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
				lib.Logf("%d tasks processing\n", len(gProcessing))
				for idx, task := range gProcessing {
					lib.Logf("heartbeat: task #%d\n", idx)
					lib.Logf("%s\n", prettyPrintTask(task))
				}
				gMtx.Unlock()
			}
		}()
	}

	// process tasks
	thrN := getThreadsNum(debug, env)
	if thrN > 1 {
		ch := make(chan error)
		nThreads := 0
		for i := range allTasks {
			go processTask(ch, i, debug, calcBin, allTasks)
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
			lib.Logf("Final threads join\n")
		}
		for nThreads > 0 {
			err := <-ch
			nThreads--
			if err != nil {
				lib.Logf("error: %+v\n", err)
			}
		}
	} else {
		for i := range allTasks {
			err := processTask(nil, i, debug, calcBin, allTasks)
			if err != nil {
				lib.Logf("error: %+v\n", err)
			}
		}
	}
	return nil
}

func prettyPrintTask(task map[string]string) string {
	var msg string
	offset := len(gPrefix)
	ks := []string{}
	for k := range task {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		msg += fmt.Sprintf("\t%s: %+v\n", k[offset:], task[k])
	}
	return msg
}

func processTask(ch chan error, idx int, debug bool, binCmd string, tasks []map[string]string) error {
	task := tasks[idx]
	gMtx.Lock()
	gProcessing[idx] = task
	gMtx.Unlock()
	defer func() {
		gMtx.Lock()
		delete(gProcessing, idx)
		gMtx.Unlock()
	}()
	if debug {
		lib.Logf("starting task #%d, details:\n", idx)
		lib.Logf("%s\n", prettyPrintTask(task))
	}
	var err error
	dtStart := time.Now()
	res, skipped, err := execCommand(
		debug,
		[]string{binCmd},
		task,
	)
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
		lib.Logf("%s\n", prettyPrintTask(task))
	}
	if debug {
		lib.Logf("task #%d (%+v) executed (skipped or no data: %v), took: %v\noutput/stderr:\n%s\n", idx, task, skipped, dtEnd.Sub(dtStart), res)
	}
	if ch != nil {
		ch <- err
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
