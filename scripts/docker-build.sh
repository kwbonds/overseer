#!/usr/bin/env bash
set -e
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

# --- Config
IMAGE_NAME="${1:-overseer}"
DOCKERFILE="${2:-Dockerfile}"

# ---
VERSION=$(git describe --tags 2>/dev/null || echo 'master')

docker build -t "$IMAGE_NAME:$VERSION" -f "$DIR/$DOCKERFILE" "$DIR/.."