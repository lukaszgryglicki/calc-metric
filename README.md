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

Example run:
- Create your own `REPLICA.secret` - it is gitignored in the repo, you can yse `REPLICA.secret.example` file as a starting point.
- `V3_CONN=[redacted] ./sync.sh` - this runs example sync, or: `` V3_CONN="`cat ./REPLICA.secret`" ./sync.sh ``.
- Other example YAMLS are in `./examples/yaml/*.yaml`.

# Full example how to get contributor leaderboard

It will generate table with data for "contributor leaderboard" table - thsi will contain all data you need inone table, without need to call a separate JSON calls to get previous period vaues (it already calculates `change from previous`) and totals (it already calculated `percent of total`) plus it even returns numbe rof all contributors that is needed for paging.

To sum up - the table created via single calculation will have that all and a single cube JSON query can get that data for a specified project & time range.
- You run `V3_CONN="`cat ./REPLICA.secret`" ./calcmetric.sh`. [calcmetric.sh](https://github.com/lukaszgryglicki/calcmetric/blob/main/calcmetric.sh).
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
- primary key: `(time_range, project_slug, date_from, date_to, row_number)`, so the calculation context is `(time_range, project_slug, date_from, date_to)` - remainign data is for this context + then rows (each have an identity data in this case).
- `contributions` - current contributions.
- `prev_contributions` - value for previous time range.
- `all_contributuions` - all contributions for current context.
- `change_from_previous` - change from previous calculated using the two above, if there are no contributions for the previous period for that contributor, it will return `100` - like 100% more.
- `percent_total` - percent of total contributions: `contributions / all_contributions`.
- other data related to identity in this case, like: `memberid, platform, username, is_bot` and so on.


NOTE: previously this needed to make at least 3 cube calls (to get current data, previous time range data and to get total counts) - all fo them were generating a very complex Activities cube query which was not based on materialized views and was using very heavy pre-aggregation - now it will be a single call to a single table specifying time range and project and THAT's IT!


We can also mass-calculate this for multiple projects at once using `sync` tool:
- You run `V3_CONN="`cat ./REPLICA.secret`" ./sync.sh`. [sync.sh](https://github.com/lukaszgryglicki/calcmetric/blob/main/sync.sh).
- It uses [calculations.yaml](https://github.com/lukaszgryglicki/calcmetric/blob/main/calculations.yaml) file that instructs `sync` tool about how shoudl it call `calcmetric` for multiple prohects/time-ranges, etc., thsi si the example contents:
```
---
metrics:
  contr_lead_acts_non_bots:
    metric: contr-lead-acts-all
    table: metric_contr_lead_acts_nbot
    project_slugs: all
    time_ranges: all
    extra_params:
      tenant_id: "'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'"
      is_bot: '!= true'
```
- It specifies metric to use [contr-lead-acts-all](https://github.com/lukaszgryglicki/calcmetric/blob/main/sql/contr-lead-acts-all.sql).
- Output table to save data: `metric_contr_lead_acts_nbot`.
- Then it specifies projects-slugs to run: `all` means that it will execute thsi query to get all slugs:
```
select distinct project_slug from mv_subprojects where project_slug is not null and trim(project_slug) != ''
```
- Specifies which time ranges to run: `all` means all excluding `c` (custom) - as it is not possible to guess all possible YYYY-MM-DD combinations.
- Extra params specifies some additional flags to be passed to `calcmetric` tool.

