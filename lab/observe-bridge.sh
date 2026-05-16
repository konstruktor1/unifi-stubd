#!/bin/sh
set -eu

PATH="/sbin:/usr/sbin:/bin:/usr/bin:${PATH:-}"

LAB_BRIDGE="${UNIFI_STUBD_LAB_BRIDGE:-stubbr0}"
TAP_IFACE="${UNIFI_STUBD_LAB_TAP_IFACE:-tap101i0}"
TAP_PEER="${UNIFI_STUBD_LAB_TAP_PEER:-stub101p}"
VETH_IFACE="${UNIFI_STUBD_LAB_VETH_IFACE:-veth200i0}"
VETH_PEER="${UNIFI_STUBD_LAB_VETH_PEER:-stub200p}"
TAP_MACS="${UNIFI_STUBD_LAB_TAP_MACS:-00:11:22:33:44:55 00:11:22:33:44:66}"
VETH_MACS="${UNIFI_STUBD_LAB_VETH_MACS:-00:11:22:33:44:77}"

need_command() {
  command -v "$1" >/dev/null 2>&1 || {
    printf '%s\n' "missing required command: $1" >&2
    exit 1
  }
}

link_exists() {
  ip link show dev "$1" >/dev/null 2>&1
}

delete_link() {
  if link_exists "$1"; then
    ip link del "$1"
  fi
}

ensure_bridge() {
  if ! link_exists "$LAB_BRIDGE"; then
    ip link add name "$LAB_BRIDGE" type bridge
  fi
  ip link set "$LAB_BRIDGE" up
}

ensure_veth() {
  iface="$1"
  peer="$2"
  if ! link_exists "$iface"; then
    delete_link "$peer"
    ip link add "$iface" type veth peer name "$peer"
  fi
  ip link set "$iface" master "$LAB_BRIDGE"
  ip link set "$iface" up
  ip link set "$peer" up
}

add_fdb_entries() {
  iface="$1"
  macs="$2"
  for mac in $macs; do
    bridge fdb replace "$mac" dev "$iface" master static
  done
}

up() {
  need_command ip
  need_command bridge
  ensure_bridge
  ensure_veth "$TAP_IFACE" "$TAP_PEER"
  ensure_veth "$VETH_IFACE" "$VETH_PEER"
  add_fdb_entries "$TAP_IFACE" "$TAP_MACS"
  add_fdb_entries "$VETH_IFACE" "$VETH_MACS"
}

down() {
  need_command ip
  delete_link "$TAP_IFACE"
  delete_link "$VETH_IFACE"
  delete_link "$LAB_BRIDGE"
}

status() {
  need_command ip
  need_command bridge
  ip -br link | grep -E "^(${LAB_BRIDGE}|${TAP_IFACE}|${TAP_PEER}|${VETH_IFACE}|${VETH_PEER})" || true
  if link_exists "$LAB_BRIDGE"; then
    bridge fdb show br "$LAB_BRIDGE"
  fi
}

case "${1:-status}" in
  up|start)
    up
    ;;
  down|stop)
    down
    ;;
  restart)
    down
    up
    ;;
  status)
    status
    ;;
  *)
    printf '%s\n' "usage: $0 {up|down|restart|status}" >&2
    exit 2
    ;;
esac
