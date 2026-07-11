# Protocol Notes

These notes are working material. They are intentionally pragmatic and must be validated in the lab against concrete UniFi Network versions.

## Discovery

Discovery uses UDP `10001`. Historical implementations send to:

- `255.255.255.255:10001`
- `233.89.188.1:10001`

Some FreeBSD/OPNsense lab routes cannot send to the all-ones broadcast address.
For those cases `discovery_targets` can set explicit UDP targets such as the
LAN broadcast address, for example `192.0.2.255:10001`. Empty
`discovery_targets` keeps the defaults above.

`discovery_interface` can pin the local source interface for discovery sends.
It is explicit on purpose; the daemon does not guess which lab or management
network should see discovery traffic.

Packet shape:

```text
u8  version
u8  packet_type
u16 payload_length_be
TLV payload
```

TLV shape:

```text
u8  type
u16 length_be
[]  value
```

Interesting TLVs from older implementations:

| Type | Meaning |
| --- | --- |
| `0x01` | MAC address |
| `0x02` | MAC + IPv4 |
| `0x0a` | Uptime |
| `0x0b` | Hostname / name |
| `0x12` | Announcement sequence |
| `0x13` | Serial / MAC |
| `0x15` | Model identifier |
| `0x16` | Firmware version |
| `0x17` | Default/factory flag |

## Inform

Inform uses HTTP POST:

```text
POST http://<controller>:8080/inform
Content-Type: application/x-binary
User-Agent: AirControl Agent v1.0
```

Binary header:

```text
0x00  4 bytes  magic "TNBU"
0x04  4 bytes  packet version, usually 0 or 1
0x08  6 bytes  device MAC
0x0e  2 bytes  flags
0x10 16 bytes  IV / nonce
0x20  4 bytes  payload version, usually 1
0x24  4 bytes  payload length
0x28  n bytes  payload
```

Known flags:

| Flag | Meaning |
| --- | --- |
| `0x01` | encrypted |
| `0x02` | zlib |
| `0x04` | snappy |
| `0x08` | AES-GCM in newer UniFi versions |

Historically AES-CBC + PKCS#7 + zlib was common. Newer controllers/devices can set `use_aes_gcm=true` in `mgmt_cfg`.

## Adoption

Minimal flow:

1. Stub starts in factory state.
2. Stub sends discovery and/or inform with the default key.
3. Controller shows `Pending Adoption`.
4. After Adopt is clicked, the controller replies with `_type: "setparam"`.
5. `mgmt_cfg` contains values such as `authkey`, `cfgversion`, `stun_url`, `mgmt_url`, `use_aes_gcm`.
6. Stub stores `authkey` and continues with that key.
7. Controller later sends `noop`, `setparam`, provisioning, or restart commands.

Observed UXG-Pro 5.0.16 controller-lab flow:

| Phase | Request/response shape | Stub behavior |
| --- | --- | --- |
| Pre-adoption | Firmware informs use the default key and receive HTTP `404` until an admin clicks `Adopt`. | Keep `default=true`, `state=1`. |
| First accepted inform | Controller responds HTTP `200` with `_type: "setparam"` and `mgmt_cfg`. | Parse and persist `authkey`, `cfgversion`, and `use_aes_gcm=true`. |
| Key switch | Firmware immediately sends the next inform with the adopted key and AES-GCM. | Use the adopted key for all later inform traffic. |
| Adopted inform | Firmware reports `default=false`, then `state=2`. | Treat the device as connected once the controller returns `noop`. |
| System config | Controller sends another `_type: "setparam"` containing `system_cfg`. | Record safe metadata only; do not apply host users, firewall, routes, certificates, tokens, or secrets. |
| WireGuard key request | Controller sends `_type: "cmd"`, `cmd: "wg-genkeypair"`, and a `request_id` without arguments. | Generate an X25519 key pair in memory and attach a firmware-compatible notification to the next ordinary inform: `inform_as_notif=true`, `notif_reason="cmd"`, and a `notif_payload` containing `event_string`, the correlated `request_id`, and `params.private_key`/`params.public_key`. Retry until one HTTP `200`; do not persist or log the keys or request ID. |
| Forget/reset | Controller sends a forget, delete, remove, or restore-default command. | Clear only local adoption state, return to factory/default inform, and never reset the host. |
| Steady state | Controller returns `_type: "noop"` with `interval` and `include_blocks`. | Clamp a positive interval to local operator bounds and reset transport-failure backoff; a 10-second interval was observed. |

The first `mgmt_cfg` observed in the lab contained `cfgversion`, `stun_url`,
`mgmt_url`, `authkey`, `use_aes_gcm=true`, and `report_crash=true`. It did not
need to include a new `inform_url`; the firmware continued to use the existing
inform URL and only changed the key/cipher context.

The adopted `system_cfg` shape was a JSON string with top-level `ubntconf` and
`udapi` keys. The `udapi` object contained interfaces, services, system users,
firewall sets/filter chains/settings, static routes, and Radius profiles. This
is provisioning data for a real gateway, not a safe set of host instructions
for `unifi-stubd`.

`wg-genkeypair` is the only controller command with an executable result. It
uses Go's in-process X25519 implementation and performs no shell execution,
file write, service action, or host-network mutation. Missing request IDs,
arguments, parameters, malformed requests, and all other commands are ignored.

Heartbeat policy remains local. Packaged defaults start at 10 seconds, accept
controller hints from 5 through 60 seconds, cap failure backoff at 300 seconds,
and apply 20 percent symmetric jitter. Transport failures double the effective
delay; a completed HTTP exchange resets the failure count. If the configured
normal interval exceeds the backoff maximum, it remains the lower bound so a
failure never accelerates inform traffic.

Alternative adoption over SSH:

```text
/usr/bin/syswrapper.sh set-adopt <inform_url> <authkey>
```

For `unifi-stubd`, a small SSH shim is possible; L3 inform adoption is still the better MVP.

## Minimal Switch Payload

Important fields:

| Field | Purpose |
| --- | --- |
| `mac` | stable fake MAC |
| `ip` | visible IP |
| `hostname` | display name |
| `model` | e.g. `US8`, `US8P60`, `US16P150` |
| `model_display` | controller display |
| `version` | firmware version |
| `serial` | usually MAC without colons |
| `num_port` | switch port count |
| `cfgversion` | controller config version |
| `uptime` | status/connected state |
| `time` | device time |
| `if_table` | management interface |
| `if_table[].management_vlan` | optional configured management VLAN metadata |
| `ethernet_table` | controller-side Ethernet/port-count table |
| `port_table` | switch ports |
| `port_table[].speed` | port speed in Mbps, e.g. `1000` or `10000` |
| `port_table[].media` | media marker, e.g. `GE` or `SFP+` |
| `port_table[].mac_table` | observed clients/VMs, with optional `hostname` and `ip` metadata |
| `port_table[].last_connection` | direct downstream infrastructure or last attached client for a port |
| `uplink.uplink_mac` | explicit direct upstream MAC for switch payloads |
| `uplink.uplink_device_name` | optional direct upstream device label |
| `uplink.uplink_remote_port` | optional configured or LLDP-derived upstream port number |
| `sys_stats` | CPU/RAM/load |

Mixed-speed switch profiles should report the complete physical port layout in
`port_table`. For example, `USW-Pro-XG-48` is modeled with 16 2.5G RJ45 ports,
32 10G RJ45 ports, and four 25G SFP28 ports. The management `if_table` speed
uses the selected uplink port speed.

Profiles should stay hardware-shaped. Lab assignments such as "this port is
WAN", "this port is LAN", or "this port represents backup WAN" belong in
`port_overrides[].role` and `port_overrides[].network_group`.

`management_lan` is modeled as safe payload metadata first. It can additionally
point switch management to a preexisting VLAN interface. In that mode the
interface IPv4 address becomes the reported management IP, discovery is sourced
from that interface, and inform HTTP connections are locally bound to the same
address. The daemon still does not create tagged host interfaces or apply VLAN
changes received through controller provisioning.

Older lab runs showed controller issues when `uptime` in `mac_table` was
missing or too small. Each MAC table entry should therefore include plausible
`uptime`. Inform payloads should also report monotonically increasing top-level
`uptime`, `sys_stats.uptime`, `system-stats.uptime`, and gateway `time_ms` /
`uptime_str` fields so controller UI freshness does not depend on a static
first-boot payload. Synthetic ports without an OS source should report a small
heartbeat counter delta and non-zero `rx_bytes-r` / `tx_bytes-r`; ports backed
by a real observed interface should keep the real OS counters.

Direct infrastructure and client MAC rows are resolved separately. The active
`uplink_port` selects the upstream edge and switch payloads report it through
`uplink` and `last_uplink`; infrastructure on other ports remains downstream.
Configured neighbors override LLDP, while FDB observations never select a
direction. Optional `remote_port` supplies `uplink_remote_port`. Real device
and interface MACs remain valid identities throughout this projection.

A profile change should be treated as a new device identity. In practice: use a new fake MAC or `-mac auto`, because UniFi caches model information per MAC and later model changes can stick.

## Model Choice

For the MVP:

- `US8`: simple, no PoE expectations.
- `US8P60`: also small, but PoE fields may be expected.
- `US16P150`: 18-port profile for US-16-150W-like behavior, with 16 1G RJ45
  ports and two 1G SFP uplinks.
- `US16XG`: 16-port 10G profile for aggregation/SFP+ checks, with twelve
  1/10G SFP+ ports and four 1/10G RJ45 ports.
- `USAGGPRO`: largest controller-known 10G profile validated against older UniFi
  model databases, with 28 10G SFP+ and four 25G SFP28 ports.
- `USW-Pro-XG-48`: largest built-in 10G access switch profile, with mixed 2.5G,
  10G, and 25G SFP28 port groups.
- `UGW3`: experimental legacy UniFi Security Gateway identity profile. It
  reports device type `ugw`, model `UGW3`, and three 1G ports named `WAN 1`,
  `LAN 1`, and `WAN 2 / LAN 2`.
- `UXG`: experimental UniFi Gateway Lite identity profile, exposed through the
  `uxg-lite` profile. It reports device type `uxg`, model `UXG`, firmware
  `5.0.16.30689`, and two 1G ports named `LAN` and `WAN`.
- `UXGPRO`: experimental UniFi Next-Generation Gateway Pro identity profile. It
  reports device type `uxg`, model `UXGPRO`, firmware `5.0.16.30689`, two 1G
  RJ45 ports named `WAN` and `LAN`, and two 10G SFP+ ports named `WAN2` and
  `LAN2`. The default active uplink remains `WAN`; remap SFP+ internet labs
  with `uplink_port` and `port_overrides`.
- `UCGF`: experimental UniFi Cloud Gateway Fiber identity profile, exposed
  through the `ucg-fiber` profile. It reports device type `udm`, firmware
  `5.0.16`, four 2.5G RJ45 LAN ports, one 10G RJ45 `WAN2` port, one 10G SFP+
  `WAN` port, and one 10G SFP+ LAN port.

`UGW3`, `UXG`, `UXGPRO`, and `UCGF` are only identity/profile stubs in this
release. A full gateway payload still needs WAN/LAN state, routing, DHCP,
firewall, DPI, and health fields. The current gateway payload sends gateway
tables such as `network_table`, a read-only physical `port_table`, `uplink`,
`uplink_table`, `config_port_table`, and `ethernet_overrides`. Explicit
operator-provided WAN/LAN assignment metadata may be mirrored, but it is not
learned from controller provisioning or applied to host networking. Gateway
models such as `UGW4`, other Cloud Gateway devices, and EFG should be checked
later.
