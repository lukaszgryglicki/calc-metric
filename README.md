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
- `V3_CALC_WEEK_DAILY` - if this is set, we calculate `7d` and `7dp` every day, instead of Mondays.
- `V3_CALC_MONTH_DAILY` - if this is set, we calculate `30d` and `30dp` every day, instead of 1st days of months.
- `V3_CALC_QUARTER_DAILY` - if this is set, we calculate `q` and `qp` every day, instead of 1st days of quarters.
- `V3_CALC_YEAR_DAILY` - if this is set, we calculate `y` and `yp` every day, instead of 1st days of years.
- `V3_CALC_YEAR2_DAILY` - if this is set, we calculate `2y` and `2yp` every day, instead of 1st days of every 2 years.
- `V3_DATE_FROM` - if `c` date range is used - this is a starting datetime. Format is YYYY-MM-DD. If you specify 'YYYY-MM-DD HH:MI:SS' it will truncate to 'YYYY MM-DD 00:00:00.000' - max resolution is daily.
- `V3_DATE_TO` - if `c` date range is used - this is an ending datetime. Format is YYYY-MM-DD.
- `V3_FORCE_CALC` - if set, then we don't check if given time range is already calculated.
- `V3_LIMIT` - limit rows to this value.
- `V3_OFFSET` - offset from this value.
- `V3_DEBUG` - set debug mode.
- `V3_PPT` - `Per-Project-Tables` - meas - create tables with `_project_slug` added to their name, we can consider using this for speedup.
- `V3_GUESS_TYPE` - attempt to guess DB type when not specified.
- `V3_DROP` - drop destination table if exists. This is to support data cleanup.
- `V3_PATH` - path to metric SQL files, `./sql/` if not specified.
- `V3_PARAM_xyz` - extra params to replace in `SQL` file, for example specifying `V3_PARAM_my_param=my_value` will replace `{{my_param}}` with `my_value` in metric's SQL file.


# Running

Example:
- Create your own `REPLICA.secret` - it is gitignored in the repo, you can yse `REPLICA.secret.example` file as a starting point.
- `V3_CONN=[redacted] ./calcmetric.sh` - this runs example calculation, or: `` V3_CONN="`cat ./REPLICA.secret`" ./calcmetric.sh ``.


Generated tables:
- Generated table will use name from `V3_TABLE`.
- This table will always have the following columns:
  - `time_range` - it will be the value passed in `V3_TIME_RANGE`.
  - `last_calculated_at` - will store the value when this table was last calculated.
  - `date_from`, `date_to` - will have time from and time to values for which a given records were calcualted.
  - `project_slug` - will have `V3_PROJECT_SLUG` value.
  - `row_number` - as returned from the SQL query.


# Running all calculations

There is a file `calculations.yaml` that specifies all metrics that needs to be calculated, it runs in a loop and checks every single metric and eventually regenerates it if needed.

Each entry in this file is a single invocation of `calcmetric` program. This is handles by `sync` program that will be written to handle this.

This program uses the following environment variables:
- `V3_SYNC_PATH` - path to where `calculations.yaml` is, `./` if not specified.
- `V3_CALC_PATH` - path to where `calcmetric` binary is, `./` if not specified.

Example run:
- Create your own `REPLICA.secret` - it is gitignored in the repo, you can yse `REPLICA.secret.example` file as a starting point.
- `V3_CONN=[redacted] ./sync.sh` - this runs example sync, or: `` V3_CONN="`cat ./REPLICA.secret`" ./sync.sh ``.
