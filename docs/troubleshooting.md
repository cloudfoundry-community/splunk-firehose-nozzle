# Troubleshooting

This topic describes how to troubleshoot Splunk Firehose Nozzle for Cloud Foundry.

## No data in Splunk

Are you searching for events and not finding them or looking at a dashboard and seeing "No result found"? Check Splunk Nozzle app logs.

To view the nozzle's logs running on CF do the following:

1. Log in as an admin via the CLI.
2. Target the org created by the tile.
   ```
   cf target -o SPLUNK-NOZZLE-ORG
   ```
3. View the recent app Splunk Nozzle logs (the version number installed by the tile will vary).
   ```
   cf logs --recent splunk-firehoze-nozzle
   ```
4. Alternatively, you can stream the app logs as they are emitted.
   ```
   cf logs splunk-firehose-nozzle
   ```

### Here are a few common errors and possible resolutions:

#### Splunk configuration related errors:

```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Post http://localhost:8088/services/collector: read tcp 10.0.0.0:62931-\u003elocalhost:8088: read: connection reset by peer"}}
```

This error usually occurs when SSL is enabled on the Splunk HEC endpoint. Confirm that you're using https' in the Splunk HEC URL.


```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Non-ok response code [400] from splunk: {\"text\":\"Incorrect index\",\"code\":7,\"invalid-event-number\":1}"}}
```
This usually means the index value specified in the configuration doesn't exist on Splunk Host. Confirm that you're using the correct Splunk index value.

```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Non-ok response code [403] from splunk: {\"text\":\"Invalid token\",\"code\":4}"}}
```

This can occur when the Splunk HEC Token value is invalid. Confirm that you're using a valid token.

```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Post https://localhost:8088/services/collector: x509: cannot validate certificate for localhost because it doesn't contain any IP SANs"}}
```

This usually means that there was no valid SSL certificate found. Confirm that you're using a valid SSL certificate for the Splunk server, 
or set 'Skip SSL Validation' to `true` under Splunk settings.

**Note:** Disabling SSL validation is not recommended for production environments.

```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Unable to talk to Splunk","log_level":2,"data":{"error":"Post https://localhost:8088/services/collector: dial tcp localhost:8088: getsockopt: connection refused"}}
```
This error can occur when the Splunk server is offline or when the Splunk HEC URL is not valid. Confirm that both the Splunk server is running and that you're using a valid URL.

#### Cloud Foundry configuration related errors:

```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Failed to run splunk-firehose-nozzle","log_level":2,"data":{"error":"Error getting token: oauth2: cannot fetch token: 401 Unauthorized\nResponse: {\"error\":\"unauthorized\",\"error_description\":\"Bad credentials\"}"}}
```
This error can occur when the credentials provided for CF environment are invalid. Confirm that the API User and API Password each have access to the CF environment.

```
{"timestamp":"<time>","source":"splunk-nozzle-logger","message":"splunk-nozzle-logger.Failed to run splunk-firehose-nozzle","log_level":2,"data":{"error":"Could not get api /v2/info: Get https://api.cfendpoint.com/v2/info: x509: certificate signed by unknown authority"}}
```

This means that no valid SSL certificate was found. To remediate this error, provide a valid SSL certificate for Cloud Foundry or set 
'Skip SSL Validation' to true under Cloud Foundry Settings.

**Note:** Disabling SSL validation is not recommended for production environments.</p>

The following troubleshooting tips assume you have access to Splunk to run basic searches against index `_internal` and the user-specified index for Firehose events.

## Ensure Splunk Nozzle is forwarding events from the Firehose:

Search app logs of the Nozzle to confirm correct behavior:

```
sourcetype="cf:splunknozzle"
```

A correct setup logs a start message with configuration parameters of the Nozzle logged as a JSON object, for example:

``` 
data:	
  {
     add-app-info: AppName,OrgName,OrgGuid,SpaceName,SpaceGuid
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
     firehose-subscription-id: splunk-firehose
     trace-logging: true
     status-monitor-interval: 0s
     version:
     wanted-events: ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
  }
  ip: 10.0.0.0
  log_level: 1
  logger_source: splunk-nozzle-logger
  message: splunk-nozzle-logger.Running splunk-firehose-nozzle with following configuration variables
  origin: splunk_nozzle
```


Search app logs of the Nozzle for any errors:
```
sourcetype="cf:splunknozzle" data.error=*
```

Errors are logged with corresponding message and stacktrace.

## Check for dropped events due to HTTP Event Collector availability:

As the Splunk Firehose Nozzle sends data to Splunk via HTTPS using the HTTP Event Collector, it is also susceptible to any network issues across the network path from point to point.
Run the following search to determine if Splunk has indexed any events indicating issues with the HEC Endpoint.

```
  sourcetype="cf:splunknozzle" "dropping events"
```

## Check for dropped events due to slow downstream(Network/Splunk):

If the nozzle emits the ‘dropped events’ warning saying that downstream is slow, then the network or Splunk environment might needs to be scaled.
(eg. Splunk HEC receiver node, Splunk Indexer, LB etc)

Run the following search to determine if Splunk has indexed any events indicating such issues.

```
  sourcetype="cf:splunknozzle" "dropped Total of"
```

## Check for data loss inside the Splunk Firehose Nozzle:

If "Event Tracing" is enabled, extra metadata will be attached to events. This allows searches to calculate the percentage of data loss inside the Splunk Firehose Nozzle, if applicable.

Each instance of the Splunk Firehose Nozzle will run with a randomly generated UUID. The query below will display the message 
success rate for each UUID (Please update the index value based on your nozzle configuration).

```
index=main | stats count as total_events, min(nozzle-event-counter) as min_number, max(nozzle-event-counter) as max_number by uuid | eval event_number =  max_number - min_number | eval success_percentage = total_events/event_number*100 | stats max(success_percentage) by uuid
```

## Authentication is not working even if correct CF Client ID/secret is configured: (applicable in v1.2.3)

Due to a known issue in an indirect dependency (an OAuth library), if the client secret has any special characters (eg. *!#$&@^) 
then it will not work. For now, user has to configure a client secret without any of this characters. 
Once the library in question is updated in the next release it will work even with the special characters.

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

## Nozzle is not collecting any data with 'websocket' (bad handshake) error

If the nozzle reports below error, then check if the configured "firehose-subscription-id" has '#' as a prefix. 
Please remove the prefix or prepend any other character than '#' to fix this issue.
```
Error dialing trafficcontroller server: websocket: bad handshake.\nPlease ask your Cloud Foundry Operator to check the platform configuration (trafficcontroller is wss://****:443).
```