#!/bin/sh

VERSION=$(date "+%y%m%d%H%M%S")-$(echo $(git rev-parse HEAD) | cut -c1-4)
echo $VERSION > ./.version
# Restore go dependencies
go mod download
# build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o exec
# build docker image
sudo docker build -t standardcore/smsresender:$VERSION .
