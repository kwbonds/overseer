#!/usr/bin/env bash
set -Eeuxmo pipefail
DIR="$(dirname "$(command -v greadlink >/dev/null 2>&1 && greadlink -f "$0" || readlink -f "$0")")"

source "$DIR/docker-build.sh"

docker tag "$IMAGE_NAME:$VERSION" "cmaster11/$IMAGE_NAME:$VERSION"
docker push "cmaster11/$IMAGE_NAME:$VERSION"