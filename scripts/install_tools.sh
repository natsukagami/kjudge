#!/usr/bin/env sh

set -ex

grep -Pioh "(?<=_ \")(.+)(?=\")" < tools.go | while IFS= read -r line # Fetch all required tools from the tools.go file.
do
    go install -v "${line}"
done

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)"/bin v1.62.2
