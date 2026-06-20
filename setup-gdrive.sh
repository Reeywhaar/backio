#!/bin/sh
set -e
docker build -q -t backio-setup ./setup-gdrive
docker run --rm -it -p 53682:53683 backio-setup
