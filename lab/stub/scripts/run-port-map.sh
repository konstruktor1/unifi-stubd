#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
sh "$script_dir/setup-port-map.sh"

exec /usr/local/bin/unifi-stubd \
    -profile "${UNIFI_STUB_PORTMAP_PROFILE:-us8}" \
    -operation-mode port-map \
    -port-map port=1,interface=pmeth1 \
    -port-map port=2,interface=pmeth2 \
    -port-map port=3,disabled=true \
    -port-map port=4,unmapped=true \
    -port-map port=5,unmapped=true \
    -port-map port=6,unmapped=true \
    -port-map port=7,unmapped=true \
    -port-map port=8,unmapped=true \
    -mac "${UNIFI_STUB_PORTMAP_MAC:-02:15:6d:00:08:22}" \
    -ip "${UNIFI_STUB_PORTMAP_IP:-172.31.242.26}" \
    -hostname "${UNIFI_STUB_PORTMAP_HOSTNAME:-stub-port-map}" \
    -controller http://unifi:8080/inform \
    -interval "${UNIFI_STUB_TEST_INTERVAL:-2s}" \
    -no-discovery \
    -ssh-state /var/lib/unifi-stubd/adoption.env \
    -ssh-host-key /var/lib/unifi-stubd/ssh_host_rsa_key \
    -status-path /var/lib/unifi-stubd/status.json \
    "$@"
