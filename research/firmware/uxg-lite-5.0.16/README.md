# UXG-Lite Firmware 5.0.16 Research

Status: rootfs imported; partial userspace simulation starts, but `mcad`
cannot yet expose its local control socket in the current Docker/QEMU wrapper.

This profile tracks the public UXG platform firmware used for UniFi Gateway
Lite research. Keep the downloaded image, extracted SquashFS, rootfs, logs, and
raw captures under the ignored `artifacts/` directory or Docker volumes.

## Firmware Image

- Product: UniFi Gateway Lite
- Firmware API platform: `UXG`
- Firmware version: `5.0.16.30689`
- Image version string: `UXG.ipq5018.v5.0.16.9d45777.260226.1635`
- Architecture: Linux ARM64 userspace
- Download URL:
  `https://fw-download.ubnt.com/data/unifi-firmware/dad0-UXG-5.0.16-996c83e4-42a4-4dc7-bfa3-26894dc59cd7.bin`
- SHA-256:
  `e2e361fc9b4296628f1b4fa10280449695c9df8fb8da95a440a7462044a4765c`
- File size: `494840050` bytes
- Container base image after local import: `uxg-lite-fw:5.0.16`

The image begins with the expected Ubiquiti header:
`UBNTUXG.ipq5018.v5.0.16.9d45777.260226.1635`.

## Rootfs

The first SquashFS header starts at byte offset `15644578`.

Observed SquashFS metadata:

- SquashFS 4.0
- Compression: `lz4`
- Block size: `262144`
- Created: `2026-02-26 09:47:59 UTC`
- Inodes: `21631`
- OS release inside rootfs: Debian GNU/Linux 11 (`bullseye`)

macOS extraction into a normal local directory can hit xattr warnings and
case-insensitive filename collisions. Use a Linux filesystem, a Docker volume,
or a case-sensitive disk image for the rootfs extraction.

## Service Chain

The firmware uses the same broad UbiOS management stack as UXG-Pro:

- `ubios-udapi-server.service`
  - binary: `/usr/bin/ubios-udapi-server`
  - packaged service command uses the persisted state file under
    `/data/udapi-config/ubios-udapi-server/`
- `udapi-bridge.service`
  - binary: `/usr/bin/udapi-bridge`
  - provides the local REST bridge used by management components
- `mcagent.service`
  - binary: `/usr/bin/mcad`
  - starts after `udapi-bridge.service` and `dbus.service`

Board data is present under:

- `/usr/share/ubios-udapi-server/config-board/uxglite-a677.json`
- `/usr/share/ubios-udapi-server/uxg-lite-a677.default`
- `/usr/share/ubios-udapi-server/uxg-lite-a677.fallback`

Safe board summary:

- Board ID: `a677`
- Board architecture: `ipq5018`
- Product ID: `uxg-lite`
- Model short: `UXG Lite`
- Model full: `Gateway Lite`
- WAN port mapping observed in board config: `wan0` on `eth1`

## Current Simulation Result

The committed wrapper in `simulation/` can start the ARM64 rootfs under Docker
with mocked `/proc/ubnthal` data.

Observed in the local run:

- `ubios-udapi-server` starts and reads the mocked board file.
- It creates `/run/ubios-udapi-server-bridge-event-notifier.sock`.
- It does not create `/var/run/ubnt-udapi-server.sock` in the current
  containerized run.
- `udapi-bridge` starts and opens its local REST listener on `lo:1080`.
- `mcad` starts, but does not create `/tmp/.mcad`, so `mca-ctrl -t dump`
  cannot yet retrieve a management dump.

That means the current UXG-Lite simulation is useful for firmware structure,
process, and startup-path analysis, but it is not yet a complete controller
adoption lab like `uxgpro-5.0.16`.

## Working Hypothesis

The blocker is before the normal UbiOS UDAPI server socket comes up. The next
debug step should run the same container with ARM64 `strace` or a broader
LD_PRELOAD shim to find the missing kernel, cgroup, systemd, or hardware path
that keeps `ubios-udapi-server` spinning before it binds
`/var/run/ubnt-udapi-server.sock`.

## Source Availability

No matching public source tree for the proprietary UniFi management binaries
was found during online research. The image contains open-source userspace
components, but any corresponding GPL source bundle needs to be handled as a
separate license request or vendor-provided archive. Do not copy vendor source
or decompiled code into this repository.
