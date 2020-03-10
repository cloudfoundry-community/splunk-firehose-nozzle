#!/usr/bin/env bash

cd testing/integration
. venv/bin/activate
pytest -rfps --durations=10
