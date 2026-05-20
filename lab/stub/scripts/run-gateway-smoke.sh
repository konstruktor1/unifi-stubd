#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
sh "$script_dir/setup-port-map.sh"

exec /usr/local/bin/unifi-stubd \
    -profile "${UNIFI_STUB_GATEWAY_PROFILE:-uxg-lite}" \
    -operation-mode port-map \
    -port-map port=1,interface=pmeth1 \
    -port-map port=2,interface=pmeth2 \
    -mac "${UNIFI_STUB_GATEWAY_MAC:-02:15:6d:00:08:23}" \
    -ip "${UNIFI_STUB_GATEWAY_IP:-172.31.242.27}" \
    -hostname "${UNIFI_STUB_GATEWAY_HOSTNAME:-stub-gateway-smoke}" \
    -controller http://unifi:8080/inform \
    -interval "${UNIFI_STUB_TEST_INTERVAL:-2s}" \
    -no-discovery \
    -ssh-state /var/lib/unifi-stubd/adoption.env \
    -ssh-host-key /var/lib/unifi-stubd/ssh_host_rsa_key \
    -status-path /var/lib/unifi-stubd/status.json \
    "$@"
