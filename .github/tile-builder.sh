#!/usr/bin/env bash
echo "start"
set -e
wget https://github.com/cf-platform-eng/tile-generator/releases/download/v15.0.2/pcf_linux-64bit > /dev/null 2>&1
chmod +x pcf_linux-64bit
sudo mv pcf_linux-64bit /usr/local/bin/tile
python3 -m venv tile-generator-env > /dev/null 2>&1
source tile-generator-env/bin/activate
echo "Installing tile-generator..."
pip install wheel
pip install jinja2==3.0.3
pip install setuptools
pip install tile-generator
cd tile
echo "Installing bosh..."
wget https://github.com/cloudfoundry/bosh-cli/releases/download/v5.5.0/bosh-cli-5.5.0-linux-amd64 > /dev/null 2>&1
mv bosh-cli-5.5.0-linux-amd64 bosh
chmod +x ./bosh
sudo mv ./bosh /usr/local/bin/bosh
echo "Building PCF Tile for Splunk-firehose-nozzle"
tile build
