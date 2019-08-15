#!/usr/bin/env bash
set -Eeumo pipefail

NOW=$(date +%s)

kubectl -n overseer create job --from=cronjob/overseer-enqueue "overseer-enqueue-manual-$NOW"