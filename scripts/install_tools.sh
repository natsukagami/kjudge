#!/bin/bash

set -ex

for i in $(grep -Pioh "(?<=_ \")(.+)(?=\")" tools.go) # Fetch all required tools from the tools.go file.
do
    go install -v $i
done

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.24.0
