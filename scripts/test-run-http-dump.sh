#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

# Run a local lhttp istener, which dumps all requests to stdout
go run "$DIR/test-http-dump/main.go" -p 10255