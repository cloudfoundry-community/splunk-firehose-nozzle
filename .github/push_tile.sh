#!/usr/bin/env bash
set -e
sudo apt-get install -y python3-pip libpython2-dev > /dev/null 2>&1
#import libpython-dev 
echo "Installing aws cli..."
sudo pip install awscli > /dev/null 2>&1
echo "Push Splunk tile to s3..."
#aws s3 cp tile/product/splunk-nozzle-*.pivotal s3://pcf-ci-artifacts/


