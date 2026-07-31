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

Each daemon run writes a bounded, sanitized `inform_session` to runtime status.
Its ordered events contain the exact request flags, the non-secret key source
(`default` or `adopted`), HTTP status, body presence, whether a TNBU response
was decoded and its JSON parsed, response flags/type, and the selected cipher.
The compatibility CBC-fallback fields remain false because one heartbeat sends
exactly one locally selected cipher request. No key, packet body, MAC, secret,
or arbitrary raw controller text is retained there. Fixed allowlisted response
types and bounded structural metadata remain visible; they are not a claim
that all controller-provided values are absent from every product surface. An
HTTP `404` or malformed HTTP `200` is visible as a failed exchange, while an
empty HTTP `200` is a successful heartbeat without decoded protocol evidence.

Transport, response-decode, JSON-syntax, dispatcher, and every non-200 result
advance the same bounded backoff. Missing, non-string, and unknown `_type`
values are fixed-text dispatcher failures; untrusted discriminator text is not
retained. Only an empty HTTP 200 or a 200 with an accepted controller object
resets the failure sequence. A failure never triggers a second request under a
different cipher.
The current product acceptance is defined by the named product requirements;
it does not require an in-process controller journey.

## Exact Build Policy Registry

The product contains two typed registries. `ProtocolProfile` binds complete
static evidence, not a model label or firmware-version prefix, to a
version-independent behavior ID. `ProtocolBehaviorProfile` separately composes
wire, discovery, field-group, command, positive-capability, and presence
policies:

| Evidence | Exact build class | Wire | Discovery | Commands | Field groups / presence | Positive capabilities |
| --- | --- | --- | --- | --- | --- | --- |
| `DT-001` | UXG Pro 5.0.16 | shared safe zlib TNBU subset | unsupported until exact gateway frames exist | top-level 7 / direct 70 | UXG producer and request gates | gateway management plus producer telemetry |
| `DT-002` | UDM Pro SE 5.1.19 | shared safe zlib TNBU subset | unsupported until exact gateway frames exist | top-level 7 / direct 65 | UDM saved-config/probe gates | gateway management plus saved-config/probe vocabulary |
| `DT-008` | Mainline BCM 7.4.1 | shared safe zlib subset; Snappy disabled | shared mainline v2 builder | top-level 7 / direct 40, host mutation rejected | common mainline groups plus BCM presence | switch/mainline plus BCM table vocabulary |
| `DT-009` | Mainline MSCC 7.4.1 | shared safe zlib subset | shared mainline v2 builder | same direct 40 policy | common mainline groups plus MSCC presence | switch/mainline plus MSCC platform contract |
| `DT-010` | Switch-WAN 7.3.109 | shared safe zlib subset | separate WAN v2 alternate-builder contract | top-level 7 / direct 41, host mutation rejected | WAN/model/source gates | source-gated WAN/RPS vocabulary; no action authority |
| `DT-012` | current `usmini2` 2.1.6 | raw JSON TNBU | shared current-ESP directed broadcast | proven top-level 5; direct `cmd` unsupported | common low-resource groups plus usmini2 presence | current-ESP base; no positive PSE claim |
| `DT-013..015` | current Ultra/M25G 2.1.8 | shared zlib with raw fallback | shared current-ESP directed broadcast | same top-level 5 policy; direct `cmd` unsupported | common low-resource groups plus separate presence | shared current-ESP base; positive PSE vocabulary only for Ultra and M25G8 |

The IDs describe semantic contracts and contain no artifact version. Equal
complete compositions cannot receive duplicate behavior IDs, while individual
policy components are reused wherever the evidence establishes the same
semantics. Missing capability flags only disable unproven product behavior;
they do not claim that hardware is absent. In particular, M25G5 has no positive
PSE policy from this app analysis, but its external hardware remains unknown.

Every selected component exposes its axis/selector, establishing `DT-*`
records, and recovered principle. A component needs executable product proof
only when a named product requirement selects its behavior. Platform selectors
are limited to the typed architecture classes from `device-types.tsv`; any
other value fails closed.

`DT-003..007` and the M487 component `DT-011` are known evidence identities
whose protocol semantics remain unsupported or `UNKNOWN`; their six policy
axes are therefore all `unsupported`. The M487 and current usmini2 entries
share a firmware container but not an agent or wire policy; likewise the BCM
and MSCC entries share an outer firmware image but retain exact component
identities and distinct complete behavior compositions. Blank hashes, one-byte
hash changes, adjacent versions, and mixed components therefore fail closed
instead of choosing a neighbor. Build markers are exact observed firmware,
rootfs, or component markers rather than normalized product labels.

The registries are metadata and a strict lookup boundary. Applying each selector
in Inform, Discovery, command dispatch, and payload rendering belongs to the
corresponding implementation requirements; these registries alone do not alter
the existing user-facing shape profiles or runtime traffic.

### Exact Controller Interval Policy

`REQ-LIFE-010` is a separate, explicit consumer of the evidence registry. It
enables positive controller interval hints only when the built-in source,
profile name, model, version, and board identity form one exact tuple:

| Built-in profile | Exact model | Exact version | Interval evidence ID | Positive `noop.interval` policy |
| --- | --- | --- | --- | --- |
| `us8` | `US8` | `7.4.1.16850` | `DT-008` | enabled; 30-second firmware minimum |
| `us8p60` | `US8P60` | `7.4.1.16850` | `DT-008` | enabled; 30-second firmware minimum |
| `us16p150` | `US16P150` | `7.4.1.16850` | `DT-008` | enabled; 30-second firmware minimum |
| `us24p250` | `US24P250` | `7.4.1.16850` | `DT-008` | enabled; 30-second firmware minimum |
| `us48p500` | `US48P500` | `7.4.1.16850` | `DT-008` | enabled; 30-second firmware minimum |
| `uxgpro` | `UXGPRO` | `5.0.16.30689` | `DT-001` | enabled; 1-second firmware minimum |
| `ucg-fiber` | `UCGF` | `5.0.16` | `DT-004` | fail closed; exact evidence is known but no firmware minimum is cataloged |
| `ugw3` | `UGW3` | `4.4.57.5578372` | `DT-007` | fail closed; exact evidence is known but no firmware minimum is cataloged |
| `uxg-lite` | `UXG` | `5.0.16.30689` | `DT-006` | fail closed; exact evidence is known but no firmware minimum is cataloged |
| `us16xg` | `US16XG` | `7.4.1.16850` | not bound | fail closed; no exact interval-policy binding |
| `usaggpro` | `USAGGPRO` | `7.4.1.16850` | not bound | fail closed; no exact interval-policy binding |
| `usw-pro-xg-48` | `USWProXG48` | `7.4.1.16850` | not bound | fail closed; no exact interval-policy binding |

These decisions concern interval hints only. They do not promote profile
maturity or enable all protocol behavior associated with an evidence ID. In
particular, the 30-second minimum belongs to each listed board/profile tuple;
it is not a build-wide `DT-008` default. External profiles, adjacent versions,
model or version overrides, and lookalike identities fail closed instead of
borrowing a neighboring decision.

For an enabled tuple, the effective lower bound is
`max(firmware minimum, operator minimum)` and the effective upper bound is
`min(operator maximum, 300 seconds)`. For every disabled tuple, a positive hint
is recorded as ignored and the operator-selected base interval remains active.
Configuration fails validation when that effective upper bound is below the
effective lower bound.
The dry-run plan and sanitized status expose whether hints are supported, the
evidence ID, firmware and effective minima, or the fixed fail-closed reason.

This LIFE-010 implementation exists only in the current dirty-worktree product
candidate. Reviewed source-hash binding, T4 acceptance of the same bytes on the
designated controller test machine, separate G3 reachability from `main`, and
release evidence remain open; this is not a T4, G3, or release claim.

## Adoption

Minimal flow:

1. Stub starts in factory/default wire state `1`.
2. Stub sends discovery and/or inform with the default key.
3. Controller shows `Pending Adoption`.
4. After Adopt is clicked, the controller replies with `_type: "setparam"`.
5. `mgmt_cfg` contains values such as `authkey`, `cfgversion`, `stun_url`, `mgmt_url`, `use_aes_gcm`.
6. Stub stores `authkey`; the immediately following request uses that key,
   GCM, and wire state `0`.
7. In the captured path an accepted `noop` completes the transition to managed
   wire state `2`.
8. The stored capture later contains `setparam(system_cfg)` and `noop`.
   Additional command types are known from static analysis, not from this flow.

The Inform wire vocabulary is exact:

| Value | Name | Product meaning |
| --- | --- | --- |
| `0` | `Unknown` | A setparam carried an authkey or changed the active Inform URL, and that identity has not yet been confirmed by a later successful response. |
| `1` | `Default` | Factory or Pending Adoption identity using the public compatibility key. |
| `2` | `Managed` | A successful `noop` or `setparam` has confirmed the active management identity. |
| `3` | `Selfrun` | Status vocabulary only; it grants no automatic lifecycle or host action. |
| `4` | `Upgrade` | Status vocabulary only; it grants no firmware, service, or host action. |

The dynamically observed main path is therefore `1 -> 0 -> 2`, not a direct
`1 -> 2` transition. Static firmware analysis adds only the general rule: a
successful `noop` or `setparam` enters Managed first; an authkey-bearing
`setparam` or an Inform-URL change then returns the reporter to Unknown. The capture does
not establish restart, repeated-identity, or keyless-noop semantics.

For the gateway path observed in UXG Pro 5.0.16 and corroborated in UDM Pro SE
5.1.19, the public compatibility key is the 16-byte value whose hexadecimal
encoding is `ba86f2bbe107c7c57eb5f2690775c712`. An independent known-answer
calculation shows that this equals MD5 over the four ASCII bytes `ubnt`; it does
not prove a generic firmware key-derivation function or establish the key for
other firmware families.

The daemon selects this value explicitly only while validated local state
represents the pending/default wire identity. This includes URL-only advanced
adoption as keyless `provisioning`, which still emits `default=true` and
`state=1`. The low-level Inform client never supplies the public value for a
missing key, and the same bytes are rejected as an individual adopted key.
Corrupt state, adopted state without a valid individual key, and failed state
without a key stop before any HTTP request.

Evidence-aligned UXG-Pro 5.0.16 handling:

| Evidence class | Phase | Request/response shape | Stub behavior |
| --- | --- | --- | --- |
| Dynamic DT-001 | Pre-adoption | Firmware informs use the default key and receive HTTP `404` until an admin clicks `Adopt`. | Keep `default=true`, `state=1`. |
| Dynamic DT-001 | First accepted inform | Controller responds HTTP `200` with `_type: "setparam"` and `mgmt_cfg`. | Parse and persist `authkey`, `cfgversion`, and `use_aes_gcm=true`; enter wire state `0`. |
| Dynamic DT-001 | Key switch | Firmware immediately sends the next inform with the adopted key and AES-GCM. | Send `default=false`, `state=0`. No restart occurred in this captured path. |
| Dynamic DT-001 | Accepted adopted inform | Controller returns `_type: "noop"`. | Advance to `default=false`, `state=2` (`Managed`). Static analysis also permits a successful `setparam`; an authkey field or actual Inform-URL change then returns to `0`. |
| Dynamic DT-001 CAP-05 shape plus product safety | System config | The surviving summary shows another `_type: "setparam"` with top-level `cfgversion`, a present `mgmt_cfg` whose contents were not disclosed, and `system_cfg`. | Ignore top-level `cfgversion` as an adoption-store input and retain only bounded `system_cfg` structure metadata; do not apply host users, firewall, routes, certificates, tokens, or secrets. |
| Static vocabulary plus product safety | Upgrade status | Firmware recognizes `_type: "upgrade"`; this row is not part of DT-001. | Record the response as ignored status only. Do not change the advertised operator version, persisted lifecycle state, services, firmware, or host. |
| Static command evidence plus product policy | WireGuard key request | Firmware recognizes `_type: "cmd"`, `cmd: "wg-genkeypair"`, and a `request_id`; this row is not part of DT-001. | Generate an X25519 key pair in memory and attach a firmware-compatible notification to the next ordinary inform: `inform_as_notif=true`, `notif_reason="cmd"`, and a `notif_payload` containing `event_string`, the correlated `request_id`, and `params.private_key`/`params.public_key`. The current core clears the slot only when an HTTP `200` also reaches its delivery point; the evidence-bound every-HTTP-200 rule remains open. Never persist or log the keys or request ID. |
| Static firmware analysis | Steady state | Firmware accepts `_type: "noop"` with `interval`, `immediate`, block lists, and optionally a string `server_time_in_utc`; this row is not part of DT-001. | For an exact supported tuple, clamp a positive interval from the stricter firmware/operator lower bound through the smaller operator/300-second upper bound. Otherwise ignore the hint and retain the operator base. Reset heartbeat-failure backoff. `immediate=true` schedules one bounded nonzero follow-up without replacing the retained normal interval. Each `noop` replaces the complete process-local block selection. A controller-time string shifts only the time reported by later Stub payloads. |

Block selection is producer-local policy, not a command surface. Every accepted
`noop` replaces the previous include/exclude selection, including when the new
lists are empty. Known producers consult only their exact block identifiers;
UDAPI descriptors remain bounded status metadata and are never executed.

`server_time_in_utc` is accepted as string metadata on `noop` only. The Stub
retains a process-local controller-minus-device offset and applies it to the
reported payload time. It never calls a host time service, sets the system
clock, or changes the host timezone.

The first `mgmt_cfg` observed in the lab contained `capability`,
`selfrun_guest_mode`, `cfgversion`, `led_enabled`, `stun_url`, `mgmt_url`,
`authkey`, `use_aes_gcm=true`, and `report_crash`. It did not include a new
`inform_url`; the firmware continued to use the existing inform URL and only
changed the key/cipher context.

GW-14 records a static parse/install/persist/next-use path for
`mgmt_cfg.inform_url`; no retained dynamic capture exercises an Inform-URL
change. This research observation creates no separate product-test obligation.
The current product contract requires only bounded adoption identity and the
host-safety boundary.

Only the enabling value `use_aes_gcm=true` is evidenced. A controller-provided
`false` does not downgrade an established Stub session; firmware-equivalent
downgrade behavior remains explicitly deferred for missing evidence.

The adopted `system_cfg` shape was a JSON string with top-level `ubntconf` and
`udapi` keys. The `udapi` object contained interfaces, services, system users,
firewall sets/filter chains/settings, static routes, and Radius profiles. This
is provisioning data for a real gateway, not a safe set of host instructions
for `unifi-stubd`.

`wg-genkeypair` is the only controller command with an executable result. It
uses Go's in-process X25519 implementation and performs no shell execution,
file write, service action, or host-network mutation. Missing request IDs,
arguments, parameters, malformed requests, and all other commands are ignored.

Heartbeat policy remains local. Packaged defaults start at 10 seconds and
configure operator bounds of 5 through 60 seconds; those bounds do not enable
controller hints by themselves. An exact LIFE-010 tuple must first supply its
firmware minimum. Failure backoff is capped at 300 seconds and uses 20 percent
symmetric jitter. Transport, response-decode, JSON, dispatcher, and non-200
failures double the effective delay up to the local cap. Only an empty HTTP 200
or an accepted controller-object 200 resets the failure count. If the
configured normal interval exceeds the backoff maximum, it remains the lower
bound so a failure never accelerates inform traffic.

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
| `uplink.speed` | selected current uplink link speed in Mbps; used by the switch Experience column |
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
or absent. The local runtime status records root, management `if_table`,
logical `uplink`, and selected `port_table` views. `source_interface` stays on
the logical/port views because it is local diagnostic metadata.

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
