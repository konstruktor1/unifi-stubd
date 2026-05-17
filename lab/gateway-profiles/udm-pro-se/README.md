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
- Simulation is expected to be partial until local `mca-ctrl` access is proven.

Use `docker-howto.md` for the local import and startup steps.
