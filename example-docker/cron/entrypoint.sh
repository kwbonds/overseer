#!/bin/bash
set -e

CRON_FILE="/etc/crontab.list"

if [[ ! -f "$CRON_FILE" ]]; then
  echo "Expecting a valid cron schedule file at $CRON_FILE"
  exit 1
fi

cp "$CRON_FILE" /etc/crontabs/root

# Start cron
crond -l 2 -d 2 -f
