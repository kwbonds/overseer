#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

# Runs a test-result listener, which sends a webhook req to our http dump
go run "$DIR/../bridges/queue-bridge/." \
  -destination-queues=overseer.results.email,overseer.results.webhook
