#!/bin/bash
if [ -z "${V3_CONN}" ]
then
  echo "$0: you must specify V3_CONN='db connect string'"
  exit 1
fi
# export V3_YAML_PATH='./'
# export V3_BIN_PATH='./'
# export V3_DEBUG=1
# export V3_THREADS=1
export V3_THREADS=1
./sync
