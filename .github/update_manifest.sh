#!/usr/bin/env bash
set -e
#Set below params in github env variable settings
# API_ENDPOINT, API_USER, API_PASSWORD, SPLUNK_TOKEN, SPLUNK_HOST, SPLUNK_INDEX
#Update manifest for deployment
sed -i 's@API_ENDPOINT:.*@'"API_ENDPOINT: $API_ENDPOINT"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@API_USER:.*@'"API_USER: $API_USER"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@API_PASSWORD:.*@'"API_PASSWORD: $API_PASSWORD"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@CLIENT_ID:.*@'"CLIENT_ID: $CLIENT_ID"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@CLIENT_SECRET:.*@'"CLIENT_SECRET: $CLIENT_SECRET"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_HOST:.*@'"SPLUNK_HOST: $SPLUNK_HOST"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_TOKEN:.*@'"SPLUNK_TOKEN: $SPLUNK_TOKEN"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_INDEX:.*@'"SPLUNK_INDEX: $SPLUNK_INDEX"'@' .github/workflows/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_METRIC_INDEX:.*@'"SPLUNK_METRIC_INDEX: $SPLUNK_METRIC_INDEX"'@' .github/workflows/ci_nozzle_manifest.yml
#copy nozzle binary from shared workspace
#cp /tmp/splunk-firehose-nozzle .
