# UDM Pro SE Docker Simulation How-To

Run commands from the repository root.

## Requirements

- Docker with Linux/ARM64 support.
- `squashfs-tools`.
- The official UDM Pro SE firmware image downloaded separately into ignored
  `research/firmware/udm-pro-se-5.0.16/artifacts/`.

## Paths

```sh
PROFILE=lab/gateway-profiles/udm-pro-se
VM_PROFILE=lab/gateway-profiles/udm-pro-se-vm
RESEARCH=research/firmware/udm-pro-se-5.0.16
ARTIFACTS="$RESEARCH/artifacts"
FW="$ARTIFACTS/473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin"
SIM=/tmp/unifi-fw-sim-udm-pro-se
mkdir -p "$ARTIFACTS"
```

## Project-Owned Source Map

The committed files in this profile are helpers around a locally imported
firmware rootfs. They do not include extracted vendor source or proprietary
runtime data.

| Path | Role |
| --- | --- |
| `Dockerfile` | Wraps the imported ARM64 firmware rootfs and copies lab helpers. |
| `compose.yaml` | Starts the networkless firmware process path and optional shim builder. |
| `webportal.compose.yaml` | Adds localhost-only HTTP/HTTPS exposure for the UniFi OS setup UI. |
| `start-firmware-processes.sh` | Thin orchestrator for `runtime/firmware/`. |
| `start-webportal-processes.sh` | Thin orchestrator for `runtime/webportal/`. |
| `runtime/` | Sourced startup modules, wrapper scripts, nginx snippets, AWK filters, templates, and deterministic data. |
| `mock/files/` | Static mock filesystem inputs staged into `/mock`. |
| `mock/ldpreload/` | C shim modules shared with the QEMU/UTM VM profile. |
| `network-app/` | CommonJS Network API facade for setup endpoints and app readiness. |
| `systemd-dbus/` | CommonJS systemd DBus facade for UniFi Core service watchers. |
| `udapi-lab-shim.cjs` | Read-only UDAPI wrapper for Docker WAN/DNS/ISP metadata. |

Verify the image:

```sh
shasum -a 256 "$FW"
```

Expected hash:

```text
7cf58f4563522220716f5025a7b2954b070df6be9364d7f60af0bc644512bce4
```

## Extract Rootfs

The first SquashFS header starts at byte offset `16141142`.

```sh
SQUASHFS_OFFSET=16141142
tail -c +$((SQUASHFS_OFFSET + 1)) "$FW" > "$ARTIFACTS/rootfs.squashfs"
```

Import through a Linux filesystem or Docker volume to avoid macOS
case-insensitive path collisions:

```sh
docker volume create unifi-udm-pro-se-rootfs

docker run --rm \
  -v "$PWD/$ARTIFACTS:/firmware:ro" \
  -v unifi-udm-pro-se-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-udm-pro-se-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - udm-pro-se-fw:5.0.16
```

## Prepare Mock Hardware Files

```sh
rm -rf "$SIM"
mkdir -p \
  "$SIM/mtd" \
  "$SIM/persistent" \
  "$SIM/sys/class/hwmon/hwmon0/device" \
  "$SIM/sys/class/mtd/mtd5" \
  "$SIM/sys/class/thermal/thermal_zone0" \
  "$SIM/ubnthal/status" \
  "$SIM/proc/sys/crypto" \
  "$SIM/proc/sys/kernel" \
  "$SIM/proc/sys/net/core" \
  "$SIM/proc/sys/net/ipv4" \
  "$SIM/proc/sys/net/ipv6/conf/all" \
  "$SIM/proc/sys/net/netfilter"

cp -R "$PROFILE/mock/ldpreload" "$SIM/ldpreload"
cp -R "$PROFILE/mock/files/." "$SIM/"
```

The UDM Pro SE startup path writes sysctl values during initialization, so this
profile mounts the mock directory read-write. Static mock inputs such as
`ubnthal/board` and `ubnthal/system.info` live under `mock/files/` so Docker
and QEMU/UTM use the same deterministic identity data.

Create the simple mock values:

```sh
printf 'false\n' > "$SIM/ubnthal/status/IsLocated"
printf '0\n' > "$SIM/proc/sys/crypto/fips_enabled"
printf 'UDM-Pro-SE\n' > "$SIM/proc/sys/kernel/hostname"
printf '(none)\n' > "$SIM/proc/sys/kernel/domainname"
printf '212992\n' > "$SIM/proc/sys/net/core/rmem_max"
printf '212992\n' > "$SIM/proc/sys/net/core/wmem_max"
printf '4096\n' > "$SIM/proc/sys/net/core/somaxconn"
printf '0\n' > "$SIM/proc/sys/net/ipv4/ip_forward"
printf '0\n' > "$SIM/proc/sys/net/ipv6/conf/all/forwarding"
printf '0\n' > "$SIM/proc/sys/net/netfilter/nf_conntrack_helper"
printf 'dev:    size   erasesize  name\nmtd5: 00010000 00010000 "eeprom"\n' > "$SIM/mtd/proc_mtd"
dd if=/dev/zero of="$SIM/mtd/mtd5" bs=65536 count=1
cp "$SIM/mtd/mtd5" "$SIM/mtd/mtdblock5"
printf 'c2 20 18\n' > "$SIM/sys/class/mtd/mtd5/jedec_id"
printf '50000\n' > "$SIM/sys/class/hwmon/hwmon0/device/temp1_input"
printf '42000\n' > "$SIM/sys/class/hwmon/hwmon0/device/temp2_input"
printf '43000\n' > "$SIM/sys/class/hwmon/hwmon0/device/temp3_input"
printf '1800\n' > "$SIM/sys/class/hwmon/hwmon0/device/fan1_input"
printf '1600\n' > "$SIM/sys/class/hwmon/hwmon0/device/fan2_input"
printf '50000\n' > "$SIM/sys/class/thermal/thermal_zone0/temp"
```

## Build Shim

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  --profile build-shim \
  run --rm shim-builder
```

## Deploy Kernel Payload

Docker does not boot a private kernel, but the lab mounts the same local kernel
payload used by the QEMU/UTM reference at `/opt/unifi-fw-sim/kernel`. Build the
VM artifacts first, then stage the shared deployment directory:

```sh
"$VM_PROFILE/scripts/prepare-vm.sh"
"$VM_PROFILE/scripts/fetch-foreign-kernel.sh"
"$VM_PROFILE/scripts/prepare-mocks.sh"
"$VM_PROFILE/scripts/build-lab-initramfs.sh"
"$VM_PROFILE/scripts/deploy-kernel-artifacts.sh"
```

The generated directory stays ignored:

```text
lab/gateway-profiles/udm-pro-se-vm/artifacts/deploy/kernel/
```

It contains the extracted vendor kernel, the foreign QEMU-virt-capable kernel,
the matching foreign module tree, and the lab initramfs. The Docker startup
scripts write the mounted payload manifest to `kernel-artifacts.txt` in their
log directory.

## Start Simulation

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  up -d --build firmware
```

For syscall and shim tracing, rebuild the shim and start with tracing enabled:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  --profile build-shim \
  run --rm shim-builder

SIM_DIR="$SIM" \
UNIFI_FW_SIM_TRACE=1 \
UBNTHAL_REDIRECT_DEBUG=1 \
UBNTHAL_REDIRECT_TRACE_ALL=1 \
docker compose \
  -f "$PROFILE/compose.yaml" \
  up -d --build firmware
```

Inspect:

```sh
docker compose -f "$PROFILE/compose.yaml" ps
docker compose -f "$PROFILE/compose.yaml" logs --tail 120 firmware
```

## Start With Local Webportal

Use the webportal override when the lab should expose the UDM Pro SE setup
surface on the host. The override maps host `127.0.0.1:9443` to container
port `443`, keeps a lab-only HTTP preview on `127.0.0.1:9080`, and starts the
minimal UniFi OS userspace needed by the setup UI:

- PostgreSQL 14 for `unifi-core` and `ulp-go`.
- `dbus-daemon` plus the lab-only `org.freedesktop.systemd1` facade for the
  `uos-agent.service` watcher. The facade source lives under `systemd-dbus/`,
  split into DBus binding, unit fixture, interface, and server modules.
- A lab `systemd-run` shim at `/usr/bin/systemd-run` so `unifi-core` can run
  its sudoers-approved transient support-bundle commands in a non-systemd
  container.
- A lab `systemctl` shim at `/usr/bin/systemctl` for known UniFi application
  lifecycle calls. `systemctl enable --now unifi` starts the Network facade and
  returns success, which keeps the webportal from marking Network as an
  update-failed, access-prohibited application.
- A minimal UniFi Network API facade on `127.0.0.1:8081`. UniFi Core calls this
  during setup for Network feature checks, previous-subnet detection,
  `network_optimization`, country settings, and the `set-installed` command.
  The facade publishes a valid `apps[]` manifest, registers the packaged
  Network UI assets from `/usr/lib/unifi/webapps/ROOT/app-unifi`, logs
  requests, and returns deterministic no-op responses. The source is split
  under `network-app/` so endpoint fixtures, routing, logging, and websocket
  behavior can be extended independently.
- Lab UDAPI wrappers for `mca-ctrl`, `ubios-udapi-client`, and `mca-dump`.
  These wrappers map Docker `eth0` to an explicit WAN, return static DNS
  servers, and expose deterministic ISP metadata. That lets UniFi Core run its
  normal physical-WAN and ping checks inside the container without adding host
  networking privileges.
- `ulp-go-app` on `127.0.0.1:9080`.
- `unifi-core` and its generated nginx configuration.
- The firmware management processes from the normal simulation path.

The webportal startup also installs lab-only wrappers for `ubnt-tools` and
`ubnt-systool`. These wrappers return deterministic board and anonymous-ID
values and turn host-mutating commands such as reboot, poweroff, reset, network,
timezone, hostname, SSH, and firmware-update changes into logged no-ops. The
`ubnt-systool support` wrapper writes deterministic lab metadata, and the
startup script pre-creates empty application support paths because `unifi-core`
packages every configured app log directory even when the app is not installed.
The UniFi Core override sets `overrideConsoleFeatures.waitForUFN: false`
because the real Java UniFi Network backend is not started by this minimal
firmware webportal; the lab facade exposes only the Core-facing API surface
needed for setup and inspection.

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  -f "$PROFILE/webportal.compose.yaml" \
  up -d --build firmware
```

Open:

```text
https://127.0.0.1:9443/
```

The certificate is generated by the firmware stack and is not trusted by the
host browser. Use the HTTP preview when a local automation browser cannot
accept that certificate:

```text
http://127.0.0.1:9080/
```

For command-line checks, use `curl -k` for HTTPS:

```sh
curl -k -I https://127.0.0.1:9443/
curl -k -sS https://127.0.0.1:9443/ | grep UNIFI_OS_MANIFEST
curl -k -sS https://127.0.0.1:9443/api/system | jq '.hasInternet'
```

Support-bundle check:

```sh
curl -k -sS \
  -o /tmp/udm-pro-se-support.tgz \
  https://127.0.0.1:9443/api/setup/support/generate

tar -tzf /tmp/udm-pro-se-support.tgz | head
```

The current full-test result for the Docker path is:

- The shim builder produces an AArch64 `libubnthal_redirect.so` from
  `mock/ldpreload/`.
- The networkless firmware service starts `ubios-udapi-server`,
  `udapi-bridge`, and `mcad`; `/var/run/ubnt-udapi-server.sock` is present.
- The mounted shared kernel deployment manifest is written to
  `/tmp/kernel-artifacts.txt` inside the container log path.
- The webportal service returns HTTP 200 on `https://127.0.0.1:9443/` and the
  response contains `UNIFI_OS_MANIFEST`.
- `unifi-core`, `ulp-go`, nginx, PostgreSQL, the systemd DBus facade, the
  Network facade, UDAPI, `udapi-bridge`, and `mcad` are all expected in the
  process/socket checks below.
- `/api/system` can briefly report `hasInternet=false` before Core and UDAPI
  readiness settle. After the lab UDAPI state is ready it reports
  `hasInternet=true`; the device setup state depends on the current local setup
  data.
- `/proxy/network/api/s/default/stat/health` returns the deterministic Docker
  facade health payload.
- `/api/setup/support/generate` produces a local archive containing lab system
  metadata such as `system/lab-system.txt`, the deterministic
  `ubnt-tools` identity payload, and `unifi-core` logs.

Useful process and socket checks:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  -f "$PROFILE/webportal.compose.yaml" \
  exec firmware \
  sh -lc 'pgrep -af "unifi-core|ulp-go|nginx|ubios-udapi-server|udapi-bridge|mcad"; ss -ltnp'
```

The webportal lab is still partial UniFi OS, not a full console boot. In
particular, storage gRPC services such as `127.0.0.1:11052` are not started,
so `unifi-core` logs expected storage connection errors while the setup UI and
management sockets remain available.

This is why the Docker full test is not as strong as the UTM full test for
firmware-boot questions. Docker proves that the extracted userspace, local
facades, nginx/Core setup path, and support-bundle wrappers can cooperate under
the host kernel. It does not prove the UDM kernel, initramfs handoff, firmware
`systemd` ordering, VM NIC enumeration, or the native Network self-view. Use the
UTM profile for those questions because it boots an ARM64 VM through the UDM
initramfs/rootfs path and exposes the two UDM-like NIC roles used by the current
web test.

Inside the container, inspect sockets and logs before attaching this profile to
any controller lab:

```sh
docker compose -f "$PROFILE/compose.yaml" exec firmware \
  find /run /tmp -maxdepth 2 -type s -o -type f

docker compose -f "$PROFILE/compose.yaml" exec firmware \
  sh -lc 'test -S /var/run/ubnt-udapi-server.sock && echo UDAPI_SOCKET_PRESENT'

docker compose -f "$PROFILE/compose.yaml" exec firmware \
  tail -80 /tmp/ubios-udapi-server.run.log

docker compose -f "$PROFILE/compose.yaml" exec firmware \
  timeout 20 /usr/bin/mca-ctrl -t dump
```

Stop:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  down
```

## Current Limitation

The networkless wrapper starts far enough to prove that the firmware reads the
mocked board identity, writes redirected sysctl values, initializes mocked
MTD/sysfs paths, configures the RTL8370-style switch through the userspace
`swconfig` ABI, creates `/var/run/ubnt-udapi-server.sock`, and returns a local
`mca-ctrl -t dump` through `/tmp/.mcad`.

The webportal override starts a partial UniFi OS setup surface. Its local
facades report Network as installed and ready, map Docker `eth0` to a lab WAN,
and block host-mutating setup actions. It is useful for inspecting setup and
frontend/API expectations, but it is still not a full UniFi OS console.

The RTL8370 is not emulated at register or kernel-driver level. The shim only
mocks the `libsw.so` API surface used by the firmware. Treat this as a
startup-analysis and setup-inspection profile. The QEMU/UTM VM path remains the
reference for native firmware boot behavior.
