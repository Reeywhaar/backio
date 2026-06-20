#!/bin/sh
set -e

if [ -z "$RCLONE_CONF_BASE64" ]; then
	echo "ERROR: RCLONE_CONF_BASE64 is not set" >&2
	exit 1
fi

mkdir -p /root/.config/rclone
printf '%s' "$RCLONE_CONF_BASE64" | base64 -d > /root/.config/rclone/rclone.conf
chmod 600 /root/.config/rclone/rclone.conf

exec /server
