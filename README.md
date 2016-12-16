## Splunk Nozzle

Cloud Foundry Firehose-to-Splunk Nozzle

### Setup

The Nozzle requires a user with the scope `doppler.firehose` and 
`cloud_controller.admin_read_only` (the latter is only required if `ADD_APP_INFO` is true). 
You can either
* Add the user manually using [uaac](https://github.com/cloudfoundry/cf-uaac)
* Add a new user to the deployment manifest; see [uaa.scim.users](https://github.com/cloudfoundry/uaa-release/blob/master/jobs/uaa/spec)

Manifest example:
```
uaa:
  scim:
    users:
      - splunk-firehose|password123|cloud_controller.admin_read_only,doppler.firehose
```

### Development

#### Software Requirements

Make sure you have the following installed on your workstation:

| Software | Version
| --- | --- |
| go | go1.7.x
| glide | 0.12.x

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
$ ./scripts/nozzle.sh
```

#### CI

https://concourse.cfplatformeng.com/teams/splunk/pipelines/splunk-firehose-tile-build

### Exploring Events

Here are a few Splunk queries to explore some the events for making a  dashboard

```
event_type=ValueMetric
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```

```
event_type=CounterEvent
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```
