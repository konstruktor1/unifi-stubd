# UDM Pro SE Source Inventory

This inventory separates project-owned helper code from observed vendor
firmware paths. It is intentionally structural only.

## Project-Owned Helpers

- `Dockerfile`
- `compose.yaml`
- `start-firmware-processes.sh`
- `docker-howto.md`
- `firmware.md`
- `README.md`
- `ubnthal_redirect.c`
- `fixtures/mca-dump-summary.json`

## Observed Vendor Components

Observed inside the extracted rootfs:

- `/usr/bin/mcad`
- `/usr/bin/mca-ctrl -> mcad`
- `/usr/bin/mca-cli -> mcad`
- `/usr/bin/mca-cli-op -> mcad`
- `/usr/bin/mca-monitor -> mcad`
- `/usr/bin/mca.sh`
- `/usr/bin/syswrapper.sh`
- `/usr/bin/ubios-udapi-server`
- `/usr/bin/ubios-udapi-client`
- `/usr/bin/udapi-bridge`
- `/usr/bin/udapic`
- `/usr/lib/libsw.so`
- `/lib/systemd/system/udapi-server.service`
- `/lib/systemd/system/udapi-bridge.service`
- `/lib/systemd/system/mcagent.service`
- `/usr/share/ubios-udapi-server/config-board/udm-pro-se-ea2c.json`
- `/usr/share/ubios-udapi-server/udm-pro-se-ea2c.default`
- `/usr/share/ubios-udapi-server/udm-pro-se-ea2c.fallback`
- `/etc/lagd/configs/udmse-ea2c.json`
- `/lib/modules/4.19.152-ui-alpine/`

These files are not copied into Git. Keep all extracted vendor content under
ignored `artifacts/`, `rootfs/`, or Docker volume storage.

## License Boundary

Use the vendor binaries only as local research inputs. If a future change needs
to copy source code, structured data, scripts, or configuration from external
projects or firmware content, update `CREDITS.md`, `NOTICE.md`, and the license
decision before merging.
