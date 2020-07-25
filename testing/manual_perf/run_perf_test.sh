#!/usr/bin/env bash

source testing/manual_perf/env.sh

#run nozzle
make build-nozzle
./splunk-firehose-nozzle &
nozzle_pid=$!

# update datagen config
sed -i "s/EPS.*/EPS: $EPS/" testing/manual_perf/data_gen_manifest.yml
sed -i "s/TOTAL_EVENTS.*/TOTAL_EVENTS: $TOTAL_EVENTS/" testing/manual_perf/data_gen_manifest.yml
sed -i "s/instances.*/instances: $DATAGEN_INSTANCE/" testing/manual_perf/data_gen_manifest.yml
# deploy datagen
cf push -f testing/manual_perf/data_gen_manifest.yml -u process -p tools/data_gen --random-route

# Run nozzle for a duration and kill it and delete datagen
sleep $NOZZLE_DURATION
kill $nozzle_pid
cf delete -f data_gen