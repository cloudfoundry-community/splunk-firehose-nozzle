#!/usr/bin/env bash

echo "Installing golang test tools"
echo ""
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

echo "Building"
echo ""
go get github.com/Masterminds/glide

cd firehose-repo
glide install
go build main.go

echo "Testing"
echo ""
ginkgo -r
