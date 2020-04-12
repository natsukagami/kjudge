#!/bin/bash

set -e

FILE=${FILE:-kjudge.db}

if test -f "$FILE"; then
    echo "Setting up the database will override all your unwritten database data."
    echo "Backing up as $FILE.bak"
    mv $FILE $FILE.bak
fi

go run cmd/migrate/main.go -file $FILE < "test/data.sql"
