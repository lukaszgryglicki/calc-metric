#!/bin/bash
if [ -z "${V3_CONN}" ]
then
  echo "$0: you must specify V3_CONN='db connect string'"
  exit 1
fi
export V3_METRIC=contr-lead-acts-all
export V3_TABLE=metric_contr_lead_acts_nbot
export V3_PROJECT_SLUG=sonic-foundation
export V3_TIME_RANGE=30dp
export V3_PARAM_tenant_id="'875c38bd-2b1b-4e91-ad07-0cfbabb4c49f'"
export V3_PARAM_is_bot='!= true'
./calcmetric
