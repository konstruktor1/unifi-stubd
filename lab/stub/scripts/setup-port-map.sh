#!/bin/sh
set -eu

ensure_iproute() {
    command -v ip >/dev/null 2>&1
}

create_source_pair() {
    iface="$1"
    peer="$2"
    mac="$3"
    ip_addr="$4"

    if ! ip link show "$iface" >/dev/null 2>&1; then
        ip link add "$iface" type veth peer name "$peer"
    fi
    ip link set "$iface" address "$mac"
    ip link set "$iface" up
    ip link set "$peer" up
    if ! ip addr show dev "$iface" | grep -q "$ip_addr"; then
        ip addr add "$ip_addr/24" dev "$iface" 2>/dev/null || true
    fi
}

if ! ensure_iproute; then
    echo "iproute2 is required in the unifi-stubd lab image" >&2
    exit 1
fi

create_source_pair "pmeth1" "pmpeer1" "02:00:5e:20:00:01" "192.0.2.201"
create_source_pair "pmeth2" "pmpeer2" "02:00:5e:20:00:02" "192.0.2.202"
