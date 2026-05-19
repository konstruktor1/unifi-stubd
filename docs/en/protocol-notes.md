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
| Steady state | Controller returns `_type: "noop"` with `interval` and `include_blocks`. | Continue inform on the requested interval; a 10-second interval was observed. |

The first `mgmt_cfg` observed in the lab contained `cfgversion`, `stun_url`,
`mgmt_url`, `authkey`, `use_aes_gcm=true`, and `report_crash=true`. It did not
need to include a new `inform_url`; the firmware continued to use the existing
inform URL and only changed the key/cipher context.

The adopted `system_cfg` shape was a JSON string with top-level `ubntconf` and
`udapi` keys. The `udapi` object contained interfaces, services, system users,
firewall sets/filter chains/settings, static routes, and Radius profiles. This
is provisioning data for a real gateway, not a safe set of host instructions
for `unifi-stubd`.

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
| `port_table[].mac_table` | observed clients/VMs |
| `sys_stats` | CPU/RAM/load |

Mixed-speed switch profiles should report the complete physical port layout in
`port_table`. For example, `USW-Pro-XG-48` is modeled with 16 2.5G RJ45 ports,
32 10G RJ45 ports, and four 25G SFP28 ports. The management `if_table` speed
uses the selected uplink port speed.

Profiles should stay hardware-shaped. Lab assignments such as "this port is
WAN", "this port is LAN", or "this port represents backup WAN" belong in
`port_overrides[].role` and `port_overrides[].network_group`.

`management_vlan` is modeled as safe payload metadata first. It reports the
management VLAN to the controller but does not create tagged host interfaces or
apply VLAN changes received through controller provisioning.

Older lab runs showed controller issues when `uptime` in `mac_table` was missing or too small. Each MAC table entry should therefore include plausible `uptime`.

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
tables such as `network_table`, `uplink`, and `uplink_table`, but the
controller may still render gateway ports from its internal model instead of a
switch-style `port_table`. Gateway models such as `UGW4`, other Cloud Gateway
devices, and EFG should be checked later.
