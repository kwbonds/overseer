#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

KUBECONFIG="$1"
CONFIG="$2"

go run "$DIR/.." k8s-event-watcher -verbose -tag Testtt -kubeconfig "$KUBECONFIG" -watcher-config "$CONFIG"
