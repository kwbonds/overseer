#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

REDIS_QUEUE_KEY="${REDIS_QUEUE_KEY:-overseer.results}"

# Runs a test-result listener, which sends a webhook req to our http dump
go run "$DIR/../bridges/email-bridge/." \
  -redis-queue-key "$REDIS_QUEUE_KEY" \
  -smtp-host "smtp.elasticemail.com" \
  -smtp-port 2525 \
  -smtp-username "$EMAIL_USERNAME" \
  -smtp-password "$EMAIL_PASSWORD" \
  -email "$EMAIL_TO"
