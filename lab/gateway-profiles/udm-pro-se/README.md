# UDM Pro SE Firmware Simulation Profile

This directory contains the project-owned wrapper files for the UDM Pro SE
firmware simulation profile.

The vendor firmware image, extracted SquashFS, imported Docker rootfs image,
logs, captures, keys, tokens, certificates, and private lab data stay outside
Git under ignored local paths.

Current state:

- Firmware image downloaded under `research/firmware/udm-pro-se-5.0.16/`.
- Rootfs SquashFS offset and checksum identified.
- Docker wrapper and networkless Compose simulation are prepared.
- Project-owned mock sources are split under `mock/`: deterministic mock files
  live in `mock/files/`, and the C `LD_PRELOAD` shim is split into modules
  under `mock/ldpreload/`.
- The Docker profile mounts the shared kernel deployment payload from the
  QEMU/UTM profile at `/opt/unifi-fw-sim/kernel` when
  `../udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh` has staged it.
- `ubios-udapi-server` reaches `/var/run/ubnt-udapi-server.sock` through the
  deterministic RTL8370-style `swconfig` mock.
- `mca-ctrl -t dump` returns local management data through `/tmp/.mcad`.
- Optional `webportal.compose.yaml` exposes the firmware-generated UniFi OS
  setup UI on `https://127.0.0.1:9443/` and an HTTP preview on
  `http://localhost:9080/`. The HTTP preview is useful for the Codex in-app
  browser because it cannot accept the lab's self-signed HTTPS certificate.
  The stack starts the minimal `unifi-core`, `ulp-go`, nginx, PostgreSQL, and
  DBus support services needed for that UI. The lab-only systemd DBus facade
  is split under `systemd-dbus/`.
- The webportal stack can generate a local support archive through
  `/api/setup/support/generate`; the archive contains lab-generated system
  metadata plus `unifi-core` logs, not a full UniFi OS support dump.
- The setup flow completes through a local Network API facade on
  `127.0.0.1:8081`; the facade publishes a valid UniFi Core app manifest,
  registers the packaged Network UI assets, logs requests, and returns no-op
  responses for the Network endpoints `unifi-core` needs during setup. Its
  CommonJS sources live under `network-app/`, split by config, payloads, routes,
  websocket handling, and HTTP helpers.
- The webportal stack installs a lab-only `systemctl` shim so UniFi Core can
  enable/start the Network service in the non-systemd container without leaving
  the application in an access-prohibited update-failed state.
- The webportal stack maps Docker `eth0` into a lab UDAPI WAN, returns static
  DNS servers, and provides deterministic ISP metadata so UniFi Core can mark
  the console as internet-connected when the container has outbound network
  access.
- The latest Docker full test confirmed the networkless firmware path, the
  webportal path, `/api/system` internet readiness after settling, Network
  facade health, and local support-bundle generation.
- Simulation remains partial. The Docker webportal path is useful for setup UI
  and UniFi Core behavior inspection; the QEMU/UTM path is the reference for
  native firmware boot behavior.

## Source Layout

Project-owned sources are intentionally grouped by runtime boundary:

- `mock/files/`: deterministic `/mock` filesystem inputs shared by Docker and
  QEMU/UTM. The current committed data is under `mock/files/ubnthal/`, with
  synthetic `board` and `system.info` identity values.
- `mock/ldpreload/`: modular C `LD_PRELOAD` shim for filesystem redirects,
  response patching, root-user compatibility, socket tracing, process
  containment, and the RTL8370-style `swconfig` ABI. The modules are split into
  path mapping, open/ioctl interception, process containment, response
  patching, socket tracing, auth compatibility, and switch ABI files.
- `fixtures/`: sanitized committed output, currently the reduced
  `mca-dump-summary.json` reference.
- `network-app/`: minimal CommonJS UniFi Network API facade used only by the
  Docker webportal path. It is split into configuration, logging, HTTP helpers,
  deterministic payloads, routes, websocket handling, and the process entry
  point.
- `systemd-dbus/`: minimal CommonJS `org.freedesktop.systemd1` DBus facade used
  only by the Docker webportal path. It is split into configuration, DBus
  binding, interface definitions, unit fixtures, server wiring, and the process
  entry point.
- `udapi-lab-shim.cjs`: read-only UDAPI and `mca-dump` wrapper that maps Docker
  `eth0` into a deterministic UDM-style WAN view.
- `runtime/`: sourced shell modules, wrapper scripts, nginx snippets, AWK
  filters, templates, and deterministic data used by the two start scripts. The
  `runtime/common/`, `runtime/firmware/`, and `runtime/webportal/` subtrees are
  described in `runtime/README.md`.
- `start-firmware-processes.sh`: thin orchestrator for the reduced firmware
  process chain.
- `start-webportal-processes.sh`: thin orchestrator for the firmware process
  chain plus minimal UniFi OS webportal support services.

Use `docker-howto.md` for the local import and startup steps.

The QEMU/UTM VM reference that reuses these mocks lives in
`../udm-pro-se-vm/`. The current cross-profile status is summarized in
`../../../docs/en/project-status.md` and `../../../docs/de/project-status.md`.
