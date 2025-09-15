# Initial setup

The Nozzle requires a client with the authorities `doppler.firehose` and `cloud_controller.admin_read_only` (the latter is only required if `ADD_APP_INFO` is enabled) and grant-types `client_credentials` and `refresh_token`. If `cloud_controller.admin_read_only` is not
available in the system, switch to use `cloud_controller.admin`.

You can either
* Add the client manually using [uaac](https://github.com/cloudfoundry/cf-uaac)
* Add the client to the deployment manifest; see [uaa.scim.users](https://github.com/cloudfoundry/uaa-release/blob/master/jobs/uaa/spec)

Manifest example:

```yaml

# Clients
uaa.clients:
    splunk-firehose:
      id: splunk-firehose
      override: true
      secret: splunk-firehose-secret
      authorized-grant-types: client_credentials,refresh_token
      authorities: doppler.firehose,cloud_controller.admin_read_only
```

`uaac` example:
```shell
uaac target https://uaa.[system domain url]
uaac token client get admin -s [admin client credentials secret]
uaac client add splunk-firehose --name splunk-firehose
uaac client add splunk-firehose --secret [your_client_secret]
uaac client add splunk-firehose --authorized_grant_types client_credentials,refresh_token
uaac client add splunk-firehose --authorities doppler.firehose,cloud_controller.admin_read_only

```

`cloud_controller.admin_read_only` will work for cf v241
or later. Earlier versions should use `cloud_controller.admin` instead.

### Push as an App to Cloud Foundry

Push Splunk Firehose Nozzle as an application to Cloud Foundry. Please refer to **Setup** section for details
on user authentication.

1. Download the latest release

    ```shell
    git clone https://github.com/cloudfoundry-community/splunk-firehose-nozzle.git
    cd splunk-firehose-nozzle
    ```

2. Authenticate to Cloud Foundry

    ```shell
    cf login -a https://api.[your cf system domain] -u [your id]
    ```

3. Copy the manifest template and fill in needed values (using the credentials created during setup)

    ```shell
    vim scripts/ci_nozzle_manifest.yml
    ```

4. Push the nozzle

    ```shell
    make deploy-nozzle
    ```

#### Dump application info to boltdb 
If in production where there are lots of CF applications (say tens of thousands) and if the user would like to enrich
application logs by including application metadata, querying all application metadata information from CF may take some time -
for example if we include: add app name, space ID, space name, org ID and org name to the events.
If there are multiple instances of Spunk nozzle deployed the situation will be even worse, since each of the Splunk nozzle(s) will query all applications meta data and
cache the metadata information to the local boltdb file. These queries will introduce load to the CF system and could potentially take a long time to finish.
Users can run this tool to generate a copy of all application metadata and copy this to each Splunk nozzle deployment. Each Splunk nozzle can pick up the cache copy and update the cache file incrementally afterwards.

Example of how to run the dump application info tool:

```
$ cd tools/dump_app_info
$ go build dump_app_info.go
$ ./dump_app_info --skip-ssl-validation --api-endpoint=https://<your api endpoint> --user=<api endpoint login username> --password=<api endpoint login password>
```

After populating the application info cache file, user can copy to different Splunk nozzle deployments and start Splunk nozzle to pick up this cache file by
specifying correct "--boltdb-path" flag or "BOLTDB_PATH" environment variable.

### Disable logging for noisy applications
Set `F2S_DISABLE_LOGGING` = true as a environment variable in applications's manifest to disable logging.


## Index routing
Index routing is a feature that can be used to send different Cloud Foundry logs to different indexes for better ACL and data retention control in Splunk.

### Per application index routing via application manifest
To enable per app index routing,
* Please set environment variable `SPLUNK_INDEX` in your application's manifest ([example below](#example-manifest-file))
* Make sure Splunk nozzle is configured with `ADD_APP_INFO` (Select at least one of AppName,OrgName,OrgGuid,SpaceName,SpaceGuid) to enable app info caching
* Make sure `SPLUNK_INDEX` specified in app's manifest exist in Splunk and can receive data for the configured Splunk HEC token.

> **WARNING**: If `SPLUNK_INDEX` is invalid, events from other apps may also get lost as splunk will drop entire event batch if any of the event from batch is invalid (i.e. invalid index)

There are two ways to set the variable:

In your app manifest provide an environment variable called `SPLUNK_INDEX` and assign it the index you would like to send the app data to.

#### Example Manifest file
```
applications:
- name: <App-Name>
  memory: 256M
  disk_quota: 256M
  ...
  env:
    SPLUNK_INDEX: <SPLUNK_INDEX>
    ...
```

You can also update the env on the fly using cf-cli command:
```
cf set-env <APP_NAME> SPLUNK_INDEX <ENV_VAR_VALUE>
```
#### Please note
> If you are updating env on the fly, make sure that `APP_CACHE_INVALIDATE_TTL` is greater tha 0s. Otherwise cached app-info will not be updated and events will not be sent to required index.


### Index routing via Splunk configuration
Logs can be routed using fields such as app ID/name, space ID/name or org ID/name.
Users can configure the Splunk configuration files props.conf and transforms.conf on Splunk indexers or Splunk Heavy Forwarders if deployed.

Below are few sample configuration:

1. Route data from application ID `95930b4e-c16c-478e-8ded-5c6e9c5981f8` to a Splunk `prod` index:

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


2. Routing application logs from any Cloud Foundry orgs whose names are prefixed with `sales` to a Splunk `sales` index.

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

3. Routing data from sourcetype `cf:splunknozzle` to index `new_index`:

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
**Note:** Moving from version 1.2.4 to 1.2.5, timestamp will use nanosecond precision instead of milliseconds.


## Monitoring (Metric data Ingestion):

| Metric Name                      | Description                                                                 |
|----------------------------------|-----------------------------------------------------------------------------|
| `nozzle.queue.percentage`        | Shows how much internal queue is filled                                     |
| `splunk.events.dropped.count`    | Number of events dropped from splunk HEC                                    |
| `splunk.events.sent.count`       | Number of events sent to splunk                                             |
| `firehose.events.dropped.count`  | Number of events dropped from nozzle                                        |
| `firehose.events.received.count` | Number of events received from firehose(websocket)                          |
| `splunk.events.throughput`       | Average Payload size                                                        |
| `nozzle.usage.ram`               | RAM Usage                                                                   |
| `nozzle.usage.cpu`               | CPU Usage                                                                   |
| `nozzle.cache.memory.hit`        | How many times it has successfully retrieved the data from memory           |
| `nozzle.cache.memory.miss`       | How many times it has unsuccessfully tried to retreive the data from memory |
| `nozzle.cache.remote.hit`        | How many times it has successfully retrieved the data from remote           |
| `nozzle.cache.remote.miss`       | How many times it has unsuccessfully tried to retrieve the data from remote |
| `nozzle.cache.boltdb.hit`        | How many times it has successfully retrieved the data from BoltDB           |
| `nozzle.cache.boltdb.miss`       | How many times it has unsuccessfully tried to retrieve the data from BoltDB |

![event_count](https://user-images.githubusercontent.com/89519924/200804220-1adff84c-e6f1-4438-8d30-6e2cce4984f5.png)

![nozzle_logs](https://user-images.githubusercontent.com/89519924/200804285-22ad7863-1db3-493a-8196-cc589837db76.png)

**Note:** Select value Rate(Avg) for Aggregation from Analysis tab on the top right.

You can find a pre-made dashboard that can be used for monitoring in the `dashboards` directory.

### Routing data through edge processor via HEC
Logs can be routed to Splunk via Edge Processor. Assuming that you have a working Edge Processor instance, you can use it with minimal
changes to nozzle configuration.

Configuration fields that you should change are:
* `SPLUNK_HOST`: Use the host of your Edge Processor instance instead of Splunk. Example: https://x.x.x.x:8088.
* `SPLUNK_TOKEN`: It is a required parameter. A token used to authorize your request, can be found in Edge Processor settings. If your
  EP token authentication is turned off, you can enter a placeholder values instead (e.x. "-").



