# calcmetric

This program is used to calculate a metrc and save its values to a dedicated table.

You specify environment variables starting with `V3_` prefix to specify which metric shoudl be calculated.

Those are mandatory parameters that must be specified, see examples in `calcmetric.sh` file:

- `V3_CONN` - database connect string.
- `V3_METRIC` - metric name, for example `contr-lead-acts` it will correspond to its SQL file in `sql/contr-lead-acts.sql`.
- `V3_TABLE` - table name where calculations will be stored. Example: `metric_contr_lead_acts`.
- `V3_PROJECT_SLUG` - specifies project slug to calculate, example: `korg`.
- `V3_TIME_RANGE` - time range to calculate for, allowed values: `7d`, `30d`, `q`, `ty`, `y`, `2y`, `a`, `c`, they mean:
  - `7d` - last week (Mon-Sun, calculated on Mondays or if not calculated yet). *Or we can calculate this every day* if `V3_CALC_WEEK_DAILY` is set.
  - `7dp` - previous last week (Mon-Sun, calculated on Mondays or if not calculated yet).
  - `30d` - last month (calculated only 1st day of a month or if not calculated yet). *Or we can calculate this every day* if `V3_CALC_MONTH_DAILY` is set.
  - `30dp` - previous month (calculated only 1st day of a month or if not calculated yet).
  - `q` - last quarter (calculated only 1st day of a new quarter or if not calculated yet). *Or we can calculate this every day* if `V3_CALC_QUARTER_DAILY` is set.
  - `qp` - previous quarter (calculated only 1st day of a new quarter or if not calculated yet).
  - `ty` - this year (calculated daily) this is from this year 1st of January till today.
  - `typ` - previous periof for this year (if today is 200th day of year then this is 1st of January this year minus 200 days till 1st of January this year).
  - `y` - last year (calculated only 1st day of a new year or if not calculated yet). *Or we can calculate this every day* if `V3_CALC_YEAR_DAILY` is set.
  - `yp` - previous year (calculated only 1st day of a new year or if not calculated yet).
  - `2y` - 2 last years (calculated only 1st day of a new 2 years or if not calculated yet). *Or we can calculate this every day* if `V3_CALC_YEAR2_DAILY` is set.
  - `2yp` - 2 previous years (calculated only 1st day of a new 2 years or if not calculated yet).
  - `a` - all time (no time filter or 1970-01-01 - 2100-01-01) - calculated daily. Note that there is no `ap` as it makes no sense.
  - `c` - custom time range - from `V3_DATE_FROM` to `V3_DATE_TO`, calculated on request.
- Optional `V3_DATE_FROM` and `V3_DATE_TO` become required when `V3_TIME_RANGE` is set to `c` (custome time range).

Those parameters are optional:

- `V3_CALC_WEEK_DAILY` - if this is set, we calculate `7d` and `7dp` every day, instead of Mondays.
- `V3_CALC_MONTH_DAILY` - if this is set, we calculate `30d` and `30dp` every day, instead of 1st days of months.
- `V3_CALC_QUARTER_DAILY` - if this is set, we calculate `q` and `qp` every day, instead of 1st days of quarters.
- `V3_CALC_YEAR_DAILY` - if this is set, we calculate `y` and `yp` every day, instead of 1st days of years.
- `V3_CALC_YEAR2_DAILY` - if this is set, we calculate `2y` and `2yp` every day, instead of 1st days of every 2 years.
- `V3_DATE_FROM` - if `c` date range is used - this is a starting datetime. Format is YYYY-MM-DD. If you specify 'YYYY-MM-DD HH:MI:SS' it will truncate to 'YYYY MM-DD 00:00:00.000' - max resolution is daily.
- `V3_DATE_TO` - if `c` date range is used - this is an ending datetime. Format is YYYY-MM-DD.
- `V3_FORCE_CALC` - if set, then we don't check if given time range is already calculated.
- `V3_LIMIT` - limit rows to this value. This replaces `{{limit}}` in the input query if present.
- `V3_OFFSET` - offset from this value. This replaces `{{offset}}` in the input query if present.
- `V3_DEBUG` - set debug mode.
- `V3_PPT` - `Per-Project-Tables` - means - create tables with `_project_slug` added to their name, we can consider using this for speedup.
- `V3_GUESS_TYPE` - attempt to guess DB type when not specified.
- `V3_INDEXED_COLUMNS` - specify comma separated list of columns where you want to add extra indices.
- `V3_DROP` - drop destination table if exists. This is to support data cleanup.
- `V3_DELETE` - `tr,ps,df,dt` - drop data from destination table for current calculation: each value `tr,ps,df,dt` specifies if `time_range, project_slug, date_from, date_to` keys should be used for deleting. This is to support data cleanup.
- `V3_CLEANUP` - cleanup previous calculations for this time range and project slug *only* after successful calculations of current status.
- `V3_SQL_PATH` - path to metric SQL files, `./sql/` if not specified.
- `V3_PARAM_xyz` - extra params to replace in `SQL` file, for example specifying `V3_PARAM_my_param=my_value` will replace `{{my_param}}` with `my_value` in metric's SQL file.


# Running calcmetric

Example:
- Create your own `REPLICA.secret` - it is gitignored in the repo, you can yse `REPLICA.secret.example` file as a starting point.
- `V3_CONN=[redacted] ./calcmetric.sh` - this runs example calculation, or: `` V3_CONN="`cat ./REPLICA.secret`" ./calcmetric.sh ``.
- Other example scripts are in `./examples/sh/*.sh`.
- Other example metrics SQLs are in `./examples/sql/*.sql`.


Generated tables:

- Generated table will use name from `V3_TABLE`.
- This table will always have the following columns:
  - `project_slug` - will have `V3_PROJECT_SLUG` value.
  - `time_range` - it will be the value passed in `V3_TIME_RANGE`.
  - `date_from`, `date_to` - will have time from and time to values for which a given records were calcualted.
  - `last_calculated_at` - will store the value when this table was last calculated.
  - `row_number` - as returned from the SQL query.
- Table's primary key is `(time_range, project_slug, date_from, date_to, row_number)`.


# Running all calculations

There is an YAML file `calculations.yaml` that specifies all metrics that needs to be calculated, it runs in a loop and checks every single metric and eventually regenerates it if needed.

Each entry in this file is a single invocation of `calcmetric` program. This is handles by `sync` program that will be written to handle this.

This program uses the following environment variables:
- `V3_YAML_PATH` - path to where `calculations.yaml` is, `./` if not specified.
- `V3_BIN_PATH` - path to where `calcmetric` binary is, `./` if not specified.
- `V3_THREADS` - specify number of threads to run in parallel (`sync` will invoke up to that many of `calcmetric` calls in parallel). Empty or zero or negative number will default to numbe rof CPU cores available.
- `V3_HEARTBEAT` - specify number of seconds for heartbeat.


YAML file fields descripution:
- `metric` - Maps to `calcmetric`'s `V3_METRIC`. Metric SQL file to use.
- `table` - Maps to `V3_TABLE` - table to save data to.
- `project_slugs`:
  - Comma separated list of `V3_PROJECT_SLUG` values, can also be SQL like `"sql:select distinct project_slug from mv_subprojects"`.
	- Can be overwritten with `V3_PROJECT_SLUGS` env variable.
	- Can also use `all` which connects to DB and gets all slugs using built-in SQL command.
  - Can also use `top:N`, for example `top:5` - it will return top 5 slugs by number of contributions for the last quarter then.
- `time_ranges`:
  - Comma separated list of time ranges (`V3_TIME_RANGE`) to calculate or `all` which means all supported time ranges excluding `c` (custom).
  - `all-current` means all current time rannges, excluding previous ones (with `p` suffix) and `c` (custom).
	- Can be overwritten with `V3_TIME_RANGES` env variable.
- `extra_params` - YAML map `k:v` with `V3_PARAM_` prefix skipped in keys, for example: `tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'"`, `is_bot='!= true'`.
- `extra_env` - YAML map `k:v` with `V3_` prefix skipped in keys, for example: `DEBUG=1`, `DATE_FROM=2023-10-01`, `DATE_TO=2023-11-01`.


Example run:
- Create your own `REPLICA.secret` - it is gitignored in the repo, you can yse `REPLICA.secret.example` file as a starting point.
- `V3_CONN=[redacted] ./sync.sh` - this runs example sync, or: `` V3_CONN="`cat ./REPLICA.secret`" ./sync.sh ``.
- Other example YAMLS are in `./examples/yaml/*.yaml`.

# Full example how to get contributor leaderboard

It will generate table with data for "contributor leaderboard" table - this will contain all data you need inone table, without need to call a separate JSON calls to get previous period vaues (it already calculates `change from previous`) and totals (it already calculated `percent of total`) plus it even returns numbe rof all contributors that is needed for paging.

To sum up - the table created via single calculation will have that all and a single cube JSON query can get that data for a specified project & time range.
- You run `` V3_CONN="`cat ./REPLICA.secret`" ./calcmetric.sh ``. [calcmetric.sh](https://github.com/lukaszgryglicki/calcmetric/blob/main/calcmetric.sh) ([source](https://github.com/lukaszgryglicki/calcmetric/blob/main/cmd/calcmetric/calcmetric.go)).
- It runs `calcmetric.sh` with DB connect string taken from `REPLICA.secret` file (you can source it from [this example file](https://github.com/lukaszgryglicki/calcmetric/blob/main/REPLICA.secret.example)), it specifies the following parameters:
```
export V3_METRIC=contr-lead-acts-all
export V3_TABLE=metric_contr_lead_acts_all
export V3_PROJECT_SLUG=envoy
export V3_TIME_RANGE=7d
export V3_PARAM_tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'"
export V3_PARAM_is_bot='!= true'
```
- So it runs [./sql/contr-lead-acts-all.sql](https://github.com/lukaszgryglicki/calcmetric/blob/main/sql/contr-lead-acts-all.sql) - this SQL returns data for current, previous period and totals including number of all contributors.
- `calcmetric` will replace all `{{placeholder_variable}}` placeholders within that SQL - thsi is the way it is parametrized.
- `calcmetric` will add `project_slug`, `time_range`, `date_from`, `date_to`, `row_number` columns.
- It will create table like this:
```
crowd=> \d metric_contr_lead_acts_all
                   Table "crowd_public.metric_contr_lead_acts_all"
        Column        |            Type             | Collation | Nullable | Default
----------------------+-----------------------------+-----------+----------+---------
 time_range           | character varying(6)        |           | not null |
 project_slug         | text                        |           | not null |
 last_calculated_at   | timestamp without time zone |           | not null |
 date_from            | date                        |           | not null |
 date_to              | date                        |           | not null |
 row_number           | integer                     |           | not null |
 logo_url             | text                        |           |          |
 memberid             | text                        |           |          |
 platform             | text                        |           |          |
 username             | text                        |           |          |
 is_bot               | boolean                     |           |          |
 contributions        | bigint                      |           |          |
 all_contributions    | bigint                      |           |          | 
 prev_contributions   | bigint                      |           |          |
 percent_total        | numeric                     |           |          |
 change_from_previous | numeric                     |           |          |
 all_contributors     | bigint                      |           |          |
Indexes:
    "metric_contr_lead_acts_all_pkey" PRIMARY KEY, btree (time_range, project_slug, date_from, date_to, row_number)
    "metric_contr_lead_acts_all_project_slug_idx" btree (project_slug)
    "metric_contr_lead_acts_all_time_range_idx" btree (time_range)
```
You can see that this table already has:
- primary key: `(time_range, project_slug, date_from, date_to, row_number)`, so the calculation context is `(time_range, project_slug, date_from, date_to)` - remaining data is for this context + then rows (each have an identity data in this case).
- `contributions` - current contributions.
- `prev_contributions` - value for previous time range.
- `all_contributuions` - all contributions for current context.
- `change_from_previous` - change from previous calculated using the two above, if there are no contributions for the previous period for that contributor, it will return `100` - like 100% more.
- `percent_total` - percent of total contributions: `contributions / all_contributions`.
- other data related to identity in this case, like: `memberid, platform, username, is_bot` and so on.


**NOTE: previously this needed to make at least 3 cube calls (to get current data, previous time range data and to get total counts) - all fo them were generating a very complex Activities cube query which was not based on materialized views and was using very heavy pre-aggregation - now it will be a single call to a single table specifying time range and project and THAT's IT!**


# Bulk calclations

We can also mass-calculate this for multiple projects at once using `sync` tool ([source](https://github.com/lukaszgryglicki/calcmetric/blob/main/cmd/sync/sync.go)):
- You run `` V3_CONN="`cat ./REPLICA.secret`" ./sync.sh ``. [sync.sh](https://github.com/lukaszgryglicki/calcmetric/blob/main/sync.sh).
- It uses [calculations.yaml](https://github.com/lukaszgryglicki/calcmetric/blob/main/calculations.yaml) file that instructs `sync` tool about how shoudl it call `calcmetric` for multiple prohects/time-ranges, etc., thsi si the example contents:
```
---
metrics:
  contr_lead_acts_non_bots:
    metric: contr-lead-activities
    table: metric_contr_lead_nbot
    project_slugs: all
    time_ranges: all-current
    extra_params:
      tenant_id: "'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'"
      is_bot: '!= true'
    extra_env:
      INDEXED_COLUMNS: 'metric'
      LIMIT: '200'
      CLEANUP: y
```
- It specifies metric to use [contr-lead-acts-all](https://github.com/lukaszgryglicki/calcmetric/blob/main/sql/contr-lead-acts-all.sql).
- Output table to save data: `metric_contr_lead_acts_nbot`.
- Then it specifies projects-slugs to run: `all` means that it will execute thsi query to get all slugs:
```
select distinct project_slug from mv_subprojects where project_slug is not null and trim(project_slug) != ''
```
- Specifies which time ranges to run: `all` means all excluding `c` (custom) - as it is not possible to guess all possible YYYY-MM-DD combinations.
- Extra params specifies some additional flags to be passed to `calcmetric` tool.

Example output for calculating Top 10 projects for all time ranges looks like this:
```
2023-11-15 10:39:07: 130 tasks
2023-11-15 10:41:08: task #0 finished in 2m0.785356259s, details:
	PARAM_is_bot: != true
	TIME_RANGE: 7d
	METRIC: contr-lead-acts-all
	TABLE: metric_contr_lead_acts_nbot
	PROJECT_SLUG: cncf
	PARAM_tenant_id: '875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'
2023-11-15 10:41:53: task #1 finished in 2m45.950157151s, details:
	TIME_RANGE: 30d
	TABLE: metric_contr_lead_acts_nbot
	PROJECT_SLUG: cncf
	PARAM_tenant_id: '875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'
	PARAM_is_bot: != true
	METRIC: contr-lead-acts-all
2023-11-15 10:48:11: task #2 finished in 9m3.681881757s, details:
	PROJECT_SLUG: cncf
	PARAM_tenant_id: '875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'
	PARAM_is_bot: != true
	METRIC: contr-lead-acts-all
	TIME_RANGE: q
	TABLE: metric_contr_lead_acts_nbot
```

And generates data like this:
```
crowd=> select project_slug, time_range, max(row_number) from metric_contr_lead_acts_nbot group by project_slug, time_range;
 project_slug | time_range | max
--------------+------------+------
 cncf         | q          | 1280
 cncf         | 30d        |  535
 cncf         | 7d         |  174
(3 rows)
```

# Tracking progress

Example query:
```
select project_slug, time_range, max(row_number) as max_row, count(*) as records, max(last_calculated_at) as calculated_at, max(all_contributions) as all_contributions, max(all_contributors) as all_contributors, avg(percent_total) as avg_percent_total, avg(change_from_previous) as avg_change_from_previous from metric_contr_lead_acts_nbot group by project_slug, time_range order by calculated_at desc limit 50
```

Example results:
```
     project_slug     | time_range | max_row | records |       calculated_at        | all_contributions | all_contributors |   avg_percent_total    | avg_change_from_previous
----------------------+------------+---------+---------+----------------------------+-------------------+------------------+------------------------+--------------------------
 grid-capacity-map    | 7dp        |       3 |       3 | 2023-11-15 14:55:51.877676 |                 6 |                3 |    33.3333333333333300 |      66.6666666666666667
 datashim             | 2yp        |       2 |       2 | 2023-11-15 14:55:47.173656 |                 2 |                2 |    50.0000000000000000 |     100.0000000000000000
 hwameistor           | qp         |      52 |      52 | 2023-11-15 14:55:42.841646 |               888 |               52 |    2.03568953568953575 |      95.9286976575234255
 elyra                | yp         |     129 |     129 | 2023-11-15 14:55:34.765653 |             11484 |              129 | 0.78187650360866080132 |     118.7295731460003990
 grpc                 | y          |    1408 |    1408 | 2023-11-15 14:55:20.841694 |             25273 |             1408 | 0.07222831228439980836 |      90.7243438799638872
 cello                | typ        |      36 |      36 | 2023-11-15 14:55:09.897671 |               474 |               36 |    3.19971870604782000 |      81.5897794664488636
 onechipml            | typ        |       5 |       5 | 2023-11-15 14:55:05.813656 |                40 |                5 |    20.0000000000000000 |     100.0000000000000000
 transact             | typ        |      17 |      17 | 2023-11-15 14:55:02.325684 |               452 |               17 |     5.8823529411764704 |     -43.4742093002456698
 besu                 | 2yp        |     138 |     138 | 2023-11-15 14:54:34.269681 |              5035 |              138 | 0.96368895988946933490 |     100.0000000000000000
 cello                | ty         |      28 |      28 | 2023-11-15 14:54:05.84164  |               418 |               28 |    3.88755980861244006 |      72.0024006061629230
 fabric               | 7d         |      28 |      28 | 2023-11-15 14:53:53.705681 |               151 |               28 |     3.8789025543992434 |     225.0331607001775073
 hwameistor           | 7dp        |      16 |      16 | 2023-11-15 14:53:47.529655 |                53 |               16 |     6.6037735849056607 |      38.5018939393939400
 carina               | yp         |      19 |      19 | 2023-11-15 14:53:45.317626 |               390 |               19 |     5.2631578947368420 |     100.0000000000000000
 elyra                | a          |     294 |     294 | 2023-11-15 14:53:40.521684 |             29455 |              294 | 0.34131391480374190567 |     100.0000000000000000
 gridlabd             | 2y         |      18 |      18 | 2023-11-15 14:53:29.085671 |              1841 |               18 |    5.55555555555555579 |      58.5044833401753987
 envoy                | 2y         |    3429 |    3429 | 2023-11-15 14:53:11.505699 |            156975 |             3429 | 0.03003359251219573044 |     185.1670255143548956
 deepcausality        | a          |       5 |       5 | 2023-11-15 14:52:46.409677 |               251 |                5 |    19.9999999999999970 |     100.0000000000000000
 fledgepower          | yp         |       5 |       5 | 2023-11-15 14:52:42.657647 |              1346 |                5 |   19.99999999999999876 |     100.0000000000000000
 deeprec              | typ        |     149 |     149 | 2023-11-15 14:52:34.157657 |              2752 |              149 | 0.72430544716716090219 |     363.8039826434260792
 everest              | 7dp        |      47 |      47 | 2023-11-15 14:52:08.549688 |               316 |               47 |    2.50471316994344200 |     153.4628225012587043
 wasmedge-runtime     | a          |     474 |     474 | 2023-11-15 14:52:03.941726 |             17593 |              474 | 0.21111436486654045775 |     100.0000000000000000
 aries                | 30d        |     116 |     116 | 2023-11-15 14:51:11.50569  |              1432 |              116 | 0.93912540936235793078 |      81.3100406270565147
 sonic-foundation     | qp         |     752 |     752 | 2023-11-15 14:50:49.969625 |             20177 |              752 | 0.13413867063720119391 |      82.7628511462174101
 cloudevents          | ty         |      78 |      78 | 2023-11-15 14:50:31.113658 |              1587 |               78 |     1.2852826652448580 |      83.5948857424205100
 strimzi              | 7d         |      30 |      30 | 2023-11-15 14:50:18.761682 |               221 |               30 |     3.4389140271493213 |      63.9119809407669233
 artifact-hub         | 7dp        |       6 |       6 | 2023-11-15 14:49:36.629603 |                13 |                6 |    16.6666666666666675 |      62.5000000000000000
 compas               | q          |      18 |      18 | 2023-11-15 14:49:29.621641 |               412 |               18 |    5.55555555555555565 |     166.4113319448440898
 aries                | a          |    1441 |    1441 | 2023-11-15 14:49:18.265672 |            101578 |             1441 | 0.07532558844610396346 |     100.0000000000000000
 caliper              | 2yp        |     252 |     252 | 2023-11-15 14:49:06.985625 |              3724 |              252 | 0.39714507356827442425 |     410.5148008244906190
 grpc                 | 30dp       |     200 |     200 | 2023-11-15 14:49:00.317646 |              2251 |              200 | 0.50444247001332741675 |      78.4809490432637998
 sogno                | 30dp       |       9 |       9 | 2023-11-15 14:48:41.305627 |                25 |                9 |    11.1111111111111111 |      56.4814814814814822
 envoy                | typ        |    1841 |    1841 | 2023-11-15 14:48:35.561685 |             65842 |             1841 | 0.05980276949273705290 |     207.7326518290088081
 kubeedge             | 30dp       |      73 |      73 | 2023-11-15 14:48:26.769663 |               611 |               73 |    1.44384906844831071 |     119.0213367463667023
 genevaers            | 30dp       |      13 |      13 | 2023-11-15 14:48:16.769697 |               141 |               13 |     8.1833060556464809 |      39.8965954361475481
 paketo               | 30d        |      91 |      91 | 2023-11-15 14:48:14.321666 |              5079 |               91 |   1.114479141649844536 |     199.6071324945212279
 bevel                | yp         |     108 |     108 | 2023-11-15 14:47:44.005702 |              2527 |              108 | 1.00067420012018350630 |     116.5294632173947123
 aries                | 2y         |     728 |     728 | 2023-11-15 14:47:39.22966  |             45150 |              728 | 0.14953512710987794441 |     144.1769971102237585
 everest              | 7d         |      46 |      46 | 2023-11-15 14:47:22.897688 |               373 |               46 |     2.2788203753351207 |      95.6820495296170747
 backstage            | yp         |    1431 |    1431 | 2023-11-15 14:47:15.641679 |             46681 |             1431 | 0.07780180004910441416 |     131.2400183317600450
 feast                | a          |     760 |     760 | 2023-11-15 14:45:15.329679 |             32562 |              760 | 0.13164764223069189177 |     100.0000000000000000
 baetyl               | 7dp        |      11 |      11 | 2023-11-15 14:45:09.97364  |                45 |               11 |     9.4949494949494950 |      85.9848484848484851
 openeemeter          | q          |       5 |       5 | 2023-11-15 14:45:06.221657 |                 7 |                5 |    19.9999999999999988 |      29.2857142857142860
 sawtooth             | 2yp        |     328 |     328 | 2023-11-15 14:45:02.637612 |             22604 |              328 | 0.31855466979727309586 |     120.5322085867881973
 risc-v-international | 7d         |      51 |      51 | 2023-11-15 14:43:53.813654 |               183 |               51 |     2.0036429872495447 |      55.2930386494796242
 iroha                | typ        |      81 |      81 | 2023-11-15 14:43:44.565686 |             15064 |               81 |   1.240468650629741077 |     263.0273507589244052
 firefly              | y          |     114 |     114 | 2023-11-15 14:43:39.32162  |              9867 |              114 | 0.91213134691395561153 |     335.6301101960662686
 edgex                | 30d        |      55 |      55 | 2023-11-15 14:43:22.221665 |              1186 |               55 |    1.84271040932086463 |     103.3239008985446886
 cncf                 | 7dp        |     170 |     170 | 2023-11-15 14:43:03.165627 |               865 |              170 | 0.61747704862291735653 |     103.1809917870196555
 ClusterDuck          | a          |      44 |      44 | 2023-11-15 14:42:58.385677 |              1302 |               44 |    2.27272727272727285 |     100.0000000000000000
 sogno                | 7d         |       9 |       9 | 2023-11-15 14:42:51.449673 |                50 |                9 |    11.1111111111111116 |      83.3333333333333333
(50 rows)
```
