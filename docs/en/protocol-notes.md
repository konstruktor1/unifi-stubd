# Protocol Notes

These notes are working material. They are intentionally pragmatic and must be validated in the lab against concrete UniFi Network versions.

## Discovery

Discovery uses UDP `10001`. Historical implementations send to:

- `255.255.255.255:10001`
- `233.89.188.1:10001`

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

Older lab runs showed controller issues when `uptime` in `mac_table` was missing or too small. Each MAC table entry should therefore include plausible `uptime`.

A profile change should be treated as a new device identity. In practice: use a new fake MAC or `-mac auto`, because UniFi caches model information per MAC and later model changes can stick.

## Model Choice

For the MVP:

- `US8`: simple, no PoE expectations.
- `US8P60`: also small, but PoE fields may be expected.
- `US16P150`: 16-port profile for US-16-150W-like behavior.
- `US16XG`: 16-port 10G profile for aggregation/SFP+ checks.
- `USAGGPRO`: largest controller-known 10G profile validated against older UniFi
  model databases, with 28 10G SFP+ and four 25G SFP28 ports.
- `USW-Pro-XG-48`: largest built-in 10G access switch profile, with mixed 2.5G,
  10G, and 25G SFP28 port groups.

Gateway models such as `UGW3`, `UGW4`, or `UXG` should be checked later.
