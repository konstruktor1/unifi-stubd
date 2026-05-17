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

Snapshot window:

```text
2026-05-17T13:53:44Z through 2026-05-17T14:00:26Z
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
| Device MAC in packet header | `00:00:00:00:00:00` |
| Request body size range | `1312` to `1763` bytes |
| TNBU payload size range | `1272` to `1723` bytes |

Observed TNBU flag distribution in the snapshot:

| Flags | Count |
| --- | ---: |
| `3` | `113` |
| `11` | `230` |

The controller returned empty responses with HTTP `404` during this run. That
means the firmware, MITM, and controller network path is working, but the UniFi
Network Application had not yet reached a state where it handles this lab
inform stream as an adopted or adoptable device.

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
