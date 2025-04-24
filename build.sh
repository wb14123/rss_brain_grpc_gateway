#!/bin/sh

set -e

echo "Generate proto files ..."
buf generate

echo "Build Go binary"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" main.go

tag=`date +'%Y-%m-%d-%s'`
image="docker-hosted.binwang.me:30008/rss_brain_grpc_gateway:$tag"
echo "Build docker image $image"
docker build -t $image .
docker push $image
