UNAME := $(shell sh -c 'uname')
COMMIT := $(shell sh -c 'git rev-parse HEAD')
BRANCH := $(shell sh -c 'git rev-parse --abbrev-ref HEAD')

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

# Standard build
default: installdeps build

installdeps:
	glide --debug install

updatedeps:
	glide --debug update

initdeps:
	glide --debug create

stripvendor:
	glide --debug install --strip-vendor

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

LDFLAGS="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH) -X main.buildos=$(UNAME)"

build-nozzle: fmt
	go build -o splunk-firehose-nozzle  -ldflags ${LDFLAGS} ./main.go

PKGS=$(shell go list ./... | grep -v vendor | grep -v scripts | grep -v testing | grep -v "splunk-firehose-nozzle$$")


testall: test vet race cov

test:
	@go test ${PKGS}

# Run "short" unit tests
test-short:
	@go test -short ${PKGS}

vet:
	@go vet ${PKGS}

race:
	@go test -race ${PKGS}

cov:
	@rm -f coverage-all.out
	@echo "mode: cover" > coverage-all.out
	$(foreach pkg,$(PKGS),\
		go test -coverprofile=coverage.out -cover -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	@go tool cover -html=coverage-all.out

SRC_CODE=$(shell find . -type f -name "*.go" -not -path "./vendor/*")

fmt:
	@gofmt -l -w ${SRC_CODE}

all: installdeps testall build
