# UDM Pro SE Firmware 5.0.16 Research

Status: rootfs identified and lab simulation wrapper prepared. The local
wrapper reaches the UbiOS UDAPI server socket and `mca-ctrl -t dump` with a
deterministic RTL8370-style switch mock. No adopted controller run has been
completed for this profile yet.

This profile tracks the official UniFi Dream Machine Special Edition firmware
image for local research. Keep the downloaded image, extracted filesystems,
logs, captures, keys, tokens, certificates, and private controller data under
ignored local paths only.

## Firmware Image

- Product: UniFi Dream Machine Special Edition
- Firmware API platform: `UDMPROSE`
- Firmware version: `5.0.16`
- Image version string: `UDMPROSE.al324.v5.0.16.238fde6.260227.0037`
- Architecture: Linux ARM64 userspace
- Release source: official Ubiquiti Community release `r/udm/5.0.16`
- Download URL:
  `https://fw-download.ubnt.com/data/unifi-dream/473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin`
- Local artifact:
  `research/firmware/udm-pro-se-5.0.16/artifacts/473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin`
- SHA-256:
  `7cf58f4563522220716f5025a7b2954b070df6be9364d7f60af0bc644512bce4`
- File size: `964689062` bytes
- Container base image after local import: `udm-pro-se-fw:5.0.16`

The image begins with the expected Ubiquiti header:
`UBNTUDMPROSE.al324.v5.0.16.238fde6.260227.0037`.

## Rootfs

The first SquashFS header starts at byte offset `16141142`.

Observed SquashFS metadata:

- SquashFS 4.0
- Compression: `zstd`
- Block size: `262144`
- Filesystem size: `943959416` bytes
- Rootfs slice SHA-256:
  `a5f710613cfd0b5f9ccb49380d36c76a9d21143db796a43d6c70e9ef1a84e17d`
- Created: `2026-02-26 18:07:07 UTC`
- Inodes: `45010`
- OS release inside rootfs: Debian GNU/Linux 11 (`bullseye`)

## Service Chain

Observed management path:

- `ubios-udapi-server.service`
  - binary: `/usr/bin/ubios-udapi-server`
- `udapi-bridge.service`
  - binary: `/usr/bin/udapi-bridge`
  - local REST bridge configured on `lo:1080`
- `mcagent.service`
  - binary: `/usr/bin/mcad`
  - helper symlinks: `/usr/bin/mca-ctrl`, `/usr/bin/mca-cli`,
    `/usr/bin/mca-cli-op`, and `/usr/bin/mca-monitor`

Board data is present under:

- `/usr/share/ubios-udapi-server/config-board/udm-pro-se-ea2c.json`

Safe board summary:

- Board ID: `ea2c`
- Board architecture: `alpinev2`
- Product ID: `udm-pro-se`
- Model short: `UDM-SE`
- Model full: `UniFi Dream Machine SE`
- Interfaces observed in board config:
  - `eth0` through `eth7`: switch-backed `GE` ports
  - `eth8`: `2.5GE`
  - `eth9` and `eth10`: `SFP-plus`
- WAN port mapping observed in board config:
  - `wan0` on `eth8`
  - `wan1` on `eth9`
- Switch driver in board config: `RTL8370`
- PoE entries are present for `eth0` through `eth7`.

## Current State

- The firmware image and extracted `rootfs.squashfs` are stored locally under
  ignored `artifacts/`.
- The project-owned simulation wrapper is under
  `lab/gateway-profiles/udm-pro-se/`.
- The wrapper redirects selected board, sysctl, MTD, sysfs, and persistent
  paths into `/mock`.
- The wrapper does not fully emulate the RTL8370 ASIC. It provides a
  deterministic userspace `libsw.so`/OpenWrt `swconfig` ABI mock so the firmware
  can configure a lab-local switch model without mutating host networking.
- A startup run reaches `/var/run/ubnt-udapi-server.sock`; `udapi-bridge`
  connects and exchanges internal UDAPI requests.
- `mcad` creates `/tmp/.mcad`, and `mca-ctrl -t dump` returns a usable local
  management dump with the mocked identity.
- A sanitized dump summary is committed at
  `lab/gateway-profiles/udm-pro-se/fixtures/mca-dump-summary.json`.
- `if_table` and `network_table` are empty until the lab provides
  deterministic `switch0` and `eth0` through `eth10` netdevs.
- No UDM Pro SE controller adoption lab has been completed yet.

## Source Availability

No vendor source or decompiled code is copied into this repository. Online
check on 2026-05-17 found the official Ubiquiti Community firmware release for
Dream Machines 5.0.16, but no matching public UDM Pro SE 5.0.16 source bundle.
Treat GPL or other open-source component source as vendor-request material
unless Ubiquiti publishes a matching archive.

## Next Research Steps

- Add deterministic Linux netdev and netlink behavior for `switch0` and `eth0`
  through `eth10`.
- Attach this profile to a controller/MITM lab and sanitize the adoption
  findings before committing them.
