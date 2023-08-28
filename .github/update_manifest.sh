#!/usr/bin/env bash
set -e
#Set below params in github env variable settings
# API_ENDPOINT, API_USER, API_PASSWORD, SPLUNK_TOKEN, SPLUNK_HOST, SPLUNK_INDEX, SPLUNK_METRIC_INDEX
#Update manifest for deployment
sed -i 's@API_ENDPOINT:.*@'"API_ENDPOINT: $API_ENDPOINT"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@API_USER:.*@'"API_USER: $API_USER"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@API_PASSWORD:.*@'"API_PASSWORD: $API_PASSWORD"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@CLIENT_ID:.*@'"CLIENT_ID: $CLIENT_ID"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@CLIENT_SECRET:.*@'"CLIENT_SECRET: $CLIENT_SECRET"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_HOST:.*@'"SPLUNK_HOST: $SPLUNK_HOST"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_TOKEN:.*@'"SPLUNK_TOKEN: $SPLUNK_TOKEN"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_INDEX:.*@'"SPLUNK_INDEX: $SPLUNK_INDEX"'@' scripts/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_METRIC_INDEX:.*@'"SPLUNK_METRIC_INDEX: $SPLUNK_METRIC_INDEX"'@' scripts/ci_nozzle_manifest.yml
#copy nozzle binary from shared workspace
#cp /tmp/splunk-firehose-nozzle .
