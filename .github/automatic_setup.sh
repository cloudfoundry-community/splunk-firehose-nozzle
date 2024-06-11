#!/bin/bash

# Color
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

function red {
    printf "${RED}$@${NC}\n"
}

function green {
    printf "${GREEN}$@${NC}\n"
}

function yellow {
    printf "${YELLOW}$@${NC}\n"
}

wait_for_splunk() {
  while [ "$(sudo docker ps | grep "splunk:latest" | grep healthy)" == "" ] ; do
    echo $(yellow "Waiting for Splunk initialization")
    sleep 1
  done
}

create_splunk_indexes() {
  index_names=$SPLUNK_METRIC_INDEX
  index_types="metric"
  if ! curl -k -u $SPLUNK_USER:$SPLUNK_PASSWORD "https://localhost:8089/services/data/indexes" \
    -d datatype="${index_types}" -d name="${index_names}" ; then
    echo "Error when creating ${index_names} of type ${index_types}"
  fi
}

create_splunk_hec() {
  if ! curl -k -u $SPLUNK_USER:$SPLUNK_PASSWORD https://localhost:8089/servicesNS/admin/splunk_httpinput/data/inputs/http -d name=some_name | grep "token" | cut -c 29-64 > hec_token ; then
    echo "Error when creating Splunk token"
  fi
}

change_min_free_space() {
  DOCKER_ID=$(sudo docker ps | grep 'splunk/splunk:latest' | awk '{ print $1 }')
  sudo docker exec --user splunk "$DOCKER_ID" bash -c "echo -e '\n[diskUsage]\nminFreeSpace = 2000' >> /opt/splunk/etc/system/local/server.conf"
  curl -k -u admin:changeme2 https://localhost:8089/services/server/control/restart -X POST
  sleep 60
}

echo "Show working directory:"
pwd

sudo docker pull splunk/splunk:latest
echo $(green "Running Splunk in Docker")
sudo docker run -d -p 8000:8000 -p 8088:8088 -p 8089:8089 -e SPLUNK_START_ARGS='--accept-license' -e SPLUNK_PASSWORD='changeme2' splunk/splunk:latest

wait_for_splunk

echo $(green "Preparing Splunk instance")

change_min_free_space
create_splunk_indexes
#create_splunk_hec

#echo "$(cat hec_token)"

