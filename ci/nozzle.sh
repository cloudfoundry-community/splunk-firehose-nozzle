#!/bin/sh

# Workaround to make sure dependency service is up and running
sleep 60

branch=${NOZZLE_BRANCH:-master}

# Build nozzle
go get -d github.com/cloudfoundry-community/splunk-firehose-nozzle
cd /go/src/github.com/cloudfoundry-community/splunk-firehose-nozzle
git checkout ${branch}
make build


# Run proc_monitor
echo "Run proc monitor"
git clone https://github.com/chenziliang/proc_monitor && git checkout develop
cd proc_monitor
screen -S proc_monitor -m -d python proc_monitor.py
cd ..

duration=${NOZZLE_DURATION:-1200}

# Run perf test
echo "Run nozzle perf tests"
python ci/perf.py --run nozzle --duration ${duration}
