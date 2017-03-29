#!/usr/bin/env bash

set -ex

echo "Fiddle with the go path"
export GOPATH=`pwd`
export PATH=$GOPATH/bin:$PATH

mkdir -p src/github.com/cloudfoundry-community
cp -r source-repo src/github.com/cloudfoundry-community/splunk-firehose-nozzle

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
