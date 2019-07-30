#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

# --- Config
IMAGE_NAME="overseer"

# ---
VERSION=$(git describe --tags 2>/dev/null || echo 'master')

docker build -t "$IMAGE_NAME:$VERSION" "$DIR"