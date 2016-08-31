## Splunk Nozzle

### Setup

The Nozzle requires a uaa user with the scope `doppler.firehose`. You can
* Add the user manually using [uaac](https://github.com/cloudfoundry/cf-uaac)
* Add a new client to the deployment manifest; see [uaa.clients](https://github.com/cloudfoundry/uaa-release/blob/master/jobs/uaa/spec)
* Run `provision/provision.go`; see `./scripts/provision.sh.template`

Full client details
```
splunk-firehose-nozzle
  scope: openid oauth.approvals doppler.firehose
  resource_ids: none
  authorized_grant_types: client_credentials
  autoapprove: 
  action: none
  authorities: oauth.login doppler.firehose
```

### Development

#### Software Requirements

Make sure you have the following installed on your workstation:

| Software | Version
| --- | --- |
| go | go1.6.x
| glide | 0.11.x

Then install all dependent packages via [Glide](https://glide.sh/):
```
$ cd <REPO_ROOT_DIRECTORY>
$ glide install
```

#### Environment

For development against [bosh-lite](https://github.com/cloudfoundry/bosh-lite),
copy `scripts/nozzle.sh.template` to `scripts/nozzle.sh` and supply missing values:
```
$ cp script/dev.sh.template scripts/nozzle.sh
$ chmod +x scripts/nozzle.sh
```

Build project:
```
$ go build main.go
```

Reinstall dependencies:
```
glide update
```

Update dependencies (without `strip-vcs` you'll end up with submodules):
```
glide install --strip-vendor --strip-vcs --update-vendored
```

Run tests with [Ginkgo](http://onsi.github.io/ginkgo/)
```
$ ginkgo -r
```

Run app
```
# this will run: go run main.go
$ ./scripts/dev.sh
```

#### CI

https://ci.run-01.haas-26.pez.pivotal.io/pipelines/splunk-firehose-nozzle

### Exploring Events

Here are a few Splunk queries to show the events you might want to dashboard

```
eventType=ValueMetric
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```

```
eventType=CounterEvent
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```
