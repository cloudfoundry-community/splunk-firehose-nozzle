#!/usr/bin/env bash
set -e
#update deps
make
#Install CF CLI
wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | sudo apt-key add -
echo "deb https://packages.cloudfoundry.org/debian stable main" | sudo tee /etc/apt/sources.list.d/cloudfoundry-cli.list
#Add support for https apt sources
sudo apt-get update
sudo apt-get install apt-transport-https ca-certificates
sudo apt-get install cf-cli
#CF Login
API_PASSWORD_DECRYPTED=$(echo "$API_PASSWORD" | openssl aes-256-cbc -d -pbkdf2 -a -pass pass:"$ENCRYPT_KEY")
cf login --skip-ssl-validation -a "$API_ENDPOINT" -u "$API_USER" -p "$API_PASSWORD_DECRYPTED"

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

gem install cf-uaac
uaac target "$API_UAA_ENDPOINT" --skip-ssl-validation
API_CLIENT_PASSWORD_DECRYPTED=$(echo "$API_CLIENT_PASSWORD" | openssl aes-256-cbc -d -pbkdf2 -a -pass pass:"$ENCRYPT_KEY")
uaac token client get "$API_USER" -s "$API_CLIENT_PASSWORD_DECRYPTED"

if [ $(uaac client get "$CLIENT_ID" | grep -woc "$CLIENT_ID") -eq 0 ]; then
  uaac client add "$CLIENT_ID" --name splunk-firehose --secret "$CLIENT_SECRET" --authorized_grant_types client_credentials,refresh_token --authorities doppler.firehose,cloud_controller.admin_read_only
fi