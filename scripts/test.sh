#!/usr/bin/env sh
set -e

# Build frontend
cd frontend && yarn && yarn build && cd ..

# Test everything
go test -tags production ./...
