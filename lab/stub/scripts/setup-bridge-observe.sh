#!/bin/sh
set -eu

bridge_name="${UNIFI_STUB_TEST_BRIDGE:-stubbr0}"
uplink_if="${UNIFI_STUB_TEST_UPLINK:-uplink0}"
uplink_peer="${UNIFI_STUB_TEST_UPLINK_PEER:-uplinkp0}"

ensure_iproute() {
    command -v ip >/dev/null 2>&1 && command -v bridge >/dev/null 2>&1
}

create_link_pair() {
    member="$1"
    peer="$2"
    mac="$3"
    ip_addr="$4"

    if ! ip link show "$member" >/dev/null 2>&1; then
        ip link add "$member" type veth peer name "$peer"
    fi
    ip link set "$peer" address "$mac"
    ip link set "$member" master "$bridge_name"
    ip link set "$member" up
    ip link set "$peer" up
    if [ -n "$ip_addr" ] && ! ip addr show dev "$peer" | grep -q "$ip_addr"; then
        ip addr add "$ip_addr" dev "$peer" 2>/dev/null || true
    fi
    bridge fdb replace "$mac" dev "$member" master dynamic
}

if ! ensure_iproute; then
    echo "iproute2 is required in the unifi-stubd lab image" >&2
    exit 1
fi

if ! ip link show "$bridge_name" >/dev/null 2>&1; then
    ip link add name "$bridge_name" type bridge
fi
ip link set "$bridge_name" up

create_link_pair "$uplink_if" "$uplink_peer" "02:00:5e:10:ff:01" ""
create_link_pair "tap101i0" "tap101p0" "02:00:5e:10:01:01" "192.0.2.101/32"
create_link_pair "tap102i0" "tap102p0" "02:00:5e:10:02:01" "192.0.2.102/32"
