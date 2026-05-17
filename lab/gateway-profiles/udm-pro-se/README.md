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
- Simulation remains partial until local `mca-ctrl` access and controller
  adoption are proven.

Use `docker-howto.md` for the local import and startup steps.
