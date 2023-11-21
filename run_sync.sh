#!/bin/bash
# DBG=1
# CWD=/root/go/src/github.com/lukaszgryglicki/calcmetric
# M * * * * VERBOSE=1 CWD=/root/go/src/github.com/lukaszgryglicki/calcmetric /usr/local/bin/run_sync.sh sync.sh /tmp/calcmetric_sync.log >> /tmp/calcmetric_sync.log 2>&1
if ( [ -z "$1" ] || [ -z "$2" ] )
then
  echo "$0: you need to specify: command and a log file"
  exit 1
fi
sha512="run_sync_$(echo -n "$0/$1/$2" | sha256sum | cut -d' ' -f1)"
sha512="/tmp/$sha512"
if [ ! -z "$DBG" ]
then
  echo "code: '$sha512'"
fi
if [ -f "$sha512" ]
then
  if [ ! -z "$DBG" ]
  then
    echo "command '$1 2>&1 >> $2' is running, exiting"
  fi
  exit 2
fi
function cleanup {
  rm -r "$sha512"
  if [ ! -z "$DBG" ]
  then
    echo "$sha512 removed"
  fi
}
trap cleanup EXIT
> $sha512
if [ ! -z "$DBG" ]
then
  echo "start: $1 2>&1 >> $2"
fi
if [ ! -z "$CWD" ]
then
  if [ ! -z "$DBG" ]
  then
    echo "cd $CWD"
  fi
  cd "$CWD" || exit 3
fi
$1 2>&1 >> $2
if [ ! -z "$DBG" ]
then
  echo "end: $1 2>&1 >> $2"
fi
