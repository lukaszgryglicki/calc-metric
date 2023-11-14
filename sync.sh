#!/bin/bash
if [ -z "${V3_CONN}" ]
then
  echo "$0: you must specify V3_CONN='db connect string'"
  exit 1
fi
export V3_DEBUG=1
./sync
