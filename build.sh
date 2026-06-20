#!/bin/sh
set -e
DOCKER_DEFAULT_PLATFORM=linux/amd64 docker build -t backio:latest .
