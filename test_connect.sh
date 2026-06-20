#!/bin/sh
set -e
docker build -q -t backio-test ./test_connect
docker run --rm --network backup-net backio-test
