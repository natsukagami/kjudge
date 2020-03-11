#!/bin/bash

set -e

# Perform a production build
#
# Turn off debugging modes
sed -i 's/^debug/# debug/' fileb0x.yaml

# Re-generate
scripts/generate.sh

# Build
go build -tags "production" -o kjudge cmd/kjudge/main.go

# Reset the debugging modes
sed -i 's/^# debug/debug/' fileb0x.yaml
