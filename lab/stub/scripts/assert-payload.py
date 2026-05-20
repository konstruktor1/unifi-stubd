#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import sys
from pathlib import Path


BRIDGE_CLIENT_MACS = {
    2: "02:00:5e:10:01:01",
    3: "02:00:5e:10:02:01",
}


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: assert-payload.py bridge-observe|port-map|gateway-smoke|management-lan <dry-run-output>", file=sys.stderr)
        return 2
    mode = sys.argv[1]
    output = Path(sys.argv[2]).read_text(encoding="utf-8")
    payload = extract_payload(output)
    if mode == "bridge-observe":
        assert_bridge_observe(payload)
    elif mode == "port-map":
        assert_port_map(payload)
    elif mode == "gateway-smoke":
        assert_gateway_smoke(payload)
    elif mode == "management-lan":
        assert_management_lan(payload)
    else:
        print(f"unknown assertion mode {mode!r}", file=sys.stderr)
        return 2
    return 0


def extract_payload(output: str) -> dict[str, object]:
    marker = "minimal_inform_payload_json:\n"
    if marker not in output:
        raise AssertionError("dry-run output does not contain inform payload marker")
    return json.loads(output.split(marker, 1)[1])


def assert_bridge_observe(payload: dict[str, object]) -> None:
    assert payload.get("hostname") == "stub-bridge-observe", payload.get("hostname")
    ports = ports_by_index(payload)
    for index, expected_mac in BRIDGE_CLIENT_MACS.items():
        port = ports[index]
        macs = [entry.get("mac") for entry in (port.get("mac_table") or [])]
        if expected_mac not in macs:
            raise AssertionError(f"port {index} mac_table {macs!r} does not contain {expected_mac}")
    if not ports[1].get("is_uplink"):
        raise AssertionError("port 1 should remain the bridge-observe uplink")


def assert_management_lan(payload: dict[str, object]) -> None:
    assert_bridge_observe(payload)
    expected_ip = os.environ.get("UNIFI_STUB_BRIDGE_IP", "")
    if payload.get("ip") != expected_ip:
        raise AssertionError(f"payload ip = {payload.get('ip')!r}, want {expected_ip!r}")
    if int(payload.get("management_vlan", 0)) != 42:
        raise AssertionError(f"management_vlan = {payload.get('management_vlan')!r}, want 42")
    if_table = payload.get("if_table")
    if not isinstance(if_table, list) or not if_table:
        raise AssertionError("payload has no if_table")
    management = if_table[0]
    if not isinstance(management, dict):
        raise AssertionError(f"if_table[0] is not an object: {management!r}")
    if int(management.get("management_vlan", 0)) != 42:
        raise AssertionError(f"if_table management_vlan = {management.get('management_vlan')!r}, want 42")
    if int(management.get("vlan", 0)) != 42:
        raise AssertionError(f"if_table vlan = {management.get('vlan')!r}, want 42")


def assert_port_map(payload: dict[str, object]) -> None:
    assert payload.get("hostname") == "stub-port-map", payload.get("hostname")
    ports = ports_by_index(payload)
    if ports[1].get("source_interface") != "pmeth1":
        raise AssertionError(f"port 1 source_interface = {ports[1].get('source_interface')!r}")
    if ports[2].get("source_interface") != "pmeth2":
        raise AssertionError(f"port 2 source_interface = {ports[2].get('source_interface')!r}")
    if ports[3].get("up") is not False or int(ports[3].get("speed", -1)) != 0:
        raise AssertionError(f"port 3 should be disabled, got {ports[3]!r}")
    if ports[4].get("source_interface"):
        raise AssertionError(f"port 4 should be unmapped, got {ports[4]!r}")
    if ports[4].get("up") is not True or int(ports[4].get("speed", 0)) == 0:
        raise AssertionError(f"port 4 should keep profile defaults, got {ports[4]!r}")


def assert_gateway_smoke(payload: dict[str, object]) -> None:
    assert payload.get("hostname") == "stub-gateway-smoke", payload.get("hostname")
    if payload.get("type") != "uxg":
        raise AssertionError(f"gateway type = {payload.get('type')!r}, want uxg")
    if payload.get("model") != "UXG":
        raise AssertionError(f"gateway model = {payload.get('model')!r}, want UXG")
    if "port_table" in payload:
        raise AssertionError("gateway payload should not expose switch port_table")
    for table_name in ("if_table", "network_table", "uplink_table", "config_port_table", "ethernet_overrides", "reported_networks"):
        table = payload.get(table_name)
        if not isinstance(table, list) or not table:
            raise AssertionError(f"gateway table {table_name} is missing or empty")
    for table_name in ("if_table", "network_table", "config_port_table", "ethernet_overrides", "reported_networks"):
        table = rows_by_index(payload, table_name, expected_ports=2)
        assert_gateway_row(table_name, table[1], "eth0", "LAN", "pmeth1")
        assert_gateway_row(table_name, table[2], "eth1", "WAN", "pmeth2")
    if int(payload.get("uptime", 0)) < 1:
        raise AssertionError(f"gateway uptime = {payload.get('uptime')!r}")
    if int(payload.get("time_ms", 0)) <= 0:
        raise AssertionError(f"gateway time_ms = {payload.get('time_ms')!r}")


def assert_gateway_row(table_name: str, row: dict[str, object], ifname: str, networkgroup: str, source: str) -> None:
    if row.get("ifname") != ifname:
        raise AssertionError(f"{table_name} ifname = {row.get('ifname')!r}, want {ifname}")
    if row.get("networkgroup") != networkgroup:
        raise AssertionError(f"{table_name} networkgroup = {row.get('networkgroup')!r}, want {networkgroup}")
    if row.get("source_interface") != source:
        raise AssertionError(f"{table_name} source_interface = {row.get('source_interface')!r}, want {source}")


def ports_by_index(payload: dict[str, object]) -> dict[int, dict[str, object]]:
    return rows_by_index(payload, "port_table", expected_ports=8)


def rows_by_index(payload: dict[str, object], table_name: str, expected_ports: int) -> dict[int, dict[str, object]]:
    table = payload.get(table_name)
    if not isinstance(table, list):
        raise AssertionError(f"payload has no {table_name} list")
    out: dict[int, dict[str, object]] = {}
    for row in table:
        if not isinstance(row, dict):
            continue
        index = row.get("port_idx")
        if isinstance(index, int):
            out[index] = row
    missing = [index for index in range(1, expected_ports + 1) if index not in out]
    if missing:
        raise AssertionError(f"missing {table_name} rows {missing}")
    return out


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except AssertionError as exc:
        print(f"payload assertion failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
