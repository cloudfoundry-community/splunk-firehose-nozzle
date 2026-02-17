#!/usr/bin/env bash

cd ..
set -e
echo "Local tile build script"

# Read current version from tile-history.yml
CURRENT_VERSION=$(grep '^version:' tile/tile-history.yml | awk '{print $2}')

if [ -z "$CURRENT_VERSION" ]; then
    echo "Could not read version from tile/tile-history.yml"
    exit 1
fi

echo "Building version: $CURRENT_VERSION"

echo "Building Go binary..."
env GOOS=linux GOARCH=amd64 go build -o splunk-firehose-nozzle -ldflags "-X main.version=$CURRENT_VERSION" ./main.go

# Make sure binary is executable
chmod +x splunk-firehose-nozzle

# Check if tile command exists
if ! command -v tile &> /dev/null; then
    echo "tile command not found. Installing tile-generator..."
    pip3 install tile-generator
fi

# Check if bosh command exists (required by tile-generator)
if ! command -v bosh &> /dev/null; then
    echo "bosh command not found. Please install bosh-cli:"
    echo "  brew install cloudfoundry/tap/bosh-cli"
    echo "  # or download from: https://github.com/cloudfoundry/bosh-cli/releases"
    exit 1
fi

# Build the tile
echo "Building tile..."
cd tile
tile build $CURRENT_VERSION
cd ..

echo "Build completed!"
ls -la tile/product/*.pivotal