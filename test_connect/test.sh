#!/bin/sh
set -e

URL="http://backio:8080/backup"

echo "Testing connectivity to $URL ..."

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$URL")

if [ "$STATUS" = "400" ]; then
	echo "OK — server reachable (got 400 for empty request, as expected)"
	exit 0
else
	echo "FAILED — expected 400, got $STATUS"
	exit 1
fi
