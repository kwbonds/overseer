#!/usr/bin/env bash
set -Eeuxmo pipefail

RULE="$1"

(echo "$RULE" | kubectl -n overseer run --restart=Never -i enqueue-manual \
  --image="cmaster11/overseer:release-1.11-4" \
  -- enqueue -redis-host redis:6379 - || true) &&
  kubectl -n overseer delete pod enqueue-manual
