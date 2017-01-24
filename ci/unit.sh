#!/usr/bin/env bash

set -ex

echo "Fiddle with the go path"
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH

cp -r source-repo/src/splunk-firehose-nozzle splunk-firehose-nozzle
mkdir -p src/github.com/cloudfoundry-community
mv splunk-firehose-nozzle src/github.com/cloudfoundry-community

cd src/github.com/cloudfoundry-community/splunk-firehose-nozzle


echo "Installing Go test tools"
echo ""
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega


echo "Building"
echo ""
go build main.go


echo "Testing"
echo ""
ginkgo -r
