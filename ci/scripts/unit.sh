#!/usr/bin/env bash

set -ex

echo "Installing golang test tools"
echo ""
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega


echo "Jumping through hoops for golang"
export GOPATH=`pwd`
cp -r firehose-repo splunk-firehose-nozzle
mkdir -p src/github.com/cf-platform-eng
mv splunk-firehose-nozzle src/github.com/cf-platform-eng

cd src/github.com/cf-platform-eng/splunk-firehose-nozzle


echo "Building"
echo ""
go build main.go


echo "Testing"
echo ""
ginkgo -r
