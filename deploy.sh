#!/bin/sh

set -eux

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
(cd api; go build -ldflags="-s -w" -o bootstrap)

stackit up --stack-name ghaoidc --template cfn.yml
