#!/bin/bash
if [ -z "${V3_CONN}" ]
then
  echo "$0: you must specify V3_CONN='db connect string'"
  exit 1
fi
export V3_HEARTBEAT=300
# export V3_YAML_PATH='./'
# export V3_BIN_PATH='./'
# export V3_DEBUG=1
export V3_THREADS=8
# export V3_THREADS=3
./sync
echo "Sync done, exit status: $?"
# clear && V3_CONN="`cat ./REPLICA.secret`" ./sync.sh 1>> sync.log 2>> sync.err &
# clear && tail -f sync.???
# ps -axu | grep -E 'sync|calcmetric'
