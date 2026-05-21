#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
config_dir="${UNIFI_STUB_CONFIG_DIR:-/usr/local/share/unifi-stubd-lab/configs}"
hostname="${UNIFI_STUB_BRIDGE_HOSTNAME:-stub-bridge-observe}"
sh "$script_dir/setup-bridge-observe.sh"

exec /usr/local/bin/unifi-stubd \
    -config "${UNIFI_STUB_BRIDGE_CONFIG:-$config_dir/hosts/$hostname/config.yaml}" \
    -profile "${UNIFI_STUB_BRIDGE_PROFILE:-us8}" \
    -bridge-observe-bridge "${UNIFI_STUB_TEST_BRIDGE:-stubbr0}" \
    -bridge-observe-uplink-interface "${UNIFI_STUB_TEST_UPLINK:-uplink0}" \
    -mac "${UNIFI_STUB_BRIDGE_MAC:-02:15:6d:00:08:21}" \
    -ip "${UNIFI_STUB_BRIDGE_IP:-172.31.242.25}" \
    -hostname "$hostname" \
    -interval "${UNIFI_STUB_TEST_INTERVAL:-2s}" \
    "$@"
