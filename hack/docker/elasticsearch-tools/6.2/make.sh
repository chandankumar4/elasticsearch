#!/bin/bash
set -xeou pipefail

IMG=elasticsearch-tools
SUFFIX=v2
TAG="6.2-$SUFFIX"
PATCH="6.2.4-$SUFFIX"

docker pull "$DOCKER_REGISTRY/$IMG:$PATCH"

docker tag "$DOCKER_REGISTRY/$IMG:$PATCH" "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
