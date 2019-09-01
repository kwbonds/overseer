#!/usr/bin/env bash
set -e
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

# --- Config
DEFAULT_DOCKERFILE="$DIR/Dockerfile"

IMAGE_NAME="${1:-overseer}"
DOCKERFILE="${2:-$DEFAULT_DOCKERFILE}"

# ---
DEFAULT_VERSION=$(git describe --tags 2>/dev/null || echo 'master')
FILE_VERSION=$(cat "$DIR/../DOCKER_BUILD_VERSION" || echo "$DEFAULT_VERSION")
VERSION="${DOCKER_BUILD_VERSION:-$FILE_VERSION}"

docker build -t "$IMAGE_NAME:$VERSION" -f "$DIR/$DOCKERFILE" "$DIR/.."
