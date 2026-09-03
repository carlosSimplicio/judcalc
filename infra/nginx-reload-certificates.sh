#!/bin/sh
set -eu

reload_interval="${NGINX_RELOAD_INTERVAL:-6h}"

(
    while sleep "$reload_interval"; do
        nginx -s reload || true
    done
) &
