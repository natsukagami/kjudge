#!/usr/bin/env sh
set -e

# Test everything
go test -tags production ./...
