#!/bin/sh
set -eu

script_dir="$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)"
config_dir="${UNIFI_STUB_CONFIG_DIR:-/usr/local/share/unifi-stubd-lab/configs}"
sh "$script_dir/setup-bridge-observe.sh"

exec /usr/local/bin/unifi-stubd \
  -config "${UNIFI_STUB_BRIDGE_CONFIG:-$config_dir/hosts/stub-bridge-observe/config.yaml}" \
  -mac "${UNIFI_STUB_BRIDGE_MAC:-02:15:6d:00:08:21}" \
  -ip "${UNIFI_STUB_BRIDGE_IP:-172.31.242.25}" \
  "$@"
