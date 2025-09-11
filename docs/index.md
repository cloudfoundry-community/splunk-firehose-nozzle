# General information

Cloud Foundry Firehose-to-Splunk Nozzle

!!!warning
    Starting from version `1.4.0`, the `go-cfclient` used by the nozzle was upgraded from v2 to v3. As a result of this breaking change, the `environment_variables` for each App object have been replaced with `Cf Labels`.

## Compatibility

### VMware Tanzu Application Service versions

Splunk Firehose Nozzle has been tested on `v3.0`, `v4.0`, `v5.0` and `v6.0` of Tanzu Application Service.

### VMware Ops Manager version

Splunk Firehose Nozzle has been tested on `v3.0.9 LTS` of VMware Ops Manager

## Usage

Splunk nozzle is used to stream Cloud Foundry Firehose events to Splunk HTTP Event Collector. Using pre-defined Splunk sourcetypes, 
the nozzle automatically parses the events and enriches them with additional metadata before forwarding to Splunk. 
For detailed descriptions of each Firehose event type and their fields, refer to underlying [dropsonde protocol](https://github.com/cloudfoundry/dropsonde-protocol). 
Below is a mapping of each Firehose event type to its corresponding Splunk sourcetype. 
Refer to [Searching Events](./troubleshooting#searching-events) for example Splunk searches.

| Firehose event type | Splunk sourcetype    | Description                                                              |
|---------------------|----------------------|--------------------------------------------------------------------------|
| Error               | `cf:error`           | An Error event represents an error in the originating process            |
| HttpStartStop       | `cf:httpstartstop`   | An HttpStartStop event represents the whole lifecycle of an HTTP request |
| LogMessage          | `cf:logmessage`      | A LogMessage contains a "log line" and associated metadata               |
| ContainerMetric     | `cf:containermetric` | A ContainerMetric records resource usage of an app in a container        |
| CounterEvent        | `cf:counterevent`    | A CounterEvent represents the increment of a counter                     |
| ValueMetric         | `cf:valuemetric`     | A ValueMetric indicates the value of a metric at an instant in time      |

In addition, logs from the nozzle itself are of sourcetype `cf:splunknozzle`.


