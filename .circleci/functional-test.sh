#!/usr/bin/env bash

cd testing/integration
. venv/bin/activate
pytest --durations=10
