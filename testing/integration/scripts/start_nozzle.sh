#!/usr/bin/env bash

FILE_PATH=$1
TIME_INTERVAL=$2

cd $FILE_PATH
source env.sh
./splunk-firehose-nozzle & sleep $TIME_INTERVAL ; kill $!
