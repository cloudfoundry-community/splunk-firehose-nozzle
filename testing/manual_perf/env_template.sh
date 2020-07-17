#! /bin/bash
# PCF config
export API_ENDPOINT=
export API_USER=
export API_PASSWORD=
# Nozzle config
export ADD_APP_INFO=True
export EVENTS=ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
export SPLUNK_TOKEN=
export SPLUNK_HOST=
export SPLUNK_INDEX=pcfperf
export FIREHOSE_SUBSCRIPTION_ID=splunk-ci
export CLIENT_ID=splunk-perf
export CLIENT_SECRET=splunk-perf
export ENABLE_EVENT_TRACING=True
export SKIP_SSL_VALIDATION_CF=True
export SKIP_SSL_VALIDATION_SPLUNK=True
export EXTRA_FIELDS=test_tag:02worker100batch
export HEC_WORKERS=2
export HEC_BATCH_SIZE=100
# Data Gen config
export EPS=1000
export TOTAL_EVENTS=200000
export NOZZLE_DURATION=400