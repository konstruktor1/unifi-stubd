# UCG-Fiber Firmware 5.0.16 Research

Status: rootfs imported; partial userspace simulation starts through
`lab/gateway-profiles/ucg-fiber/`, but `mcad` cannot yet expose its local
control socket in the current Docker wrapper.

This profile tracks the public UCG-Fiber platform firmware used for local
UniFi Cloud Gateway Fiber research. Keep the downloaded image, extracted
SquashFS, rootfs, logs, captures, keys, tokens, certificates, and private
controller data under ignored local paths only.

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

## Service Chain

The firmware uses the same broad UbiOS management stack as the other ARM64
gateway profiles:

- `/usr/bin/ubios-udapi-server`
- `/usr/bin/udapi-bridge`
- `/usr/bin/mcad`

Board data is present under:

- `/usr/share/ubios-udapi-server/config-board/ucg-fiber-a6a8.json`

Safe board summary:

- Board ID: `a6a8`
- Board architecture: `ipq9574`
- Product ID: `ucg-fiber`
- Model short: `UCGF`
- Model full: `UniFi Cloud Gateway Fiber`
- WAN port mapping observed in board config: `wan0` on `eth4`, `wan1` on
  `eth6`

## Simulation Wrapper

The project-owned wrapper lives under:

```text
lab/gateway-profiles/ucg-fiber/
```

Use `lab/gateway-profiles/ucg-fiber/docker-howto.md` to import the rootfs into
a local Docker image and start the isolated simulation.

Current blocker: `ubios-udapi-server` creates the bridge event notifier socket
but not `/var/run/ubnt-udapi-server.sock` in the current containerized run.
Because that socket is missing, `mcad` does not expose `/tmp/.mcad`, so
`mca-ctrl -t dump` cannot complete yet.

## Source Availability

No vendor source or decompiled code is copied into this repository. The image
contains proprietary UniFi management binaries and open-source userspace
components; any corresponding source bundle needs to be handled as a separate
license request or vendor-provided archive.
