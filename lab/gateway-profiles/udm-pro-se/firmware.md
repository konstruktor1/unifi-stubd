# UDM Pro SE Firmware 5.0.16 Research

Status: rootfs identified and simulation wrapper prepared. The networkless
Docker wrapper reaches the UbiOS UDAPI server socket and `mca-ctrl -t dump`
with a deterministic RTL8370-style switch mock. The optional Docker webportal
override exposes a partial UniFi OS setup surface through local facades for
Network, systemd DBus, and UDAPI metadata. A full controller-adopted firmware
profile has not been completed for this path.

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

## Project-Owned Lab Surfaces

The repository-owned code around this firmware is split by runtime boundary:

- `mock/ldpreload/`: C interposition modules shared by Docker and QEMU/UTM.
- `mock/files/`: deterministic board, system, MTD, sysfs, and persistent mock
  inputs staged into `/mock`.
- `network-app/`: CommonJS Network API facade that returns deterministic setup
  payloads and app-readiness metadata for `unifi-core`.
- `systemd-dbus/`: CommonJS `org.freedesktop.systemd1` facade that exposes only
  the manager/unit/service properties UniFi Core reads in this lab.
- `udapi-lab-shim.cjs`: read-only wrapper for `mca-ctrl`,
  `ubios-udapi-client`, and `mca-dump` when Docker needs a UDM-style WAN view.
- `runtime/firmware/`: reduced firmware process-chain configuration and process
  helpers for UDAPI and switch-mock inspection.
- `runtime/webportal/`: webportal wrapper installation, HTTP patching, service
  startup helpers, nginx snippets, AWK filters, and templates.
- `start-firmware-processes.sh` and `start-webportal-processes.sh`: thin
  orchestrators that source the runtime modules.

## Current Networkless Result

The committed wrapper is prepared to run the ARM64 rootfs under Docker with
mocked `/proc/ubnthal` and selected `/proc/sys` data.

The mock directory is mounted read-write because early UDM Pro SE startup writes
sysctl values such as `/proc/sys/net/core/rmem_max`; the LD_PRELOAD shim
redirects those writes into `/mock/proc/sys/...`.

The RTL8370 is not fully emulated as an ASIC. Instead, the modular
`LD_PRELOAD` shim under `mock/ldpreload/` exports the small
`libsw.so`/OpenWrt `swconfig` ABI surface that `ubios-udapi-server` uses and
returns deterministic lab values for switch, VLAN, port, PoE, mirror,
isolation, link, and MIB lookups. Configuration writes are accepted for
analysis but do not mutate host networking. Static mock inputs such as
`ubnthal/board` and `ubnthal/system.info` live under `mock/files/`.

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

## Current Webportal Result

`webportal.compose.yaml` starts the same firmware process chain plus the
minimum UniFi OS userspace needed by the setup UI:

- PostgreSQL 14 for `unifi-core` and `ulp-go`.
- `dbus-daemon` and the local `systemd-dbus/` facade for service watchers.
- A lab `systemctl` wrapper that accepts known UniFi application lifecycle
  calls without running systemd as PID 1.
- A lab `systemd-run` wrapper for sudoers-approved support-bundle commands.
- The `network-app/` facade on `127.0.0.1:8081`, publishing the Network app
  manifest, app status, setup endpoints, device summary, and websocket
  handshake.
- UDAPI wrappers that expose Docker `eth0` as a plugged WAN, static DNS
  metadata, and clearly labeled lab ISP metadata.
- nginx and UniFi Core routes for setup, support-bundle generation, and a
  localhost-only HTTP preview.

The webportal path can be reached on host `https://127.0.0.1:9443/` in Docker.
The latest Docker full test returned the UniFi OS setup HTML with
`UNIFI_OS_MANIFEST`, started `unifi-core`, `ulp-go`, nginx, PostgreSQL, the
systemd DBus facade, the Network facade, UDAPI, `udapi-bridge`, and `mcad`, and
generated a local support archive through `/api/setup/support/generate`.
`/api/system` can report `hasInternet=false` during early readiness, then
reports `hasInternet=true` once the lab UDAPI/Core state settles.

In the UTM VM profile the comparable guest `443` path is the native firmware
boot reference. Direct guest HTTPS through UTM Shared/NAT worked in the latest
test. The UTM profile writes an intended localhost mapping for
`https://127.0.0.1:10443/`, but that host port was not proven to be bound by
UTM itself; an explicit local TCP helper can make the URL work for browser
convenience. The Docker and QEMU/UTM paths intentionally stay separate: Docker
is a setup/API inspection wrapper, while QEMU/UTM is the native boot reference.

Docker is therefore a weaker test for "does the UDM firmware boot correctly?"
It skips the UDM kernel question entirely, runs under the host kernel, and uses
lab facades for systemd, Network, and selected UDAPI/Core expectations. It is
valuable for learning which API and UI surfaces the firmware expects, but the
UTM VM is the stronger reference for systemd boot order, VM networking, and the
console's own view of WAN/LAN readiness.

This profile is not yet a complete controller adoption lab. Keep external
controller attachment isolated until a sanitized controller/MITM run confirms
the adopted inform path.

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
