#!/bin/sh
set -eu

python -m python.scripts.init_database --output "$DATABASE_PATH"
exec /app/judcalc-api
