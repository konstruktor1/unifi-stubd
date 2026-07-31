#!/bin/sh
set -eu

bridge_name=stubbr0
uplink_if=uplink0
uplink_peer=uplinkp0

ip link add name "$bridge_name" type bridge
ip link set "$bridge_name" up
ip link add "$uplink_if" type veth peer name "$uplink_peer"
ip link set "$uplink_if" master "$bridge_name"
ip link set "$uplink_if" up
ip link set "$uplink_peer" up

# Keep the traffic fixture entirely inside the disposable container. Bind the
# documentation /30 to the bridge and the peer so the gate can select a source
# on either side of uplink0.
ip addr add 198.51.100.1/30 dev "$bridge_name"
ip addr add 198.51.100.2/30 dev "$uplink_peer"
bridge_mac="$(cat "/sys/class/net/$bridge_name/address")"
peer_mac="$(cat "/sys/class/net/$uplink_peer/address")"

# Two unreachable documentation targets create deterministic one-way streams
# in opposite directions. They exercise uplink RX and TX without contacting an
# external host or depending on ICMP replies.
ip route add 203.0.113.1/32 dev "$uplink_peer" src 198.51.100.2
ip neigh add 203.0.113.1 lladdr "$bridge_mac" dev "$uplink_peer" nud permanent
ip route add 203.0.113.2/32 dev "$bridge_name" src 198.51.100.1
ip neigh add 203.0.113.2 lladdr "$peer_mac" dev "$bridge_name" nud permanent
