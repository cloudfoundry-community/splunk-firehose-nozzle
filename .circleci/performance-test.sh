#!/usr/bin/env bash

cd testing/integration
. venv/bin/activate
pytest -v -m Perf_Binary
pytest -v -m Perf_Romote