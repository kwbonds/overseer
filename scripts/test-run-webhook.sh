#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

# Runs a test-result listener, which sends a webhook req to our http dump
go run "$DIR/../bridges/webhook-bridge/." \
  -url "http://localhost:10255" \
  -send-test-success true \
  -send-test-recovered true