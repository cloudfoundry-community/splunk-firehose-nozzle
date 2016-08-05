#!/usr/bin/env bash

echo "Installing golang test tools"
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

echo "Building"
cd firehose-repo
go build main.go

echo "Testing"
ginkgo -r
