#!/bin/bash
set -x

VERSION=${VERSION:-"v0.1"}
NAMESPACE=${NAMESPACE:-"kencochrane"}
IMAGE_NAME=${IMAGE_NAME:-"swarm-prune"}
DOCKER_TAG_LATEST=${DOCKER_TAG_LATEST:-"yes"}
FULL_IMAGE_NAME=$NAMESPACE/$IMAGE_NAME

# build the binary
./build.sh

docker build -t $FULL_IMAGE_NAME:$VERSION -f Dockerfile .
docker push $FULL_IMAGE_NAME:$VERSION

if [[ "$DOCKER_TAG_LATEST" == "yes" ]] ; then
    docker push $FULL_IMAGE_NAME:latest
fi

rm -f bin/swarm-prune
