# calc-metric

This program is used to calculate a metrc and save its values to a dedicated table.

You specify environment variables starting with `V3_` prefix to specify which metric shoudl be calculated.

Those are mandatory parameters that must be specified:

- `V3_CONN` - database connect string.
- `V3_METRIC` - metric name, it will correspond to its SQL file in `sql/metric.sql`.
- `V3_TABLE` - table name where calculations will be stored. Example: `leaderboard`.
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
  - `y` - last year (calculated daily).
  - `yp` - previous year (2 years ago until 1 year ago).
  - `a` - all time (no time filter or 1970-01-01 - 2100-01-01) - calculated daily. Note that there is no `ap` as it makes no sense.
  - `c` - custom time range - from `V3_DATE_FROM` to `V3_DATE_TO`, calculated on request.
- `V3_CALC_WEEK_DAILY` - if this is set, we calculate `7d` and `7dp` every day, instead of Mondays.
- `V3_CALC_MONTH_DAILY` - if this is set, we calculate `30d` and `30dp` every day, instead of 1st days of months.
- `V3_CALC_QUARTER_DAILY` - if this is set, we calculate `q` and `qp` every day, instead of 1st days of quarters.
- `V3_DATE_FROM` - if `c` date range is used - thsi is a starting date. Format is YYYY-MM-DD.
- `V3_DATE_TO` - if `c` date range is used - thsi is an ending date (including ythat date). Format is YYYY-MM-DD.
- `V3_FORCE_CALC` - if set, then we don't check if given time range is already calculated.
- `V3_LIMIT` - limit rows to this value.
- `V3_OFFSET` - offset from this value.
- `V3_DEBUG` - set debug mode.
- `V3_PATH` - path to metric SQL files, `./sql/` if not specified.
- `V3_PARAM_xyz` - extra params to replace in `SQL` file, for example specifying `V3_PARAM_my_param=my_value` will replace `{{my_param}}` with `my_value` in metric's SQL file.

Example:
- `V3_CONN=[redacted] ./calcparams.sh` - this runs example calculation, or: `` V3_CONN="`cat ./REPLICA.secret`" ./calcmetric.sh ``.


Generated tables:
- Generated table will use name from `V3_TABLE`.
- This table will always have the following columns:
  - `time_range` - it will be the value passed in `V3_TIME_RANGE`.
  - `last_calculated_at` - will store the value when this table was last calculated.
  - `date_from`, `date_to` - will have time from and time to values for which a given records were calcualted.
