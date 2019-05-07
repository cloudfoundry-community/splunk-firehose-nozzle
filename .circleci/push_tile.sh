#!/usr/bin/env bash
set -e
sudo apt-get install -y python-pip libpython-dev > /dev/null 2>&1
echo "Installing aws cli..."
sudo pip install awscli > /dev/null 2>&1
echo "Push Splunk tile to s3..."
aws s3 cp tile/product/splunk-nozzle-*.pivotal s3://pcf-ci-artifacts/