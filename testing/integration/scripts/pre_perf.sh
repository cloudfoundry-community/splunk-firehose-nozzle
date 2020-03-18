#!/usr/bin/env bash

cf login --skip-ssl-validation -a $API_ENDPOINT -u $API_USER -p $API_PASSWORD
#Create splunk-ci org and space
if [  "`cf o | grep "splunk-ci-org-perf"`" == "splunk-ci-org-perf" ]; then
   echo "splunk-ci-org-perf org already exists"
   cf target -o "splunk-ci-org-perf" -s "splunk-ci-space-perf"
else
   echo "creating splunk-ci-org-perf org and space"
   cf create-org splunk-ci-org-perf
   cf target -o splunk-ci-org-perf
   cf create-space splunk-ci-space-perf
   cf target -o "splunk-ci-org-perf" -s "splunk-ci-space-perf"
fi
