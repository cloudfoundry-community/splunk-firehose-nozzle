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
	go mod vendor

updatedeps:
	go get -u
	go mod tidy
	go mod vendor

initdeps:
	go mod init
	go mod tidy
	go mod vendor

# -gcflags '-N -l' for debug
# -ldflags -w for prod

linux:
	GOOS=linux GOARCH=amd64 make

mac:
	GOOS=darwin GOARCH=amd64 make

build-linux:
	GOOS=linux GOARCH=amd64 make build

build: build-nozzle build-app-dump build-data-gen

debug:
	DEBUG_FLAGS="-gcflags '-N -l'" make build

LDFLAGS="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH) -X main.buildos=$(UNAME)"

build-nozzle: fmt
	go build -o splunk-firehose-nozzle  -ldflags ${LDFLAGS} ${DEBUG_FLAGS} ./main.go

build-app-dump:
	go build -o tools/dump_app_info/dump_app_info ./tools/dump_app_info/dump_app_info.go

build-data-gen:
	go build -o .github/data_gen/data_gen tools/data_gen/data_gen.go

PKGS=$(shell go list ./... | grep -v vendor | grep -v tools | grep -v testing | grep -v "splunk-firehose-nozzle$$")

deploy: deploy-nozzle deploy-data-gen

deploy-nozzle:
	@cf push -f github/workflows/ci_nozzle_manifest.yml -u process --random-route

deploy-data-gen:
	@cf push -f github/workflows/data_gen_manifest.yml -u process -p tools/data_gen --random-route

integration-test: deploy-nozzle deploy-data-generation-app test

testall: test vet race cov

test:
	@go test ${PKGS}

testv:
	@go test -v ${PKGS}

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
