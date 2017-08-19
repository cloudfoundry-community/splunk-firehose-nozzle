#!/bin/bash

docker images

alias orca='docker run --rm -it --name orca -e USER=kchen -v /var/run/docker.sock:/var/run/docker.sock  -v /Users/kchen/.orca:/root/.orca -v /Users/kchen/.ssh:/root/.ssh -v /Users/kchen/.docker:/root/.docker -v $(pwd -P):/orca-home repo.splunk.com/splunk/products/orca'
orca create --sc nozzle
