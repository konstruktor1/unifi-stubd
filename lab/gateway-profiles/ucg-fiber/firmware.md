# UCG-Fiber Firmware 5.0.16 Research

Status: rootfs imported; partial userspace simulation starts, but `mcad`
cannot yet expose its local control socket in the current Docker wrapper.

This profile tracks the public UCG-Fiber platform firmware used for local
UniFi Cloud Gateway Fiber research. Keep the downloaded image, extracted
SquashFS, rootfs, logs, and raw captures under ignored `artifacts/` directories
or Docker volumes.

## Firmware Image

- Product: UniFi Cloud Gateway Fiber
- Firmware API platform: `UCGF`
- Firmware version: `5.0.16`
- Image version string: `UCGF.ipq9574.v5.0.16.238fde6.260227.0038`
- Architecture: Linux ARM64 userspace
- Release source: official Ubiquiti Community release `r/ucgf/5.0.16`
- Image name:
  `ca3a-UCGF-5.0.16-109206e5-e71e-4be4-b9c8-a4d4ae2ac799.bin`
- SHA-256:
  `7c8635974513413f19b4542c85b188d93a2fef38707ad53d237ce6a657e88ce5`
- File size: `814850522` bytes
- Container base image after local import: `ucg-fiber-fw:5.0.16`

The image begins with the expected Ubiquiti header:
`UBNTUCGF.ipq9574.v5.0.16.238fde6.260227.0038`.

## Rootfs

The first SquashFS header starts at byte offset `12914826`.

Observed SquashFS metadata:

- SquashFS 4.0
- Compression: `zstd`
- Block size: `262144`
- Filesystem size: `797345789` bytes
- Rootfs slice SHA-256:
  `ee244235939905984f6897c9e5e91d0c951e1fdc3a0677bc3f4f426d97a44f21`
- Created: `2026-02-26 16:46:03 UTC`
- Inodes: `44854`
- OS release inside rootfs: Debian GNU/Linux 11 (`bullseye`)

macOS extraction into a normal local directory can hit xattr warnings and
case-insensitive filename collisions. Use a Linux filesystem, a Docker volume,
or a case-sensitive disk image for the rootfs extraction.

## Service Chain

The firmware uses the same broad UbiOS management stack as the other ARM64
gateway profiles:

- `ubios-udapi-server.service`
  - binary: `/usr/bin/ubios-udapi-server`
  - packaged service command uses the persisted state file under
    `/data/udapi-config/ubios-udapi-server/`
- `udapi-bridge.service`
  - binary: `/usr/bin/udapi-bridge`
  - local REST bridge configured on `lo:1080`
- `mcagent.service`
  - binary: `/usr/bin/mcad`
  - starts the controller-facing management agent

Board data is present under:

- `/usr/share/ubios-udapi-server/config-board/ucg-fiber-a6a8.json`

Safe board summary:

- Board ID: `a6a8`
- Board architecture: `ipq9574`
- Product ID: `ucg-fiber`
- Model short: `UCGF`
- Model full: `UniFi Cloud Gateway Fiber`
- Interfaces observed in board config:
  - `eth0` through `eth3`: `2.5GE`
  - `eth4`: `10GE`
  - `eth5` and `eth6`: `SFP-plus`
- WAN port mapping observed in board config:
  - `wan0` on `eth4`
  - `wan1` on `eth6`
- Switch driver in board config: `RTL8372`

## Current Simulation Result

The committed wrapper in this directory starts the ARM64 rootfs under Docker
with mocked `/proc/ubnthal` and selected `/proc/sys` data.

- `ubios-udapi-server` starts and reads mocked board data.
- It creates `/run/ubios-udapi-server-bridge-event-notifier.sock`.
- It does not create `/var/run/ubnt-udapi-server.sock` in the current
  containerized run.
- `udapi-bridge` starts with model `UCGF` and opens its local REST listener on
  `lo:1080`.
- `mcad` starts, but does not create `/tmp/.mcad`, so `mca-ctrl -t dump`
  cannot yet retrieve a management dump.

This profile is not yet a complete controller adoption lab. Keep it networkless
until `mca-ctrl -t dump` works with deterministic mock hardware data.

## Working Hypothesis

UCG-Fiber should follow the same early UbiOS process shape as UXG-Lite, but the
larger IPQ9574 platform may need additional mocked thermal, GPIO, SFP, sysctl,
or kernel-module paths before `ubios-udapi-server` reaches its normal socket
bind. The next debug step is to run the wrapper with ARM64 `strace` or broader
LD_PRELOAD logging and add only deterministic lab values.

## Source Availability

No vendor source or decompiled code is copied into this repository. The image
contains proprietary UniFi management binaries and open-source userspace
components; any corresponding source bundle needs to be handled as a separate
license request or vendor-provided archive.
