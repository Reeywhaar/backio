#!/bin/sh
set -e

printf "Google Drive folder ID (from folder URL): "
read -r FOLDER_ID

# rclone binds its auth server to 127.0.0.1:53682 (container loopback).
# Docker Desktop on macOS can't forward to loopback, so we proxy:
# Mac:53682 → container:53683 (socat) → rclone:127.0.0.1:53682
socat TCP-LISTEN:53683,bind=0.0.0.0,reuseaddr,fork TCP:127.0.0.1:53682 &
SOCAT_PID=$!
trap 'kill $SOCAT_PID 2>/dev/null; exit' INT TERM EXIT

echo ""
echo "A browser authorization URL will appear below."
echo "Open it in your Mac browser — the callback is received automatically."
echo "Once authorized, copy the JSON token printed between the arrows and paste it here."
echo ""

rclone authorize "drive"

echo ""
printf "Paste the JSON token from above: "
read -r TOKEN_JSON

mkdir -p /root/.config/rclone
cat > /root/.config/rclone/rclone.conf << EOF
[gdrive]
type = drive
scope = drive
token = ${TOKEN_JSON}
root_folder_id = ${FOLDER_ID}
EOF

echo ""
echo "=== Add this to your .env ==="
echo ""
printf "RCLONE_CONF_BASE64="
base64 /root/.config/rclone/rclone.conf | tr -d '\n'
echo ""
