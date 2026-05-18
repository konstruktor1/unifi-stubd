# UDM Pro SE Runtime Helpers

This directory contains project-owned runtime files copied into the Docker
firmware wrapper image at `/usr/local/lib/udm-pro-se-runtime/`.

The top-level start scripts stay thin:

- `start-firmware-processes.sh` sources `common/` and `firmware/`.
- `start-webportal-processes.sh` sources `common/` and `webportal/`.

Layout:

- `common/`: helpers shared by the firmware and webportal entry points. It
  currently logs the mounted shared QEMU/UTM kernel deployment manifest.
- `firmware/`: reduced firmware process-chain configuration and process
  helpers for `ubios-udapi-server`, `udapi-bridge`, `mcad`, and local socket
  inspection.
- `webportal/`: webportal configuration, wrapper installation, HTTP patching,
  and service startup helpers.
- `webportal/wrappers/`: real wrapper scripts installed into `/usr/bin`,
  `/sbin`, or `/usr/local/bin` at container startup. They cover controlled
  `systemctl`, `systemd-run`, support-bundle `tar`, `timedatectl`,
  `ubnt-systool`, `ubnt-tools`, and UDAPI command behavior.
- `webportal/http/`: nginx snippets and AWK filters used to patch generated
  UniFi Core HTTP config. The filters keep local preview URLs usable and block
  destructive setup routes.
- `webportal/templates/`: sudoers and UniFi Core config templates rendered at
  startup.
- `webportal/data/`: deterministic lab data used by runtime wrappers, currently
  the stable `ubnt-tools id` payload.

Do not place extracted vendor files here. These files are local lab
compatibility helpers only.
