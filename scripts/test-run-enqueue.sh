#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

TMP_FILE=$(mktemp)
cat >"$TMP_FILE" <<EOL
fake-name1 must run dumb-test with retries 0 with duration-min 2s with duration-max 10s
fake-name2 must run dumb-test with retries 0 with duration-min 2s with duration-max 10s
fake-name3 must run dumb-test with retries 0 with duration-min 2s with duration-max 10s
fake-name4 must run dumb-test with retries 0 with duration-min 2s with duration-max 10s
fake-name5 must run dumb-test with retries 0 with duration-min 2s with duration-max 10s
fake-name6 must run dumb-test with retries 0 with duration-min 2s with duration-max 10s
EOL

go run "$DIR/.." enqueue "$TMP_FILE"