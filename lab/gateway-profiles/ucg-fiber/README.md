# UCG-Fiber Firmware Simulation

This directory contains the project-owned helper files for the UCG-Fiber real
firmware profile. It wraps a locally imported UbiOS ARM64 rootfs and keeps the
vendor firmware image, extracted rootfs data, logs, captures, keys, tokens, and
private controller data out of Git.

Use `docker-howto.md` for the reproducible local setup.

Current state:

- The official UCG-Fiber 5.0.16 image has been verified locally by SHA-256.
- The first SquashFS rootfs has been isolated and inspected.
- The process wrapper starts `ubios-udapi-server`, `udapi-bridge`, and `mcad`
  under Docker with mocked board data.
- `udapi-bridge` opens its local REST listener on `lo:1080`.
- `ubios-udapi-server` does not yet create
  `/var/run/ubnt-udapi-server.sock`, and `mcad` does not yet expose
  `/tmp/.mcad`.

Do not place `internal/device` stub profile data here.
