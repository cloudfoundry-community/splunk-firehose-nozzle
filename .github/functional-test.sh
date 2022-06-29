#!/usr/bin/env bash

cd testing/integration
. venv/bin/activate
pytest -v -m Critical