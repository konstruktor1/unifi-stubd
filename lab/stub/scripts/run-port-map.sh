#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
config_dir="${UNIFI_STUB_CONFIG_DIR:-/usr/local/share/unifi-stubd-lab/configs}"
hostname="${UNIFI_STUB_PORTMAP_HOSTNAME:-stub-port-map}"
sh "$script_dir/setup-port-map.sh"

exec /usr/local/bin/unifi-stubd \
    -config "${UNIFI_STUB_PORTMAP_CONFIG:-$config_dir/hosts/$hostname/config.yaml}" \
    -profile "${UNIFI_STUB_PORTMAP_PROFILE:-us8}" \
    -mac "${UNIFI_STUB_PORTMAP_MAC:-02:15:6d:00:08:22}" \
    -ip "${UNIFI_STUB_PORTMAP_IP:-172.31.242.26}" \
    -hostname "$hostname" \
    -interval "${UNIFI_STUB_TEST_INTERVAL:-2s}" \
    "$@"
