#!/usr/bin/env bash
set -e
#update deps
make
#Install CF CLI
wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | sudo apt-key add -
echo "deb https://packages.cloudfoundry.org/debian stable main" | sudo tee /etc/apt/sources.list.d/cloudfoundry-cli.list
#Add support for https apt sources
sudo apt-get install apt-transport-https ca-certificates
sudo apt-get update
sudo apt-get install cf-cli
#CF Login
cf login --skip-ssl-validation -a $API_ENDPOINT -u $API_USER -p $API_PASSWORD -o system -s system
#Create splunk-ci org and space
if [  "`cf o | grep "splunk-ci-org"`" == "splunk-ci-org" ]; then
   echo "splunk-ci-org org already exists"
   cf target -o "splunk-ci-org" -s "splunk-ci-space"
else
   echo "creating splunk-ci-org org and space"
   cf create-org splunk-ci-org
   cf target -o splunk-ci-org
   cf create-space splunk-ci-space
   cf target -o "splunk-ci-org" -s "splunk-ci-space"
fi