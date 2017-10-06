#!/usr/bin/env bash

export SKIP_SSL_VALIDATION_CF=true
export SKIP_CF_SSL_VALIDATION_SPLUNK=false

export ADD_APP_INFO=true
export API_ENDPOINT=https://api.sys.pie-21.cfplatformeng.com
export API_USER=admin
export API_PASSWORD=mjnpuuAIyyvc450_34Kon_RMfBP7jIdj

# possible values: ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
#export EVENTS=ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric

export SPLUNK_TOKEN=AA10CEA0-F806-4CEB-A1CD-154609F852BC
export SPLUNK_HOST=https://localhost:8088
export SPLUNK_INDEX=atomic

go run main.go --enable-event-tracing --hec-retries=2 --events=ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric