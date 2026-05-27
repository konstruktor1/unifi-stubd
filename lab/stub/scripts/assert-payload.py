#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import sys
from pathlib import Path


BRIDGE_CLIENT_MACS = {
    2: {
        "mac": "02:00:5e:10:01:01",
        "hostname": "lab-client-101",
        "ip": "192.0.2.101",
    },
    3: {
        "mac": "02:00:5e:10:02:01",
        "hostname": "lab-client-102",
        "ip": "192.0.2.102",
    },
}


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: assert-payload.py bridge-observe|port-map|gateway-smoke|management-lan|opnsense-uxg <dry-run-output>", file=sys.stderr)
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
    elif mode == "opnsense-uxg":
        assert_opnsense_uxg(payload)
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
    for index, expected in BRIDGE_CLIENT_MACS.items():
        port = ports[index]
        entries = {
            entry.get("mac"): entry
            for entry in (port.get("mac_table") or [])
            if isinstance(entry, dict)
        }
        client = entries.get(expected["mac"])
        if client is None:
            raise AssertionError(f"port {index} mac_table {list(entries)!r} does not contain {expected['mac']}")
        for field in ("hostname", "ip"):
            if client.get(field) != expected[field]:
                raise AssertionError(f"port {index} client {field} = {client.get(field)!r}, want {expected[field]!r}")
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
    gw_caps = payload.get("gw_caps")
    if not isinstance(gw_caps, dict):
        raise AssertionError(f"gateway gw_caps = {gw_caps!r}, want object")
    for table_name in (
        "ethernet_table",
        "if_table",
        "network_table",
        "config_port_table",
        "ethernet_overrides",
        "port_table",
        "reported_networks",
        "uplink_table",
    ):
        table = payload.get(table_name)
        if not isinstance(table, list) or not table:
            raise AssertionError(f"gateway table {table_name} is missing or empty")
    for table_name in ("internet", "port_overrides", "wan2"):
        if table_name in payload:
            raise AssertionError(f"gateway payload should not define switch-style {table_name}")
    if payload.get("outlet_enabled") is not False:
        raise AssertionError(f"gateway outlet_enabled = {payload.get('outlet_enabled')!r}, want false")
    for table_name in ("outlet_table", "outlet_overrides"):
        table = payload.get(table_name)
        if not isinstance(table, list) or table:
            raise AssertionError(f"gateway {table_name} = {table!r}, want empty list")
    for table_name in ("if_table", "network_table", "config_port_table", "port_table", "reported_networks"):
        table = rows_by_index(payload, table_name, expected_ports=2)
        assert_gateway_row(table_name, table[1], "eth0", "pmeth1")
        assert_gateway_row(table_name, table[2], "eth1", "pmeth2")
    for table_name in ("config_port_table", "port_table"):
        table = rows_by_index(payload, table_name, expected_ports=2)
        assert_gateway_assignment(table_name, table[2], {
            "portconf_id": "portconf-gateway-wan",
            "networkconf_id": "network-gateway-wan",
            "native_networkconf_id": "network-gateway-wan",
            "network_name": "gateway_wan",
            "vlan": 3,
        })
    config_network_wan = payload.get("config_network_wan")
    if not isinstance(config_network_wan, dict):
        raise AssertionError("gateway payload has no config_network_wan")
    if config_network_wan.get("type") != "dhcp":
        raise AssertionError(f"gateway config_network_wan = {config_network_wan!r}, want DHCP")
    assert_gateway_wan_config("config_network_wan", config_network_wan)
    wan1 = payload.get("wan1")
    if not isinstance(wan1, dict):
        raise AssertionError("gateway payload has no wan1 status row")
    assert_gateway_row("wan1", wan1, "eth1", "pmeth2")
    if int(payload.get("uptime", 0)) < 1:
        raise AssertionError(f"gateway uptime = {payload.get('uptime')!r}")
    if int(payload.get("time_ms", 0)) <= 0:
        raise AssertionError(f"gateway time_ms = {payload.get('time_ms')!r}")


def assert_opnsense_uxg(payload: dict[str, object]) -> None:
    hostname = payload.get("hostname")
    if not isinstance(hostname, str) or not hostname.startswith("opnsense-uxg"):
        raise AssertionError(f"opnsense gateway hostname = {hostname!r}, want opnsense-uxg*")
    if payload.get("type") != "uxg":
        raise AssertionError(f"opnsense gateway type = {payload.get('type')!r}, want uxg")
    if payload.get("model") != "UXGPRO":
        raise AssertionError(f"opnsense gateway model = {payload.get('model')!r}, want UXGPRO")
    if int(payload.get("num_port", 0)) != 4:
        raise AssertionError(f"opnsense gateway num_port = {payload.get('num_port')!r}, want 4")
    for table_name in (
        "ethernet_table",
        "if_table",
        "network_table",
        "config_port_table",
        "ethernet_overrides",
        "port_table",
        "reported_networks",
        "uplink_table",
    ):
        table = payload.get(table_name)
        if not isinstance(table, list) or not table:
            raise AssertionError(f"opnsense gateway table {table_name} is missing or empty")
    for table_name in ("internet", "port_overrides"):
        if table_name in payload:
            raise AssertionError(f"opnsense gateway payload should not define switch-style {table_name}")
    rows = rows_by_port_idx(payload, "if_table")
    assert_gateway_row("if_table", rows[3], "eth2", "ixl0")
    assert_gateway_row("if_table", rows[4], "eth3", "vtnet0")
    for table_name in ("network_table", "reported_networks"):
        rows = rows_by_port_idx(payload, table_name)
        assert_gateway_row(table_name, rows[3], "eth2", "ixl0")
        assert_gateway_row(table_name, rows[4], "eth3", "vtnet0")
    for table_name in ("ethernet_table", "config_port_table", "port_table"):
        table = rows_by_index(payload, table_name, expected_ports=4)
        if table[3].get("ifname") != "eth2":
            raise AssertionError(f"{table_name} port 3 ifname = {table[3].get('ifname')!r}, want eth2")
        if table[4].get("ifname") != "eth3":
            raise AssertionError(f"{table_name} port 4 ifname = {table[4].get('ifname')!r}, want eth3")
        if table_name in ("config_port_table", "ethernet_overrides", "port_table"):
            if table[3].get("source_interface") != "ixl0":
                raise AssertionError(f"{table_name} port 3 source_interface = {table[3].get('source_interface')!r}, want ixl0")
            if table[4].get("source_interface") != "vtnet0":
                raise AssertionError(f"{table_name} port 4 source_interface = {table[4].get('source_interface')!r}, want vtnet0")
    port_table = rows_by_index(payload, "port_table", expected_ports=4)
    assert_gateway_switchport_neighbor("port_table", port_table[4], "server-lan1")
    ethernet_overrides = rows_by_port_idx(payload, "ethernet_overrides")
    expected_override_groups = {
        1: ("eth0", "Unassigned", ""),
        2: ("eth1", "Unassigned", ""),
        3: ("eth2", "WAN", "ixl0"),
        4: ("eth3", "LAN", "vtnet0"),
    }
    if set(ethernet_overrides) != set(expected_override_groups):
        raise AssertionError(f"ethernet_overrides ports = {sorted(ethernet_overrides)}, want [1, 2, 3, 4]")
    for index, (ifname, networkgroup, source) in expected_override_groups.items():
        row = ethernet_overrides[index]
        if row.get("ifname") != ifname:
            raise AssertionError(f"ethernet_overrides port {index} ifname = {row.get('ifname')!r}, want {ifname}")
        if row.get("networkgroup") != networkgroup:
            raise AssertionError(
                f"ethernet_overrides port {index} networkgroup = {row.get('networkgroup')!r}, want {networkgroup}"
            )
        if source and row.get("source_interface") != source:
            raise AssertionError(
                f"ethernet_overrides port {index} source_interface = {row.get('source_interface')!r}, want {source}"
            )
        if index in (1, 2) and row.get("disabled") is not True:
            raise AssertionError(f"ethernet_overrides port {index} disabled = {row.get('disabled')!r}, want true")
        if index in (3, 4) and "disabled" in row:
            raise AssertionError(f"ethernet_overrides port {index} should not be disabled: {row!r}")
    config_network_lan = payload.get("config_network_lan")
    if not isinstance(config_network_lan, dict):
        raise AssertionError("opnsense gateway payload has no config_network_lan")
    lan_ip = payload.get("lan_ip")
    if not isinstance(lan_ip, str) or not lan_ip:
        raise AssertionError(f"opnsense lan_ip = {lan_ip!r}, want populated LAN IP")
    if config_network_lan.get("cidr") != f"{lan_ip}/24":
        raise AssertionError(f"config_network_lan cidr = {config_network_lan.get('cidr')!r}, want {lan_ip}/24")
    if config_network_lan.get("ifname") != "eth3":
        raise AssertionError(f"config_network_lan ifname = {config_network_lan.get('ifname')!r}, want eth3")
    if config_network_lan.get("port_idx") != 4:
        raise AssertionError(f"config_network_lan port_idx = {config_network_lan.get('port_idx')!r}, want 4")
    if payload.get("has_eth1") is not True:
        raise AssertionError(f"opnsense has_eth1 = {payload.get('has_eth1')!r}, want true")
    assignments = {
        3: {
            "portconf_id": "portconf-opnsense-wan",
            "networkconf_id": "network-opnsense-wan",
            "native_networkconf_id": "network-opnsense-wan",
            "network_name": "opnsense_wan",
            "vlan": 3,
        },
        4: {
            "portconf_id": "portconf-opnsense-lan",
            "networkconf_id": "network-opnsense-lan",
            "native_networkconf_id": "network-opnsense-lan",
            "network_name": "opnsense_lan",
            "vlan": 1,
        },
    }
    for table_name in ("config_port_table", "port_table"):
        table = rows_by_index(payload, table_name, expected_ports=4)
        for port, expected in assignments.items():
            if hostname == "opnsense-uxg-sfp-lab":
                assert_gateway_assignment(table_name, table[port], expected)
    config_network_wan = payload.get("config_network_wan")
    if not isinstance(config_network_wan, dict):
        raise AssertionError("opnsense gateway payload has no config_network_wan")
    assert_gateway_wan_config("config_network_wan", config_network_wan)
    if config_network_wan.get("ifname") != "eth2":
        raise AssertionError(f"config_network_wan ifname = {config_network_wan.get('ifname')!r}, want eth2")
    wan1 = payload.get("wan1")
    if not isinstance(wan1, dict):
        raise AssertionError("opnsense gateway payload has no wan1 status row")
    assert_gateway_row("wan1", wan1, "eth2", "ixl0")
    wans = payload.get("wans")
    if not isinstance(wans, list) or not wans:
        raise AssertionError("opnsense gateway payload has no wans inventory")
    if wans[0].get("interface") != "eth2" or wans[0].get("port") != 3:
        raise AssertionError(f"opnsense wans[0] = {wans[0]!r}, want interface eth2 port 3")
    if payload.get("uplink") != "eth2":
        raise AssertionError(f"opnsense uplink = {payload.get('uplink')!r}, want eth2")
    assert_no_host_ifnames(payload, ("ixl0", "vtnet0", "igb0"))


def assert_gateway_row(table_name: str, row: dict[str, object], ifname: str, source: str) -> None:
    if row.get("ifname") != ifname:
        raise AssertionError(f"{table_name} ifname = {row.get('ifname')!r}, want {ifname}")
    if row.get("source_interface") != source:
        raise AssertionError(f"{table_name} source_interface = {row.get('source_interface')!r}, want {source}")


def assert_gateway_wan_config(table_name: str, row: dict[str, object]) -> None:
    if row.get("type") != "dhcp":
        raise AssertionError(f"{table_name} type = {row.get('type')!r}, want dhcp")
    if not row.get("ip") or not row.get("netmask"):
        raise AssertionError(f"{table_name} has no ip/netmask: {row!r}")
    if row.get("speed") != "auto":
        raise AssertionError(f"{table_name} speed = {row.get('speed')!r}, want auto")
    if row.get("autoneg") is not True or row.get("full_duplex") is not True:
        raise AssertionError(f"{table_name} autoneg/full_duplex invalid: {row!r}")
    if row.get("ifname") and row.get("ifname") not in {"eth0", "eth1", "eth2", "eth3", "eth4", "eth5"}:
        raise AssertionError(f"{table_name} ifname = {row.get('ifname')!r}, want controller ethN")
    if "source_interface" in row:
        raise AssertionError(f"{table_name} must not expose host source interface, got {row!r}")


def assert_gateway_assignment(table_name: str, row: dict[str, object], expected: dict[str, object]) -> None:
    for key, want in expected.items():
        if row.get(key) != want:
            raise AssertionError(f"{table_name} {key} = {row.get(key)!r}, want {want!r}")


def assert_gateway_switchport_neighbor(table_name: str, row: dict[str, object], hostname: str) -> None:
    last_connection = row.get("last_connection")
    if not isinstance(last_connection, dict):
        raise AssertionError(f"{table_name} last_connection = {last_connection!r}, want {hostname}")
    if last_connection.get("hostname") != hostname:
        raise AssertionError(f"{table_name} last_connection hostname = {last_connection.get('hostname')!r}, want {hostname}")
    if last_connection.get("type") != "usw":
        raise AssertionError(f"{table_name} last_connection type = {last_connection.get('type')!r}, want usw")
    mac_table = row.get("mac_table")
    if not isinstance(mac_table, list) or not mac_table:
        raise AssertionError(f"{table_name} mac_table = {mac_table!r}, want switchport neighbor")
    if not any(isinstance(entry, dict) and entry.get("hostname") == hostname and entry.get("type") == "usw" for entry in mac_table):
        raise AssertionError(f"{table_name} mac_table lacks {hostname}: {mac_table!r}")


def assert_no_host_ifnames(payload: dict[str, object], blocked: tuple[str, ...]) -> None:
    stack: list[object] = [payload]
    while stack:
        value = stack.pop()
        if isinstance(value, dict):
            if value.get("ifname") in blocked:
                raise AssertionError(f"host interface leaked into ifname: {value!r}")
            stack.extend(value.values())
        elif isinstance(value, list):
            stack.extend(value)


def ports_by_index(payload: dict[str, object]) -> dict[int, dict[str, object]]:
    return rows_by_index(payload, "port_table", expected_ports=8)


def rows_by_port_idx(payload: dict[str, object], table_name: str) -> dict[int, dict[str, object]]:
    table = payload.get(table_name)
    if not isinstance(table, list):
        raise AssertionError(f"{table_name} is not a list")
    out: dict[int, dict[str, object]] = {}
    for row in table:
        if not isinstance(row, dict):
            continue
        try:
            index = int(row.get("port_idx", 0))
        except (TypeError, ValueError):
            continue
        if index > 0:
            out[index] = row
    return out


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
