#!/bin/sh
set -e

if [ -f .env ]; then
	. ./.env
fi

if [ -z "$RCLONE_CONF_BASE64" ]; then
	echo "ERROR: RCLONE_CONF_BASE64 is not set" >&2
	exit 1
fi

docker network create backup-net 2>/dev/null || true

docker run -d \
	--network backup-net \
	--name backio \
	--restart unless-stopped \
	-e RCLONE_CONF_BASE64="$RCLONE_CONF_BASE64" \
	-e PORT="${PORT:-8080}" \
	backio:latest
