# UDM Pro SE Firmware 5.0.16 Research

Status: rootfs identified and simulation wrapper prepared. The local wrapper
now reaches the UbiOS UDAPI server socket and `mca-ctrl -t dump` with a
deterministic RTL8370-style switch mock. Controller adoption has not been
completed for this profile yet.

This profile tracks the official UniFi Dream Machine Special Edition firmware
image for local research. Keep the downloaded image, extracted SquashFS,
rootfs, logs, and raw captures under ignored `artifacts/` directories or Docker
volumes.

## Firmware Image

- Product: UniFi Dream Machine Special Edition
- Firmware API platform: `UDMPROSE`
- Firmware version: `5.0.16`
- Image version string: `UDMPROSE.al324.v5.0.16.238fde6.260227.0037`
- Architecture: Linux ARM64 userspace
- Release source: official Ubiquiti Community release `r/udm/5.0.16`
- Image name:
  `473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin`
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

macOS extraction into a normal local directory can hit xattr warnings and
case-insensitive filename collisions. Use a Linux filesystem, a Docker volume,
or a case-sensitive disk image for rootfs extraction.

## Service Chain

The firmware uses the UbiOS management stack seen in the other ARM64 gateway
profiles:

- `ubios-udapi-server.service`
  - binary: `/usr/bin/ubios-udapi-server`
  - service command uses state under `/data/udapi-config/ubios-udapi-server/`
- `udapi-bridge.service`
  - binary: `/usr/bin/udapi-bridge`
  - local REST bridge configured on `lo:1080`
- `mcagent.service`
  - binary: `/usr/bin/mcad`
  - helper symlinks: `/usr/bin/mca-ctrl`, `/usr/bin/mca-cli`,
    `/usr/bin/mca-cli-op`, and `/usr/bin/mca-monitor`
  - controller-facing management agent

Board data is present under:

- `/usr/share/ubios-udapi-server/config-board/udm-pro-se-ea2c.json`

Safe board summary:

- Board ID: `ea2c`
- Board architecture: `alpinev2`
- Product ID: `udm-pro-se`
- Family: `UniFi Dream Machine`
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

## Current Simulation Result

The committed wrapper is prepared to run the ARM64 rootfs under Docker with
mocked `/proc/ubnthal` and selected `/proc/sys` data.

The mock directory is mounted read-write because early UDM Pro SE startup writes
sysctl values such as `/proc/sys/net/core/rmem_max`; the LD_PRELOAD shim
redirects those writes into `/mock/proc/sys/...`.

The RTL8370 is not fully emulated as an ASIC. Instead, the LD_PRELOAD shim
exports the small `libsw.so`/OpenWrt `swconfig` ABI surface that
`ubios-udapi-server` uses and returns deterministic lab values for switch,
VLAN, port, PoE, mirror, isolation, link, and MIB lookups. Configuration writes
are accepted for analysis but do not mutate host networking.

Observed startup path:

- `ubios-udapi-server` starts and accepts the mocked Ubiquiti board identity.
- It writes redirected sysctl values into the mock tree.
- It opens the mocked MTD EEPROM paths and accepts the mocked `MTD_OTPSELECT`
  ioctl used during hardware initialization.
- It configures the simulated switch through `swconfig dev switch0 ...`
  operations, including VLAN, port, PoE, isolation, and mirror attributes.
- It creates `/var/run/ubnt-udapi-server.sock` and logs
  `Listening on UNIX socket /var/run/ubnt-udapi-server.sock`.
- `udapi-bridge` can connect to that socket and sends internal UDAPI requests
  such as `/vpn/wireguard/servers` and `/qos/fw/queues`.
- `mcad` creates `/tmp/.mcad`, and `mca-ctrl -t dump` returns a usable local
  management dump with the mocked identity.
- A sanitized summary is committed at
  `lab/gateway-profiles/udm-pro-se/fixtures/mca-dump-summary.json`.
- `if_table` and `network_table` are still empty because the networkless
  wrapper does not yet provide deterministic `switch0` and `eth0` through
  `eth10` netdevs.

When `UNIFI_FW_SIM_ALLOW_PARTIAL=1`, the startup wrapper keeps the container
alive after a firmware process exits so logs remain inspectable.

This profile is not yet a complete controller adoption lab. Keep it networkless
until a sanitized controller/MITM run confirms the adopted inform path.

## Working Hypothesis

UDM Pro SE follows the same broad UbiOS process shape as UCG-Fiber and UXG-Pro,
but the Alpine AL324 platform needs the switch-driver layer to exist before
`ubios-udapi-server` reaches normal socket bind. The current lab approach is to
mock the userspace `swconfig` interface instead of emulating RTL8370 registers.
The next debug step is to add deterministic Linux netdev and netlink behavior
for `switch0` and `eth0` through `eth10`, then attach the profile to a
controller/MITM lab.

## Source Availability

No vendor source or decompiled code is copied into this repository. The image
contains proprietary UniFi management binaries and open-source userspace
components; any corresponding source bundle needs to be handled as a separate
license request or vendor-provided archive.

Online check on 2026-05-17 found the official Ubiquiti Community firmware
release for Dream Machines 5.0.16, but no matching public UDM Pro SE 5.0.16
source bundle. Treat GPL or other open-source component source as
vendor-request material unless Ubiquiti publishes a matching archive.
