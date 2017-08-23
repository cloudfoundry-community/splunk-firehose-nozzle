#!/usr/bin/env bash

set -e

if [ "$0" != "./build.sh" ]; then
  echo "build.sh should be run from within the tile directory"
  exit 1
fi

echo "building go binary"
pushd ..
curdir=`pwd`
go get github.com/cloudfoundry-community/splunk-firehose-nozzle
cd $GOPATH/src/github.com/cloudfoundry-community/splunk-firehose-nozzle && git checkout develop && env GOOS=linux GOARCH=amd64 make build VERSION=1.0
cp $GOPATH/src/github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk-firehose-nozzle ${curdir}/../splunk-firehose-nozzle/
cd ${curdir}
popd

echo "building tile"
tile build