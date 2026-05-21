#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
sh "$script_dir/setup-bridge-observe.sh"

exec /usr/local/bin/unifi-stubd \
    -config "${UNIFI_STUB_BRIDGE_CONFIG:-/usr/local/lib/unifi-stubd-lab/bridge-observe.config.yaml}" \
    -profile "${UNIFI_STUB_BRIDGE_PROFILE:-us8}" \
    -operation-mode bridge-observe \
    -bridge-observe-bridge "${UNIFI_STUB_TEST_BRIDGE:-stubbr0}" \
    -bridge-observe-uplink-interface "${UNIFI_STUB_TEST_UPLINK:-uplink0}" \
    -bridge-member-port tap101i0=2 \
    -bridge-member-port tap102i0=3 \
    -mac "${UNIFI_STUB_BRIDGE_MAC:-02:15:6d:00:08:21}" \
    -ip "${UNIFI_STUB_BRIDGE_IP:-172.31.242.25}" \
    -hostname "${UNIFI_STUB_BRIDGE_HOSTNAME:-stub-bridge-observe}" \
    -controller http://unifi:8080/inform \
    -interval "${UNIFI_STUB_TEST_INTERVAL:-2s}" \
    -no-discovery \
    -ssh-state /var/lib/unifi-stubd/adoption.env \
    -ssh-host-key /var/lib/unifi-stubd/ssh_host_rsa_key \
    -status-path /var/lib/unifi-stubd/status.json \
    "$@"
