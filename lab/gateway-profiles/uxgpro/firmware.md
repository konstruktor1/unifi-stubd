# UXG-Pro Firmware 5.0.16 Research

This folder records the local analysis of the UniFi Gateway UXG-Pro firmware
image used to understand discovery, inform, adoption, and gateway-agent
behavior for `unifi-stubd`.

## Input

- Product: UniFi Gateway UXG-Pro
- Firmware image: `UXGPROV2-5.0.16.bin`
- Firmware header: `UBNTUXGPRO.al324.v5.0.16.9d45777.260226.1635`
- SHA-256: `18a7f198f71edc0161365114356239b0b370b4b90f664bb90f253b33f8b5658c`
- Local image path during analysis: `/tmp/unifi-stubd-fw/UXGPROV2-5.0.16.bin`
- Extracted rootfs mount during analysis: `/Volumes/unifi-stubd-uxgfw/rootfs`
- Local file manifest: `/tmp/unifi-stubd-fw/uxgpro-rootfs-files.txt`

The firmware image and extracted rootfs are intentionally not copied into this
repository. They are vendor artifacts and are covered by the ignore rules for
`lab/gateway-profiles/**/artifacts/`.

## Archive Layout

Observed HIT archive parts:

| Part | Offset |
| --- | ---: |
| U-Boot | `268` |
| Kernel | `1325736` |
| Rootfs | `15807154` |
| SquashFS start | `15807210` |
| Updater | `346989298` |

Rootfs identification:

- Debian `11.11`
- Architecture: `aarch64`
- `/usr/lib/version`: `UXGPRO.al324.v5.0.16.9d45777.260226.1635`

## Service Chain

The relevant startup chain from systemd is:

```text
udapi-server.service
  ExecStart=/usr/bin/ubios-udapi-server -c /data/udapi-config/ubios-udapi-server/ubios-udapi-server.state --silent

udapi-bridge.service
  After=udapi-server.service
  ExecStart=/usr/bin/udapi-bridge --watchdog --rest-api-port 1080 --rest-api-secure-port 0 --rest-api-interface lo

mcagent.service
  After=udapi-bridge.service dbus.service
  ExecStart=/usr/bin/mcad
```

Runtime model:

```text
UniFi Network Controller
  <-> mcad
      <-> /tmp/.mcad local datagram control socket
      -> udapi-bridge REST on localhost:1080
          -> ubios-udapi-server Unix socket /var/run/ubnt-udapi-server.sock
```

## Key Findings

- `mcad` is the management and inform agent.
- `mca-ctrl`, `mca-cli`, and `mca-cli-op` are symlinks to `/usr/bin/mcad`;
  behavior changes by `argv[0]`.
- `mca-ctrl -t dump` talks to `mcad` over `/tmp/.mcad` and returns the current
  inform/status JSON.
- `ubios-udapi-server` is the local gateway configuration engine.
- `udapi-bridge` exposes a localhost REST API in front of UDAPI.
- `syswrapper.sh set-adopt <inform_url> <authkey>` ultimately calls
  `mca-ctrl -t connect -s <inform_url> -k <authkey>`.
- Controller command strings in `mcad` include provisioning and sensitive
  operations such as firmware update, restart, restore-default, shell command,
  shadow-mode adoption, and packet capture.

For `unifi-stubd`, these findings support the current safety boundary: emulate
discovery, inform, and adoption state, but do not execute controller-triggered
shell, upgrade, restart, or host-networking mutations.

## Simulation Snapshot

Imported Docker image:

```text
uxgpro-fw:5.0.16
sha256:2ffeaf29e59cac944e0deba39c4a5c5a0d9f6460902fc6f8c738a510d8a7ad03
```

Container used for isolated runtime checks:

```text
uxgpro-fw-fullsim
--network none
```

Processes verified in the isolated simulation:

```text
/usr/bin/ubios-udapi-server -c /data/udapi-config/ubios-udapi-server/ubios-udapi-server.state -x -t
/usr/bin/udapi-bridge -m UXGPRO -M 02:15:6d:de:ad:00 --rest-api-port 1080 --rest-api-secure-port 0 --rest-api-interface lo -l - -x -
/usr/bin/mcad -n -s -v
```

Useful inspection commands:

```sh
docker exec uxgpro-fw-fullsim /usr/bin/mca-ctrl -t dump
docker exec uxgpro-fw-fullsim /usr/bin/ubios-udapi-client -r GET /device
docker exec uxgpro-fw-fullsim /usr/bin/ubios-udapi-client -r GET /interfaces
docker exec uxgpro-fw-fullsim /usr/bin/ubios-udapi-client -r GET /system/configuration
```

`controller-lab.compose.yaml` adds a local UniFi Network
Application and MongoDB so the firmware wrapper can talk to
`http://unifi:8080/inform` inside a private Docker network. See
`controller-lab.md`.

On 2026-05-17, after a lab admin logged into the UniFi Network web portal and
clicked `Adopt`, the simulated UXG-Pro firmware completed adoption:

- `mca-ctrl -t dump` reported `default=false`, `state=2`, and
  `last_error=null`.
- The controller database stored MAC `02:15:6d:de:ad:00` as model `UXGPRO`,
  type `uxg`, with `adopted=true`.
- The controller sent `setparam` messages for management configuration and
  gateway system configuration, then settled into `noop` responses with a
  10-second inform interval.

The committed adoption fixture is sanitized and omits the adopted inform key,
tokens, certificates, password hashes, and raw `system_cfg`:

```text
fixtures/adoption-mitm-timeline.json
fixtures/adopted-system-config-summary.json
```

## Source Availability

No complete public source tree for `mcad`, `udapi-bridge`, or
`ubios-udapi-server` was found during the 2026-05-17 online and firmware
inventory pass. The binaries in the image are stripped and proprietary.
Ubiquiti's public GitHub organization currently does not publish the UXG-Pro
firmware agent sources.

GPL-covered components may require a source request to Ubiquiti support. That
does not provide the proprietary UniFi agent source.

References checked:

- https://github.com/ubiquiti
- https://help.ui.com/hc/en-us/articles/204910064-UniFi-Advanced-Updating-Techniques
- https://community.ui.com/questions/Source-Code-UniFi/967271b5-bbd0-4c4a-84af-1c01ddb95a8c
