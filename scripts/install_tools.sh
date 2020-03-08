#!/bin/bash

set -ex

for i in $(grep -Pioh "(?<=_ \")(.+)(?=\")" tools.go) # Fetch all required tools from the tools.go file.
do
    go install -v $i
done

