#!/bin/bash
if [ -z "${V3_CONN}" ]
then
  echo "$0: you must specify V3_CONN='db connect string'"
  exit 1
fi
export V3_METRIC=contr-lead-acts
export V3_TABLE=metric_contr_lead_acts
export V3_PROJECT_SLUG=envoy
export V3_TIME_RANGE=30d
export V3_PARAM_tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'"
export V3_PARAM_is_bot='!= true'
# export V3_INDEXED_COLUMNS='is_bot,username,memberid,platform'
# export V3_PPT=y
# export V3_METRIC=contr-lead-acts-total
# export V3_TABLE=metric_contr_lead_acts_total
# export V3_TIME_RANGE=c
# export V3_TIME_RANGE=7d
# export V3_PARAM_is_bot='in (true, false)'
# export V3_PARAM_is_bot_value='m.is_bot'
# export V3_PARAM_is_bot_value='false'
# export V3_LIMIT=20
# export V3_OFFSET=0
# export V3_PATH='./sql/'
# export V3_CALC_WEEK_DAILY=1
# export V3_CALC_MONTH_DAILY=1
# export V3_CALC_QUARTER_DAILY=1
# export V3_CALC_YEAR_DAILY=1
# export V3_CALC_YEAR2_DAILY=1
# export V3_DATE_FROM=2023-10-01
# export V3_DATE_TO=2023-11-01
# export V3_FORCE_CALC=1
# export V3_GUESS_TYPE=1
# export V3_PARAM_my_param="my value"
# export V3_PARAM_type=contributions
# export V3_DROP=1
# export V3_DEBUG=1
./calcmetric
