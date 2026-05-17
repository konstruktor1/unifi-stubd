# UXG-Lite Source Inventory

This inventory separates project-owned helper code from observed vendor
firmware paths. It is intentionally structural only.

## Project-Owned Helpers

- `Dockerfile`
- `compose.yaml`
- `start-firmware-processes.sh`
- `docker-howto.md`
- `README.md`

The Dockerfile reuses the project-owned LD_PRELOAD shim from the UXG-Pro
research folder at build time:

- `lab/gateway-profiles/uxg-lite/ubnthal_redirect.c`

## Observed Vendor Components

Observed inside the extracted rootfs:

- `/usr/bin/mcad`
- `/usr/bin/mca-ctrl`
- `/usr/bin/mca-cli`
- `/usr/bin/mca-cli-op`
- `/usr/bin/syswrapper.sh`
- `/usr/bin/ubios-udapi-server`
- `/usr/bin/udapi-bridge`
- `/lib/systemd/system/udapi-server.service`
- `/lib/systemd/system/udapi-bridge.service`
- `/lib/systemd/system/mcagent.service`
- `/usr/share/ubios-udapi-server/config-board/uxglite-a677.json`
- `/usr/share/ubios-udapi-server/uxg-lite-a677.default`
- `/usr/share/ubios-udapi-server/uxg-lite-a677.fallback`

These files are not copied into Git. Keep all extracted vendor content under
ignored `artifacts/`, `rootfs/`, or Docker volume storage.

## License Boundary

Use the vendor binaries only as local research inputs. If a future change needs
to copy source code, structured data, scripts, or configuration from external
projects or firmware content, update `CREDITS.md`, `NOTICE.md`, and the license
decision before merging.
