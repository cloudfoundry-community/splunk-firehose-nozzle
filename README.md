## Splunk Nozzle

Cloud Foundry Firehose-to-Splunk Nozzle

### Usage
Splunk nozzle is used to stream Cloud Foundry Firehose events to Splunk HTTP Event Collector. Using pre-defined Splunk sourcetypes, the nozzle automatically parses the events and enriches them with additional metadata before forwarding to Splunk. For detailed descriptions of each Firehose event type and their fields, refer to underlying [dropsonde protocol](https://github.com/cloudfoundry/dropsonde-protocol). Below is a mapping of each firehose event type to its corresponding Splunk sourcetype. Refer to [Searching Events](#searching-events) for example Splunk searches.

| Firehose event type | Splunk sourcetype | Description
|---|---|---
| Error | `cf:error` | An Error event represents an error in the originating process
| HttpStartStop | `cf:httpstartstop` | An HttpStartStop event represents the whole lifecycle of an HTTP request
| LogMessage | `cf:logmessage` | A LogMessage contains a "log line" and associated metadata
| ContainerMetric | `cf:containermetric` | A ContainerMetric records resource usage of an app in a container
| CounterEvent | `cf:counterevent` | A CounterEvent represents the increment of a counter
| ValueMetric | `cf:valuemetric` | A ValueMetric indicates the value of a metric at an instant in time

In addition, logs from the nozzle itself are of sourcetype `cf:splunknozzle`.

### Setup

The Nozzle requires a user with the scope `doppler.firehose` and
`cloud_controller.admin_read_only` (the latter is only required if `ADD_APP_INFO` is true). If `cloud_controller.admin` is not
available in the system, switch to use `cloud_controller.admin`.

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


#### Environment Paramaters (declare parameters by making a copy of scripts/nozzle.sh.template)

Cloud Foundry configuration parameters:

DEBUG -
Enable debug mode (forward to standard out instead of Splunk).

SKIP_SSL_VALIDATION -
Skip cert validation (for dev environments).

JOB_NAME -
Job name to tag nozzle's own log events with.

JOB_INDEX -
Job index to tag nozzle's own log events with.

JOB_HOST -
Job host to tag nozzle's own log events with.

ADD_APP_INFO -
Query API to fetch app details.

API_ENDPOINT -
Cloud Foundry API endpoint address.

API_USER -
Cloud Foundry user name. (Must have scope described above)

API_PASSWORD -
Cloud Foundry user password.

BOLTDB_PATH -
Bolt Database path.

EVENTS -
Comma separated list of events to include.
possible values: ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric

EXTRA_FIELDS -
Extra fields you want to annotate your events with (format is key:value,key:value).

FIREHOSE_KEEP_ALIVE -
Keep Alive duration for the firehose consumer.

FIREHOSE_SUBSCRIPTION_ID -
Id for the firehose subscription.

Splunk configuration parameters:

SPLUNK_TOKEN -
[Splunk HTTP event collector token](http://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector/)

SPLUNK_HOST -
Splunk HTTP event collector host.
example: https://example.cloud.splunk.com:8088

SPLUNK_INDEX -
The Splunk index events will be sent to.
Warning: Setting an invalid index will cause events to be lost.

FLUSH_INTERVAL -
Set the interval for flushing to the heavy forwarder.

CONSUMER_QUEUE_SIZE -
Set the internal consumer queue buffer size.

HEC_BATCH_SIZE -
Set the batch size for the events to push to HEC (Splunk HTTP Event Collector).

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
$ make installdeps
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
$ make VERSION=1.0
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

#### Dump app info to boltdb ####
If in production, there are lots of PCF applications say tens of thounsands and if user would like to enrich appliation log by using application meta data,
for example, add app name, space ID, space name, org ID, org name to the events, querying all application metadata information from PCF may take lots of time.
If there are multiple instances of Spunk nozzle deployed, the situation will be even worse since each of Splunk nozzle will query all applications meta data and
cache the meta data information to a local file (boltdb file). These querys will introduce load to PCF system and also probably take a long time to finish.
The scripts/dump_app_info.go is a tool which is used to mitegate this problem. User can run this tool to generate a copy of all application meta data and copy
to each Splunk nozzle deployment. Each Splunk nozzle can pick up the cache copy and update the cache file incrementally afterwards.

To run this tool, user can do

```
$ cd scripts
$ go build dump_app_info.go
$ ./dump_app_info --skip-ssl-validation --api-endpoint=https://<your api endpoint> --user=<api endpoint login username> --password=<api endpoint login password>
```

After populating the application info cache file, user can copy to different Splunk nozzle deployments and start Splunk nozzle to pick up this cache file by
specifying correct "--boltdb-path" flag or "BOLTDB_PATH" environment variable.


#### Troubleshooting
In most cases, you would only need to troubleshoot from Splunk which include not only firehose data but also this nozzle internal logs. However, if the nozzle is still not forwarding any data, a good place to start is to get app internal logs directly:

```shell
cf logs splunk-firehose-nozzle
```

A common mis-configuration occurs when having invalid or unsigned certificate for Cloud Foundry API endpoint. In that case, for non-production environments, you can set `SKIP_SSL_VALIDATION` to `true` in above manifest.yml before re-deploying the app.

#### Searching Events

Here are two short Splunk queries to start exploring some of the events

```
sourcetype="cf:valuemetric"
    | stats avg(value) by job_instance, name
```

```
sourcetype="cf:counterevent"
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```
