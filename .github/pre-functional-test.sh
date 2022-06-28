#!/usr/bin/env bash

sudo apt-get install python3.7
sudo apt-get install python3-pip
cd testing/integration
pip3 install virtualenv
virtualenv venv
source venv/bin/activate
pip3 install -r requirements.txt
