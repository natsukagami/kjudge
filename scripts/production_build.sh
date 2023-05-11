#!/usr/bin/env sh

set -e

# Perform a production build
#

# Re-generate
scripts/generate.sh

# Build
go build -tags "production" -o kjudge cmd/kjudge/main.go
