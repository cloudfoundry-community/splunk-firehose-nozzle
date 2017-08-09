#!/usr/bin/env bash

set -e

if [ "$0" != "./build.sh" ]; then
  echo "build.sh should be run from within the tile directory"
  exit 1
fi

echo "building go binary"
pushd ..
make build
popd

echo "building tile"
tile build
