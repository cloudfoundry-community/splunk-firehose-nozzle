[![CircleCI](https://circleci.com/gh/git-lfs/git-lfs.svg?style=shield&circle-token=856152c2b02bfd236f54d21e1f581f3e4ebf47ad)](https://circleci.com/gh/cloudfoundry-community/splunk-firehose-nozzle)
## Splunk Nozzle

Cloud Foundry Firehose-to-Splunk Nozzle

### Usage
Splunk nozzle is used to stream Cloud Foundry Firehose events to Splunk HTTP Event Collector. Using pre-defined Splunk sourcetypes, the nozzle automatically parses the events and enriches them with additional metadata before forwarding to Splunk. For detailed descriptions of each Firehose event type and their fields, refer to underlying [dropsonde protocol](https://github.com/cloudfoundry/dropsonde-protocol). Below is a mapping of each Firehose event type to its corresponding Splunk sourcetype. Refer to [Searching Events](#searching-events) for example Splunk searches.

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
`cloud_controller.admin_read_only` (the latter is only required if `ADD_APP_INFO` is true). If `cloud_controller.admin_read_only` is not
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

- - - -
#### Environment Parameters
You can declare parameters by making a copy of the scripts/nozzle.sh.template.
* `DEBUG`: Enable debug mode (forward to standard out instead of Splunk).

__Cloud Foundry configuration parameters:__
* `API_ENDPOINT`: Cloud Foundry API endpoint address.
* `API_USER`: Cloud Foundry user name. (Must have scope described above)
* `API_PASSWORD`: Cloud Foundry user password.

__Splunk configuration parameters:__
* `SPLUNK_TOKEN`: [Splunk HTTP event collector token](http://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector/).
* `SPLUNK_HOST`: Splunk HTTP event collector host. example: https://example.cloud.splunk.com:8088
* `SPLUNK_INDEX`: The Splunk index events will be sent to. Warning: Setting an invalid index will cause events to be lost. This index must match one of the selected indexes for the Splunk HTTP event collector token used for the SPLUNK_TOKEN parameter.

__Advanced Configuration Features:__
* `JOB_NAME`: Tags nozzle log events with job name.
* `JOB_INDEX`: Tags nozzle log events with job index.
* `JOB_HOST`: Tags nozzle log events with job host.
* `SKIP_SSL_VALIDATION_CF`: Skips SSL certificate validation for connection to Cloud Foundry. Secure communications will not check SSL certificates against a trusted certificate authority.
This is recommended for dev environments only.
* `SKIP_SSL_VALIDATION_SPLUNK`: Skips SSL certificate validation for connection to Splunk. Secure communications will not check SSL certificates against a trusted certificate authority.
This is recommended for dev environments only.
* `FIREHOSE_SUBSCRIPTION_ID`: Tags nozzle events with a Firehose subscription id. See https://docs.pivotal.io/pivotalcf/1-11/loggregator/log-ops-guide.html.
* `FIREHOSE_KEEP_ALIVE`: Keep alive duration for the Firehose consumer.
* `ADD_APP_INFO`: Enriches raw data with app details.
* `IGNORE_MISSING_APP`: If the application is missing, then stop repeatedly querying application info from Cloud Foundry.
* `MISSING_APP_CACHE_INVALIDATE_TTL`:  How frequently the missing app info cache invalidates.
* `APP_CACHE_INVALIDATE_TTL`: How frequently the app info local cache invalidates.
* `ORG_SPACE_CACHE_INVALIDATE_TTL`: How frequently the org and space cache invalidates.
* `APP_LIMITS`: Restrict to APP_LIMITS the most updated apps per request when populating the app metadata cache.
* `BOLTDB_PATH`: Bolt database path.
* `EVENTS`: A comma separated list of events to include. Possible values: ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
* `EXTRA_FIELDS`: Extra fields to annotate your events with (format is key:value,key:value).
* `FLUSH_INTERVAL`: Time interval for flushing queue to Splunk regardless of CONSUMER_QUEUE_SIZE. Protects against stale events in low throughput systems.
* `CONSUMER_QUEUE_SIZE`: Sets the internal consumer queue buffer size. Events will be pushed to Splunk after queue is full.
* `HEC_BATCH_SIZE`: Set the batch size for the events to push to HEC (Splunk HTTP Event Collector).
* `HEC_RETRIES`: Retry count for sending events to Splunk. After expiring, events will begin dropping causing data loss.
* `HEC_WORKERS`: Set the amount of Splunk HEC workers to increase concurrency while ingesting in Splunk.
* `ENABLE_EVENT_TRACING`: Enables event trace logging. Splunk events will now contain a UUID, Splunk Nozzle Event Counts, and a Subscription-ID for Splunk correlation searches.
* `SPLUNK_VERSION`: The Splunk version that determines how HEC ingests metadata fields. Only required for Splunk version 6.3 or below.
    ###  Please note 
    > SPLUNK_VERSION configuration parameter is only required for Splunk version 6.3 and below. 
    For Splunk version 6.3 or below, please deploy nozzle via CLI. Update nozzle_manifest.yml with splunk_version (eg:- SPLUNK_VERSION: 6.3) as an env variable and [deploy nozzle as an app via CLI](#push-as-an-app-to-cloud-foundry).
    
    **[Tile](https://network.pivotal.io/products/splunk-nozzle/)** only supports deployment for Splunk version 6.4 or above
    
- - - -

### Push as an App to Cloud Foundry

Push Splunk Firehose Nozzle as an application to Cloud Foundry. Please refer to **Setup** section for details
on user authentication. 

1. Download the latest release

    ```shell
    git clone https://github.com/cloudfoundry-community/splunk-firehose-nozzle.git
    cd splunk-firehose-nozzle
    ```

1. Authenticate to Cloud Foundry

    ```shell
    cf login -a https://api.[your cf system domain] -u [your id]
    ```

1. Copy the manifest template and fill in needed values (using the credentials created during setup)

    ```shell
    vim ci/nozzle_manifest.yml
    ```

1. Push the nozzle

    ```shell
    make deploy-nozzle
    ```

#### Dump application info to boltdb ####
If in production there are lots of PCF applications(say tens of thousands) and if the user would like to enrich
application logs by including application meta data,querying all application metadata information from PCF may take some time.
For example if we include, add app name, space ID, space name, org ID and org name to the events.
If there are multiple instances of Spunk nozzle deployed the situation will be even worse, since each of the Splunk nozzle(s) will query all applications meta data and
cache the meta data information to the local boltdb file. These queries will introduce load to the PCF system and could potentially take a long time to finish.
Users can run this tool to generate a copy of all application meta data and copy this to each Splunk nozzle deployment. Each Splunk nozzle can pick up the cache copy and update the cache file incrementally afterwards.

Example of how to run the dump application info tool:

```
$ cd tools/dump_app_info
$ go build dump_app_info.go
$ ./dump_app_info --skip-ssl-validation --api-endpoint=https://<your api endpoint> --user=<api endpoint login username> --password=<api endpoint login password>
```

After populating the application info cache file, user can copy to different Splunk nozzle deployments and start Splunk nozzle to pick up this cache file by
specifying correct "--boltdb-path" flag or "BOLTDB_PATH" environment variable.

### Index routing
Index routing is a feature that can be used to send different Cloud Foundry logs to different indexes for better ACL and data retention control in Splunk.

#### Per application index routing via application manifest
In your app manifest provide an environment variable called `SPLUNK_INDEX` and assign it the index you would like to send the app data to

```
applications:
- name: console
  memory: 256M
  disk_quota: 256M
  host: console
  timeout: 180
  buildpack: https://github.com/SUSE/stratos-buildpack
  health-check-type: port
  services:
  - splunk-index
  env:
    SPLUNK_INDEX: testing_index
```

#### Index routing via Splunk configuration
Logs can be routed using fields such as app ID/name, space ID/name or org ID/name.
Users can configure the Splunk configuration files props.conf and transforms.conf on Splunk indexers or Splunk Heavy Forwarders if deployed.

Below are few sample configuration: 

<span>1. </span> Route data from application ID `95930b4e-c16c-478e-8ded-5c6e9c5981f8` to a Splunk `prod` index:

*$SPLUNK_HOME/etc/system/local/props.conf*
```
[cf:logmessage]
TRANSFORMS-index_routing = route_data_to_index_by_field_cf_app_id
```


*$SPLUNK_HOME/etc/system/local/transforms.conf*
```
[route_data_to_index_by_field_cf_app_id]
REGEX = "(\w+)":"95930b4e-c16c-478e-8ded-5c6e9c5981f8"
DEST_KEY = _MetaData:Index
FORMAT = prod
```


<span>2.</span> Routing application logs from any Cloud Foundry orgs whose names are prefixed with `sales` to a Splunk `sales` index.

*$SPLUNK_HOME/etc/system/local/props.conf*
```
[cf:logmessage]
TRANSFORMS-index_routing = route_data_to_index_by_field_cf_org_name

```

*$SPLUNK_HOME/etc/system/local/transforms.conf*
```
[route_data_to_index_by_field_cf_org_name]
REGEX = "cf_org_name":"(sales.*)"
DEST_KEY = _MetaData:Index
FORMAT = sales
```

<span>3.</span> Routing data from sourcetype `cf:splunknozzle` to index `new_index`:
 
*$SPLUNK_HOME/etc/system/local/props.conf*
```
[cf:splunknozzle]
TRANSFORMS-route_to_new_index = route_to_new_index
```

*$SPLUNK_HOME/etc/system/local/transforms.conf*
```
[route_to_new_index]
SOURCE_KEY = MetaData:Sourcetype
DEST_KEY =_MetaData:Index
REGEX = (sourcetype::cf:splunknozzle)
FORMAT = new_index
```

## <a id='walkthrough'></a> Troubleshooting
This topic describes how to troubleshoot Splunk Firehose Nozzle for Cloud Foundry.

#### 1. I can't find my data!

  Are you searching for events and not finding them or looking at a dashboard and seeing "No result found"? Check Splunk Nozzle app logs.

  To view the nozzle's logs running on PCF do the following:

<ol>
  <li>Log in as an admin via the CLI.</li>
  <li>Target the org created by the tile.<br/>
  <pre class="terminal">cf target -o SPLUNK-NOZZLE-ORG</pre>
  </li>
  <li>View the recent app Splunk Nozzle logs (the version number installed by the tile will vary).<br/>
  <pre class="terminal">cf logs --recent splunk-firehoze-nozzle</pre>
  </li>
  <li>Alternatively, you can stream the app logs as they're emitted.<br/>
  <pre class="terminal">cf logs splunk-firehose-nozzle</pre>
  </li>
</ol>


### Here are a few common errors and possible resolutions:

#### Splunk configuration related errors:

<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Post http://localhost:8088/services/collector: read tcp 10.0.0.0:62931-\u003elocalhost:8088: read: connection reset by peer"}}</pre>

This error usually occurs when SSL is enabled on the Splunk HEC endpoint. Confirm that you're using https' in the Splunk HEC URL.


<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Non-ok response code [400] from splunk: {\"text\":\"Incorrect index\",\"code\":7,\"invalid-event-number\":1}"}}</pre>

This usually means the index value specified in the configuration doesn't exist on Splunk Host. Confirm that you're using the correct Splunk index value.

  
<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Non-ok response code [403] from splunk: {\"text\":\"Invalid token\",\"code\":4}"}}</pre>

This can occur when the Splunk HEC Token value is invalid. Confirm that you're using a valid token.

<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Post https://localhost:8088/services/collector: x509: cannot validate certificate for localhost because it doesn't contain any IP SANs"}}</pre>

This usually means that there was no valid SSL certificate found. Confirm that you're using a valid SSL certificate for the Splunk server, or set 'Skip SSL Validation' to `true` under Splunk settings.

<p class="note"><strong>Note:</strong>Disabling SSL validation is not recommended for production environments.</p>


<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Post https://localhost:8088/services/collector: dial tcp localhost:8088: getsockopt: connection refused"}}</pre>

This error can occur when the Splunk server is offline or when the Splunk HEC URL is not valid. Confirm that both the Splunk server is running and that you're using a valid URL. 

#### Cloud Foundry configuration related errors:

<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Failed to run splunk-firehose-nozzle","log_level":2,"data":{"error":"Error getting token: oauth2: cannot fetch token: 401 Unauthorized\nResponse: {\"error\":\"unauthorized\",\"error_description\":\"Bad credentials\"}"}}</pre>

This error can occur when the credentials provided for CF environment are invalid. Confirm that the API User and API Password each have access to the CF environment. 


<pre class="terminal">{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Failed to run splunk-firehose-nozzle","log_level":2,"data":{"error":"Could not get api /v2/info: Get https://api.cfendpoint.com/v2/info: x509: certificate signed by unknown authority"}}</pre>

This means that no valid SSL certificate was found. To remediate this error, provide a valid SSL certificate for Cloud Foundry or set 'Skip SSL Validation' to true under Cloud Foundry Settings.

<p class="note"><strong>Note:</strong>Disabling SSL validation is not recommended for production environments.</p>

The following troubleshooting tips assume you have access to Splunk to run basic searches against index `_internal` and the user-specified index for Firehose events.

### 2. Ensure Splunk Nozzle is forwarding events from the Firehose:

Search app logs of the Nozzle to confirm correct behavior:

<pre class="terminal">
sourcetype="cf:splunknozzle"
</pre>

A correct setup logs a start message with configuration parameters of the Nozzle logged as a JSON object, for example:

<pre class="terminal">
  data:	{
     add-app-info: true
     api-endpoint: https://api.endpoint.com
     app-cache-ttl: 0
     app-limits: 0
     batch-size: 1000
     boltdb-path: cache.db
     branch: null
     buildos: null
     commit: null
     debug:	 false
     extra-fields:
     flush-interval: 5000000000
     hec-workers: 8
     ignore-missing-apps: true
     job-host:
     job-index: -1
     job-name: splunk-nozzle
     keep-alive: 25000000000
     missing-app-cache-ttl:	 0
     queue-size: 10000
     retries: 2
     skip-ssl: true
     splunk-host: http://localhost:8088
     splunk-index: atomic
     splunk-version: 6.6
     subscription-id: splunk-firehose
     trace-logging: true
     version:
     wanted-events: ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
  }
  ip: 10.0.0.0
  log_level: 1
  logger_source: splunk-nozzle-logger
  message: splunk-nozzle-logger.Running splunk-firehose-nozzle with following configuration variables
  origin: splunk_nozzle
</pre>

  Search app logs of the Nozzle for any errors:
<pre class="terminal">
sourcetype="cf:splunknozzle" data.error=*
</pre>

Errors are logged with corresponding message and stacktrace.

### 3. Check for dropped events due to HTTP Event Collector availability:

As the Splunk Firehose Nozzle sends data to Splunk via HTTPS using the HTTP Event Collector, it is also susceptible to any network issues across the network path from point to point. Run the following search to determine if Splunk has indexed any events indicating issues with the HEC Endpoint.

<pre class="terminal">
  sourcetype="cf:splunknozzle" "dropping events"
</pre>

### 4. Check for data loss inside the Splunk Firehose Nozzle:

If "Event Tracing" is enabled, extra metadata will be attached to events. This allows searches to calculate the percentage of data loss inside the Splunk Firehose Nozzle, if applicable.

Each instance of the Splunk Firehose Nozzle will run with a randomly generated UUID. The query below will display the message success rate for each UUID (Please update the index value based on your nozzle configuration).

<pre class="terminal">
index=main | stats count as total_events, min(nozzle-event-counter) as min_number, max(nozzle-event-counter) as max_number by uuid | eval event_number =  max_number - min_number | eval success_percentage = total_events/event_number*100 | stats max(success_percentage) by uuid
</pre>

#### Searching Events

Here are two short Splunk queries to start exploring some of the Cloud Foundry events in Splunk.

```
sourcetype="cf:valuemetric"
    | stats avg(value) by job_instance, name
```

```
sourcetype="cf:counterevent"
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```
### Development

#### Software Requirements

Make sure you have the following installed on your workstation:

| Software | Version
| --- | --- |
| go | go1.8.x
| glide | 0.12.x

Then install all dependent packages via [Glide](https://glide.sh/):

```
$ cd <REPO_ROOT_DIRECTORY>
$ make installdeps
```

#### Environment

For development against [bosh-lite](https://github.com/cloudfoundry/bosh-lite),
copy `tools/nozzle.sh.template` to `tools/nozzle.sh` and supply missing values:

```
$ cp script/dev.sh.template tools/nozzle.sh
$ chmod +x tools/nozzle.sh
```

Build project:

```
$ make VERSION=1.1
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

# Maintenance And Support

Splunk Firehose Nozzle project is supported through Splunk Support assuming the customer has a current Splunk support entitlement.  For customers that do not have a current Splunk support entitlement, please file an issue at [create a new issue](https://github.com/cloud-foundry/splunk-firehose-nozzle/issues/new)

