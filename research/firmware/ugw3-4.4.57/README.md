# UGW3 Firmware 4.4.57 Research

Status: rootfs imported; legacy `mcad` runs through a Docker/QEMU-MIPS chroot
runner and exposes `/tmp/.mcad` for `mca-ctrl` inspection.

This profile tracks the public UniFi Security Gateway 3P firmware. Keep the
downloaded tarball, extracted kernel, SquashFS, rootfs, logs, and raw captures
under ignored `artifacts/` paths or Docker volumes.

## Firmware Image

- Product: UniFi Security Gateway 3P
- Firmware API platform: `UGW3`
- Firmware version: `4.4.57.5578372`
- Image version string:
  `UniFiSecurityGateway.ER-e120.v4.4.57.5578372.230112.0823`
- Kernel image architecture: MIPS64 big-endian
- Userspace architecture: MIPS 32-bit big-endian
- Download URL:
  `https://fw-download.ubnt.com/data/unifi-firmware/7920-UGW3-4.4.57-803dc5671c6745dbb68c8dfa10145a8f.tar`
- SHA-256:
  `08a35a626e9733018b2e49af92aa3474255136bb0178e0697f44c6d8042cdd74`
- File size: `109199360` bytes
- Extracted rootfs Docker volume used locally: `unifi-ugw3-rootfs`

The firmware tar contains:

- `vmlinux.tmp`
- `vmlinux.tmp.md5`
- `squashfs.tmp`
- `squashfs.tmp.md5`
- `version.tmp`
- `compat`

Observed package metadata:

- `version.tmp`: `v4.4.57.5578372.230112.0823`
- `compat`: `20004:5`
- `squashfs.tmp.md5`: `873e35196d6338b8c5538ca37eb109af`
- `vmlinux.tmp.md5`: `10f64178501d8aa399fab034751e7e80`

## Rootfs

The rootfs is `squashfs.tmp` from the firmware tar.

Observed SquashFS metadata:

- SquashFS 4.0
- Compression: `gzip`
- Inodes: `33876`
- Debian version inside rootfs: `7.11`

The MIPS userspace cannot be run directly as a normal Docker image on this
host. The committed simulation uses a Debian runner container with
`qemu-mips-static` and chroots into the extracted UGW3 rootfs.

## Service Chain

This is a legacy USG/EdgeOS style firmware, not the UbiOS stack used by
UXG-Pro and UXG-Lite.

Observed startup chain:

- `/etc/init.d/unifi-init`
  - starts `/usr/bin/mcad`
  - starts `/usr/bin/mca-monitor`
  - starts `/usr/bin/linkcheck`
  - prepares device fingerprint state from platform storage
- `/usr/bin/mcad`
  - controller-facing management agent
  - owns `/tmp/.mcad` for local CLI control
- `/usr/bin/mca-ctrl`
  - local management CLI
- `/usr/bin/syswrapper.sh`
  - invoked by `mcad` for platform actions
- `/opt/vyatta/`
  - EdgeOS/Vyatta configuration and operational templates

## Current Simulation Result

The QEMU-MIPS chroot runner starts `mcad` and exposes the local control socket.

Verified locally:

- `mcad` starts in default state.
- It attempts default inform to `http://unifi:8080/inform`.
- `mca-ctrl -t dump` works through the same chroot and returns a JSON dump.
- The dump currently has placeholder hardware identity fields because the
  runner does not yet emulate USG platform EEPROM, board ID, or interface
  state.

Safe observed dump fields:

- `architecture`: `mips`
- `default`: `true`
- `inform_url`: `http://unifi:8080/inform`
- `state`: `1`
- `version`: `4.4.57.5578372`
- `selfrun_beacon`: `true`

The next step is to add a small legacy hardware mock layer for board identity,
MAC address, serial, and interface table so the controller sees a stable UGW3
identity instead of placeholders.

## Source Availability

No matching public source tree for the proprietary USG management binaries was
found during online research. The firmware contains GPL/open-source components,
but any corresponding source bundle needs to be handled as a separate vendor
archive or license request. Do not copy vendor source or decompiled code into
this repository.
