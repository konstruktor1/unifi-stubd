#!/bin/sh
set -eu
umask 077

# The disposable helper receives credentials only on stdin. Its private runtime
# file never becomes a Compose environment value, argument or image layer.
if [ "${UNIFI_STUB_CONTROLLER_AUTH_STDIN:-0}" = 1 ]; then
  auth_dir=/run/unifi-stubd-controller
  mkdir -m 700 "$auth_dir"
  auth_file="$auth_dir/auth.json"
  head -c 65537 >"$auth_file"
  size=$(wc -c <"$auth_file")
  if [ "$size" -eq 0 ] || [ "$size" -gt 65536 ]; then
    printf 'controller authentication input must contain 1..65536 bytes\n' >&2
    exit 2
  fi
  chmod 600 "$auth_file"
  export UNIFI_STUB_CONTROLLER_AUTH_FILE="$auth_file"
fi
exec "$@"
