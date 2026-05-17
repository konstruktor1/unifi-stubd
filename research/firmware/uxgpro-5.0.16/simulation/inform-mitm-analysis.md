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

The controller still returned empty responses with HTTP `404` for the real
UXG-Pro firmware `POST /inform` stream. Direct invalid requests to `/inform`
returned `400`, so the controller HTTP path exists; the current Dockerized
UniFi Network Application still does not accept this simulated real-gateway
inform stream as an adopted or adoptable device.

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

`mcad` repeatedly tries to log in to the local `udapi-bridge` as `root`, but
the bridge reports `RESTAPI login failed for user root`. Direct UDAPI probing
showed `/user/check` returning an `A12` error while `/system/users` still lists
`root`. Because of that bridge authentication failure, UDAPI-derived blocks are
not present in the inform payload yet.

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
