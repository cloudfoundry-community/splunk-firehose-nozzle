UNAME := $(shell sh -c 'uname')
COMMIT := $(shell sh -c 'git rev-parse HEAD')
BRANCH := $(shell sh -c 'git rev-parse --abbrev-ref HEAD')
VERSION := "0.1"

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

# Standard build
default: installdeps build install

installdeps:
	glide --debug install

updatedeps:
	glide --debug update

initdeps:
	glide --debug create

# -gcflags -N -l for debug
# -ldflags -w for prod
#
#
linux:
	GOOS=linux GOARCH=amd64 make

mac:
	GOOS=darwin GOARCH=amd64 make

build-linux:
	GOOS=linux GOARCH=amd64 make build

build: build-nozzle

build-nozzle: fmt
	go build -o splunk-firehose-nozzle  -ldflags \
		"-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)" \
		./main.go

PKGS=$(shell go list ./... | grep -v vendor | grep -v scripts | grep -v testing | grep -v "splunk-firehose-nozzle$$")

test:
	go test ${PKGS}

# Run "short" unit tests
test-short:
	go test -short ${PKGS}

vet:
	go vet ${PKGS}

race:
	go test -race ${PKGS}


SRC_CODE=$(shell find . -type f -name "*.go" -not -path "./vendor/*")

fmt:
	gofmt -l -w ${SRC_CODE}


.PHONY: test test-short vet build default
all: fmt build
all_test: test test-short vet
