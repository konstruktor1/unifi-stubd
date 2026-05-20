# Inform MITM Analysis

This note records the safe, non-sensitive findings from the local controller
lab MITM run on 2026-05-17. Raw capture files are intentionally not committed.
They can contain controller state, adoption material, keys, or device data once
the lab moves past initial setup.

## Lab Path

```text
UXG-Pro firmware mcad
  -> http://unifi:8080/inform
  -> inform-mitm at 172.31.240.12
  -> http://unifi-controller:8080/inform
  -> UniFi Network Application at 172.31.240.10
```

`unifi` is injected into the firmware container through `/etc/hosts` and points
to the mitmproxy container. mitmproxy runs in reverse mode and forwards the
request to the real controller service name, `unifi-controller`.

## Observed Inform Traffic

Snapshot source:

```text
research/firmware/uxgpro-5.0.16/simulation/captures/events.jsonl
```

Committed sanitized telegram samples:

```text
research/firmware/uxgpro-5.0.16/simulation/fixtures/inform-telegrams.jsonl
```

Committed decoded gateway payload sample:

```text
research/firmware/uxgpro-5.0.16/simulation/fixtures/decoded-gateway-inform-sample.json
```

Committed sanitized adoption timeline:

```text
research/firmware/uxgpro-5.0.16/simulation/fixtures/adoption-mitm-timeline.json
```

Snapshot window:

```text
2026-05-17T14:48:00Z through 2026-05-17T14:53:05Z
```

The useful device-originated traffic had this shape:

| Field | Observed value |
| --- | --- |
| Method | `POST` |
| URL | `http://unifi:8080/inform` |
| Host header | `unifi:8080` |
| User-Agent | `AirControl Agent v1.0` |
| Content-Type | `application/x-binary` |
| Packet magic | `TNBU` |
| Packet version | `0` |
| Payload version | `1` |
| Device MAC in packet header | `00:15:6d:de:ad:00` |
| Device MAC in decoded payload | `00:15:6d:de:ad:00` |
| Device serial in decoded payload | `00156DDEAD00` |
| Request body size range | `1329` to `1930` bytes |
| TNBU payload size range | `1289` to `1890` bytes |

Observed TNBU flag distribution in the snapshot:

| Flags | Count |
| --- | ---: |
| `3` | `120` |
| `11` | `179` |

Earlier simulation runs reported `00:00:00:00:00:00` as the top-level device
MAC. That was caused by a zero serial in the mocked `/proc/ubnthal` board and
system metadata. Setting both `serial` and `serialno` to `00156DDEAD00` made
`mcad` emit a stable top-level MAC and serial in both the TNBU packet header
and decoded payload.

Before the portal adoption action, the controller returned empty responses with
HTTP `404` for the real UXG-Pro firmware `POST /inform` stream. Direct invalid
requests to `/inform` returned `400`, so the controller HTTP path existed, but
the simulated gateway was not accepted until an admin explicitly clicked
`Adopt` in the UniFi Network web portal.

## Portal Adoption Sequence

After web-portal login as the lab admin and an explicit `Adopt` click, the
next visible inform exchange changed from HTTP `404` to HTTP `200` and the
controller stored the device as adopted.

Adoption result:

| Field | Observed value |
| --- | --- |
| Device MAC | `00:15:6d:de:ad:00` |
| Device serial | `00156DDEAD00` |
| Controller device type | `uxg` |
| Controller adopted flag | `true` |
| Firmware `default` | `false` |
| Firmware `state` | `2` |
| Firmware `last_error` | `null` |
| Firmware `cfgversion` | `87893ca41993f905` |
| Steady-state inform interval | `10` seconds |

High-level transition:

| Time | Event ID | Direction | Decoded type | Result |
| --- | --- | --- | --- | --- |
| `2026-05-17T17:01:20Z` | `1779037280245-fb26ae02` | firmware -> controller | normal inform | HTTP `404` before portal adoption |
| `2026-05-17T17:01:29Z` | `1779037289438-6b67a31d` | controller -> firmware | `setparam` | First HTTP `200`; `mgmt_cfg` delivered with new inform auth key redacted |
| `2026-05-17T17:01:29Z` | `1779037289595-fa52b98b` | firmware -> controller | inform with adopted key | Firmware switched to AES-GCM using the new key |
| `2026-05-17T17:01:30Z` | `1779037290647-73097ae3` | controller -> firmware | `setparam` | Controller delivered larger gateway `system_cfg`; sensitive content redacted |
| `2026-05-17T17:01:42Z` | `1779037302714-2133134a` | controller -> firmware | `noop` | Steady-state poll interval set to `10` seconds |

The first adoption `setparam` carried `mgmt_cfg` fields including
`cfgversion`, `stun_url`, `mgmt_url`, `use_aes_gcm=true`, and an `authkey`.
The `authkey` is the point where the inform cipher context changes; it is not
committed. Later request and response bodies require that adopted inform key
for decoding.

The larger `system_cfg` response contains controller tokens, certificates,
password hashes, Radius secrets, and network/firewall configuration. It is
useful for local reverse engineering, but it is treated as sensitive lab data
and only summarized in committed fixtures.

The safe structural summary is committed in:

```text
research/firmware/uxgpro-5.0.16/simulation/fixtures/adopted-system-config-summary.json
```

Its `system_cfg` shape has top-level `ubntconf` and `udapi` keys. The `udapi`
object contains five interfaces (`lo`, `eth0`, `eth1`, `eth2`, `eth3`), ten
service categories, one root user entry with password hash present, 18 firewall
sets, eight filter chains, no NAT rules, no static routes, and one Radius
profile with a secret present.

The decoded payload makes the gateway reporting shape visible. Gateway
interfaces are reported through `if_table` and `network_table`; `port_table`
remains `null` and was not observed in this gateway inform stream. The lab
currently reports four interfaces:

| Interface | Address | MAC | Source table |
| --- | --- | --- | --- |
| `eth0` | `172.31.240.20/24` | `00:15:6d:de:ad:00` | `if_table`, `network_table` |
| `eth1` | `192.0.2.2/24` | `00:15:6d:de:ad:01` | `if_table`, `network_table` |
| `eth2` | `198.51.100.2/24` | `00:15:6d:de:ad:02` | `network_table` |
| `eth3` | `203.0.113.2/24` | `00:15:6d:de:ad:03` | `network_table` |

Before adoption, `mcad` repeatedly tried to log in to the local `udapi-bridge`
as `root`, and the bridge reported `RESTAPI login failed for user root`.
Direct UDAPI probing showed `/user/check` returning an `A12` error while
`/system/users` still listed `root`. That explained missing UDAPI-derived
blocks in the pre-adoption payload. After adoption, the inform stream included
additional gateway blocks such as `ipv4_active_leases` and `gw_caps`, but the
raw adopted payloads remain local because they require the adopted inform key.

## Local Raw Files

The MITM writes local runtime artifacts below:

```text
research/firmware/uxgpro-5.0.16/simulation/captures/
```

The directory is ignored by Git. Typical files are:

- `inform.flows`: mitmproxy flow archive.
- `events.jsonl`: one JSON line per request or response.
- `*-request.bin`: raw HTTP request body.
- `*-response.bin`: raw HTTP response body.

Do not commit those files. For shareable analysis, regenerate a sanitized
summary from `events.jsonl` and omit raw bodies, body hashes, adoption tokens,
controller URLs, private addresses, and real device identifiers.
