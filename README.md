## Splunk Nozzle

Cloud Foundry Firehose-to-Splunk Nozzle

### Setup

The Nozzle requires a user with the scope `doppler.firehose` and 
`cloud_controller.admin_read_only` (the latter is only required if `ADD_APP_INFO` is true). 
You can either
* Add the user manually using [uaac](https://github.com/cloudfoundry/cf-uaac)
* Add a new user to the deployment manifest; see [uaa.scim.users](https://github.com/cloudfoundry/uaa-release/blob/master/jobs/uaa/spec)

Manifest example:
```yaml
uaa:
  scim:
    users:
      - splunk-firehose|password123|cloud_controller.admin_read_only,doppler.firehose
```

`uaac` example:
```shell
uaac target https://uaa.[system domain url]
uaac token client get admin -s [admin client credentials secret]
uaac -t user add splunk-nozzle --password password123 --emails na
uaac -t member add cloud_controller.admin_read_only splunk-nozzle
uaac -t member add doppler.firehose splunk-nozzle
```

`cloud_controller.admin_read_only` will work for cf v241 
or later. Earlier versions should use `cloud_controller.admin` instead.

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

Add a new dependency:
```
glide get github.com/kelseyhightower/envconfig --strip-vendor
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

### Push as an App to Cloud Foundry

[splunk-firehose-nozzle-release](https://github.com/cloudfoundry-community/splunk-firehose-nozzle-release)
packages this code into a
[BOSH](https://bosh.io) release for deployment. The code could also be run on
Cloud Foundry as an application. See the **Setup** section for details
on making a user and credentials.

1. Download the latest release
    
    ```shell
    git clone https://github.com/cloudfoundry-community/splunk-firehose-nozzle.git
    cd splunk-firehose-nozzle
    ```
    
1. Authenticate to Cloud Foundruy
    
    ```shell
    cf login -a https://api.[your cf system domain] -u [your id]
    ```

1. Copy the manifest template and fill in needed values (using the credentials created during setup)

    ```shell
    cp manifest.yml.template manifest.yml 
    vim manifest.yml
    ```

1. Push the nozzle

    ```shell
    cf push
    ```


#### Exploring Events

Here are two short Splunk queries to start exploring some of the events

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
