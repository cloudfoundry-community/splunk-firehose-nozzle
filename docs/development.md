# Development

## Software Requirements

Make sure you have the following installed on your workstation:

| Software | Version  |
|----------|----------|
| go       | go1.17.x |

Then make sure that all dependent packages are there:

```
$ cd <REPO_ROOT_DIRECTORY>
$ make installdeps
```

## Environment

For development against [bosh-lite](https://github.com/cloudfoundry/bosh-lite),
copy `tools/nozzle.sh.template` to `tools/nozzle.sh` and supply missing values:

```
$ cp tools/nozzle.sh.template tools/nozzle.sh
$ chmod +x tools/nozzle.sh
```

Build project:

```
$ make VERSION=1.3.1
```

Run tests with [Ginkgo](http://onsi.github.io/ginkgo/)

```
$ ginkgo -r
```

Run all kinds of testing

```
$ make test # run all unittest
$ make race # test if there is race condition in the code
$ make vet  # examine GoLang code
$ make cov  # code coverage test and code coverage html report
```

Or run all testings: unit test, race condition test, code coverage etc
```
$ make testall
```

Run app

```
# this will run: go run main.go
$ ./tools/nozzle.sh
```
