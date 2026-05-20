#!/bin/sh
set -eu

CONTROLLER_URL="${CONTROLLER_URL:-http://192.0.2.10:8080/inform}"
DEVICE_IP="${DEVICE_IP:-192.0.2.50}"
HOSTNAME_VALUE="${HOSTNAME_VALUE:-auto}"

go run ./cmd/unifi-stubd \
  -profile us16xg \
  -mac auto \
  -ip "$DEVICE_IP" \
  -hostname "$HOSTNAME_VALUE" \
  -controller "$CONTROLLER_URL" \
  -link-speed 10000 \
  -once
