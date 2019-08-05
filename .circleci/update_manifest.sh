#!/usr/bin/env bash
set -e
#Set below params in CircleCI env variable settings
# API_ENDPOINT, API_USER, API_PASSWORD, SPLUNK_TOKEN, SPLUNK_HOST, SPLUNK_INDEX
#Update manifest for deployment
sed -i 's@API_ENDPOINT:.*@'"API_ENDPOINT: $API_ENDPOINT"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@API_USER:.*@'"API_USER: $API_USER"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@API_PASSWORD:.*@'"API_PASSWORD: $API_PASSWORD"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@CLIENT_ID:.*@'"CLIENT_ID: $CLIENT_ID"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@CLIENT_SECRET:.*@'"CLIENT_SECRET: $CLIENT_SECRET"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_HOST:.*@'"SPLUNK_HOST: $SPLUNK_HOST"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_TOKEN:.*@'"SPLUNK_TOKEN: $SPLUNK_TOKEN"'@' .circleci/ci_nozzle_manifest.yml
sed -i 's@SPLUNK_INDEX:.*@'"SPLUNK_INDEX: $SPLUNK_INDEX"'@' .circleci/ci_nozzle_manifest.yml
#copy nozzle binary from shared workspace
cp /tmp/splunk-firehose-nozzle .
