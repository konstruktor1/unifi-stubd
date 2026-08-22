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

The configured Inform URL is an explicit trust boundary. The client does not
follow HTTP `301`, `302`, `307`, or `308`; it keeps the first response as a
non-200 lifecycle result, does not decrypt or dispatch its body, and never
adopts `Location`. Only a validated `mgmt_cfg.inform_url` can change the
persisted management URL. For a controller-provided value, validated means at
most 2,048 bytes, an absolute `http` or `https` URL with a host, and no
userinfo, query, or fragment. This URL policy is a deliberate Stub safety
decision, not a recovered firmware claim. It governs learned and persisted
management URLs; it is not a host/CIDR allowlist, does not redact URL paths,
and does not redefine the operator-configured initial controller URL. Paths
must therefore not carry secrets.

Binary header:

```text
0x00  4 bytes  magic "TNBU"
0x04  4 bytes  packet version, exactly 0
0x08  6 bytes  device MAC
0x0e  2 bytes  flags
0x10 16 bytes  IV / nonce
0x20  4 bytes  payload version, exactly 1
0x24  4 bytes  payload length
0x28  n bytes  payload
```

All integer fields are big-endian. The decoder requires the exact invariant
`len(packet) == 40 + payload_length`; short headers, unsupported versions,
overflow-sized declarations, truncated payloads, and trailing bytes are
rejected before decryption or JSON processing.

Known flags:

| Flag | Meaning |
| --- | --- |
| `0x01` | encrypted |
| `0x02` | zlib |
| `0x04` | snappy |
| `0x08` | AES-GCM in newer UniFi versions |

Historically AES-CBC + PKCS#7 + zlib was common. The implemented CBC form uses
AES-128-CBC with an exact 16-byte key and a fresh cryptographically random
16-byte IV for every packet. Plaintext is padded strictly according to PKCS#7;
an already aligned plaintext receives a complete additional padding block.
CBC ciphertext must therefore be a non-empty multiple of 16 bytes, and
malformed or inconsistent padding is rejected.

Compression is an explicit, already-resolved wire-policy input. When enabled,
zlib runs before PKCS#7 padding and encryption; otherwise CBC carries raw JSON
before padding. The low-level Inform client does not infer zlib from zero-value
options. Zlib-expanded plaintext and encoder input are bounded to 16 MiB. The
exact build-to-wire selection remains a separate profile-activation contract.

Newer controllers/devices can set `use_aes_gcm=true` in `mgmt_cfg`. The
product's AES-128-GCM envelope uses the protocol-specific 16-byte nonce,
authenticates the complete final 40-byte header as additional data, appends a
16-byte tag, and includes that tag in `payload_length`. Header, nonce,
ciphertext, and tag tampering fail closed; a conventional 12-byte GCM nonce is
not accepted for this wire format.

Controller responses use a direction-specific receive policy. The client binds
response flag `0x08` to the cipher mode already selected for the request and
rejects either CBC/GCM mismatch instead of negotiating from the response
header. After matching-mode authentication and decryption, the plaintext is
passed directly to JSON processing: response-side zlib and Snappy are not
accepted. A policy failure exposes no plaintext. The bounded client result may
retain the HTTP status, raw response bytes, and copied TNBU packet metadata;
the persisted runtime status retains only body presence, response flags, and
the sanitized error. The direction-neutral decoder remains available for
device requests and protocol tools, where zlib request bodies are valid.

The receive path also has explicit local resource boundaries. A TNBU header may
declare at most 16 MiB plus 64 KiB of encoded payload, which covers bounded zlib
framing plus cipher overhead; larger declarations are rejected before the body
is copied. Decryption or decompression may return at most 16 MiB of plaintext.
Before controller JSON is materialized, it is limited to 4 MiB, 64 nested
objects/arrays, and 65,536 tokens (names, values, and delimiters). The same JSON
limits are applied again to a JSON document carried as an encoded `system_cfg`
string. These limits are stub safety policy, not inferred firmware thresholds;
an exceeded boundary fails the complete response without an adoption-state
update.

Runtime status retains only the latest sanitized Inform exchange. It includes
operational result, selected cipher metadata, response type and sizes, and the
effective schedule, but no key, packet body, secret, or raw controller text.

Transport, response-decode, JSON-syntax, dispatcher, and every non-200 result
advance the same bounded backoff. Missing, non-string, and unknown `_type`
values are fixed-text dispatcher failures; untrusted discriminator text is not
retained. Only an empty HTTP 200 or a 200 with an accepted controller object
resets the failure sequence. A failure never triggers a second request under a
different cipher.
The current product acceptance is defined by the named product requirements;
it does not require an in-process controller journey.

## Product Evidence Boundary

Firmware and reverse-engineering matrices are research input, not a product
registry. They stay in the companion research record and do not create runtime
behavior, tests, or compatibility claims by themselves.

Only behavior required by the product contract is implemented. Unknown build
identities, profile overrides, unsupported controller values, and unrecognized
commands fail closed instead of borrowing behavior from a neighboring device.
Controller interval hints likewise apply only through an exact reviewed
built-in identity; otherwise the operator-selected interval remains active.

## Adoption

Minimal flow:

1. The factory/default identity uses wire state `1` and the public compatibility
   key.
2. Discovery or Inform makes that identity available for adoption.
3. An accepted `setparam` may persist bounded management identity. When it
   changes the authkey or active Inform URL, the new identity starts in wire
   state `0` and the next request uses the validated individual key.
4. A later successful `noop` or non-identity-changing `setparam` confirms the
   active identity as managed wire state `2`.

The Inform wire vocabulary is exact:

| Value | Name | Product meaning |
| --- | --- | --- |
| `0` | `Unknown` | A setparam carried an authkey or changed the active Inform URL, and that identity has not yet been confirmed by a later successful response. |
| `1` | `Default` | Factory or Pending Adoption identity using the public compatibility key. |
| `2` | `Managed` | A successful `noop` or `setparam` has confirmed the active management identity. |
| `3` | `Selfrun` | Status vocabulary only; it grants no automatic lifecycle or host action. |
| `4` | `Upgrade` | Status vocabulary only; it grants no firmware, service, or host action. |

The implemented identity-changing path is `1 -> 0 -> 2`. A later accepted
authkey or Inform-URL change returns the active identity to `0` until it is
confirmed again.

The pending/default identity explicitly uses the 16-byte public compatibility
key `ba86f2bbe107c7c57eb5f2690775c712`; the daemon does not derive or generalize
it at runtime. URL-only advanced adoption remains keyless `provisioning` with
`default=true` and `state=1`. The low-level Inform client never substitutes the
public value for a missing key, and the same bytes are rejected as an individual
adopted key. Corrupt state, adopted state without a valid individual key, and
failed state without a key stop before any HTTP request.

Controller response handling stays deliberately small:

- accepted `setparam` data may update only bounded adoption identity, cipher,
  configuration version, and validated Inform URL state;
- `system_cfg` is treated as untrusted provisioning data and is never applied
  to host users, services, firewall, routes, certificates, or networking;
- `_type: "upgrade"`, `_type: "cmd"`, and unknown response types are ignored
  without executing a command or changing the advertised product version; and
- `noop` metadata may affect only bounded process-local reporting and
  scheduling. It cannot set the host clock or gain host authority.

Heartbeat policy remains local. Controller interval hints require an exact
reviewed built-in identity and are clamped to operator bounds; every other
identity keeps the operator-selected base interval. Transport, decode, JSON,
dispatcher, and non-200 failures use bounded backoff. Only an empty HTTP 200 or
an accepted controller object resets the failure count.

Alternative adoption over SSH:

```text
/usr/bin/syswrapper.sh set-adopt <inform_url> <authkey>
```

`unifi-stubd` includes an optional SSH shim for this command; inform-based
adoption remains the normal path.

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
| `if_table[].rx_bytes`, `tx_bytes`, packet/error counters | cumulative physical counters mirrored from the selected uplink |
| measured `if_table[].bytes-r`, `rx_bytes-r`, `tx_bytes-r` | selected-uplink rates in byte/s used by the normalized current-uplink path |
| measured `if_table[].rx_rate`, `tx_rate` | selected-uplink rates in bit/s used by Network 10.5 live statistics |
| `ethernet_table` | controller-side Ethernet/port-count table |
| `port_table` | switch ports |
| `port_table[].speed` | port speed in Mbps, e.g. `1000` or `10000` |
| `port_table[].media` | media marker, e.g. `GE` or `SFP+` |
| measured uplink `port_table[].rx_rate`, `tx_rate` | physical RX/TX rate in bit/s; omitted from downstream ports |
| `port_table[].mac_table` | observed clients/VMs, with optional `hostname` and `ip` metadata |
| `port_table[].last_connection` | direct downstream infrastructure or last attached client for a port |
| top-level `uplink` | management-interface selector such as `eth0`; UniFi Network combines it with `port_table[].is_uplink` to build the current uplink object |
| `last_uplink.uplink_mac` | explicit direct upstream MAC for switch payloads |
| `last_uplink.uplink_remote_port` | optional configured or LLDP-derived upstream port number |
| root `bytes`, `rx_bytes`, `tx_bytes` | cumulative selected-uplink traffic summary |
| root `bytes-r`, `rx_bytes-r`, `tx_bytes-r` | selected-uplink rates in byte/s (`B/s`) |
| root `rx_rate`, `tx_rate` | selected-uplink rates in bit/s |
| `sys_stats` | CPU/RAM/load |

Mixed-speed switch profiles should report the complete physical port layout in
`port_table`. For example, `USW-Pro-XG-48` is modeled with 16 2.5G RJ45 ports,
32 10G RJ45 ports, and four 25G SFP28 ports. The management `if_table` speed
uses the selected uplink port speed.

The selected physical uplink's cumulative byte, packet, and error counters are
also mirrored into the management `if_table` row named by `uplink`. This is the
native switch shape and uses physical direction. With explicit rate reporting,
that row receives measured `*-r` byte/s and `rx_rate`/`tx_rate` bit/s, while the
selected `port_table` row also receives the measured bit/s fields. Network 10.5
reads its normalized live uplink rate from the selected interface path. The
management-rate and selected-port bit/s fields are omitted when no real rate
source is enabled and are never synthesized. Legacy synthetic per-port `*-r`
values are Stub-only compatibility data. They are not measured telemetry and
do not satisfy `PR-METRICS`; acceptance uses only the selected local
interface source, including its real zero samples.

The switch root traffic summary is emitted only when rate reporting is enabled
and is based on the same selected uplink. Root direction is opposite to the
physical interface counter names: physical uplink RX maps to root TX
(Download), while physical uplink TX maps to root RX (Upload). The first
sample, idle/down links, and counter resets report zero; traffic that never
crosses the selected uplink is not part of this summary.

For switches, the Inform wire value is the string selector `uplink: "eth0"`,
not an object. UniFi Network normalizes the selected `port_table` row into the
current REST `uplink` object. Network 10.4.57 retains the root cumulative
counters but omits root rate keys from that normalized response. Network
10.5.62 normalizes the measured interface byte/s and bit/s fields into the
current uplink and official latest-statistics API; its root fields remain zero
or absent. `source_interface` remains local payload metadata on the logical and
port views; it is not a controller observation.

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
first-boot payload. Synthetic ports without an OS source may retain
compatibility-only freshness data, but it must not be described or tested as a
product metric. Ports backed by a real observed interface keep the real OS
counters and rates.

Direct infrastructure and client MAC rows are resolved separately. The active
`uplink_port` selects the upstream edge and switch payloads report it through
`uplink` and `last_uplink`; infrastructure on other ports remains downstream.
Configured neighbors override LLDP, while FDB observations never select a
direction. Optional `remote_port` supplies `uplink_remote_port`. Real device
and interface MACs remain valid identities throughout this projection.

A profile change should be treated as a new device identity. In practice: use a new fake MAC or `-mac auto`, because UniFi caches model information per MAC and later model changes can stick.
