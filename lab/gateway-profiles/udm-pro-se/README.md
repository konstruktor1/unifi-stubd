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
- `ubios-udapi-server` reaches `/var/run/ubnt-udapi-server.sock` through the
  deterministic RTL8370-style `swconfig` mock.
- `mca-ctrl -t dump` returns local management data through `/tmp/.mcad`.
- Optional `webportal.compose.yaml` exposes the firmware-generated UniFi OS
  setup UI on `https://127.0.0.1:9443/` and an HTTP preview on
  `http://localhost:9080/`. The HTTP preview is useful for the Codex in-app
  browser because it cannot accept the lab's self-signed HTTPS certificate.
  The stack starts the minimal `unifi-core`, `ulp-go`, nginx, PostgreSQL, and
  DBus support services needed for that UI.
- The webportal stack can generate a local support archive through
  `/api/setup/support/generate`; the archive contains lab-generated system
  metadata plus `unifi-core` logs, not a full UniFi OS support dump.
- The setup flow completes through a local Network API facade on
  `127.0.0.1:8081`; the facade publishes a valid UniFi Core app manifest,
  registers the packaged Network UI assets, logs requests, and returns no-op
  responses for the Network endpoints `unifi-core` needs during setup.
- The webportal stack installs a lab-only `systemctl` shim so UniFi Core can
  enable/start the Network service in the non-systemd container without leaving
  the application in an access-prohibited update-failed state.
- The webportal stack maps Docker `eth0` into a lab UDAPI WAN, returns static
  DNS servers, and provides deterministic ISP metadata so UniFi Core can mark
  the console as internet-connected when the container has outbound network
  access.
- Simulation remains partial until deterministic lab netdevs and controller
  adoption are proven.

Use `docker-howto.md` for the local import and startup steps.
