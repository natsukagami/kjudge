#!/usr/bin/env bash

set -ev

# Frontend templates
cd frontend && yarn && yarn run --prod build && cd ..
# Go source code
go generate
