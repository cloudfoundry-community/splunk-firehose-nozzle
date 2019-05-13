#!/usr/bin/env bash
set -e
wget https://github.com/cf-platform-eng/tile-generator/releases/download/v13.0.2/pcf_linux-64bit > /dev/null 2>&1
chmod +x pcf_linux-64bit
sudo mv pcf_linux-64bit /usr/local/bin/tile
sudo apt install python-pip > /dev/null 2>&1
echo "installing vurtualenv"
sudo /usr/bin/easy_install virtualenv > /dev/null 2>&1
virtualenv -p python tile-generator-env
source tile-generator-env/bin/activate
echo "Installing tile-generator..."
pip install tile-generator
cd tile
echo "Installing bosh..."
wget https://github.com/cloudfoundry/bosh-cli/releases/download/v5.5.0/bosh-cli-5.5.0-linux-amd64 > /dev/null 2>&1
mv bosh-cli-5.5.0-linux-amd64 bosh
chmod +x ./bosh
sudo mv ./bosh /usr/local/bin/bosh
echo "Building PCF Tile for Splunk-firehose-nozzle"
tile build