#!/bin/sh
# Usage: send-backup.sh <path-to-tar>
# Env:   BACKUP_URL, BACKUP_NAME, BACKUP_SUBDIRECTORY, BACKUP_PROVIDER, BACKUP_TOKEN
set -e

TAR_PATH="${1:?Usage: send-backup.sh <path-to-tar>}"
BACKUP_URL="${BACKUP_URL:?BACKUP_URL is required}"
BACKUP_NAME="${BACKUP_NAME:?BACKUP_NAME is required}"
BACKUP_SUBDIRECTORY="${BACKUP_SUBDIRECTORY:?BACKUP_SUBDIRECTORY is required}"
BACKUP_PROVIDER="${BACKUP_PROVIDER:?BACKUP_PROVIDER is required}"
BACKUP_TOKEN="${BACKUP_TOKEN:?BACKUP_TOKEN is required}"

curl -sf -X POST "$BACKUP_URL" \
	-H "Authorization: Bearer ${BACKUP_TOKEN}" \
	-F "backup=@${TAR_PATH}" \
	-F "name=${BACKUP_NAME}" \
	-F "subdirectory=${BACKUP_SUBDIRECTORY}" \
	-F "provider=${BACKUP_PROVIDER}"
