# UGW3 Source Inventory

This inventory separates project-owned helper code from observed vendor
firmware paths. It is intentionally structural only.

## Project-Owned Helpers

- `simulation/Dockerfile`
- `simulation/compose.yaml`
- `simulation/start-ugw3-qemu.sh`
- `simulation/docker-howto.md`
- `simulation/README.md`

These helpers run a Debian/QEMU-MIPS container around a locally extracted
vendor rootfs stored outside Git.

## Observed Vendor Components

Observed inside the extracted UGW3 rootfs:

- `/etc/init.d/unifi-init`
- `/usr/bin/mcad`
- `/usr/bin/mca-ctrl`
- `/usr/bin/mca-cli`
- `/usr/bin/mca-cli-op`
- `/usr/bin/mca-monitor`
- `/usr/bin/linkcheck`
- `/usr/bin/redirector`
- `/usr/bin/syswrapper.sh`
- `/usr/bin/ubnt-upgrade`
- `/opt/vyatta/`

These files are not copied into Git. Keep all extracted vendor content under
ignored `artifacts/`, `rootfs/`, or Docker volume storage.

## License Boundary

Use the vendor binaries only as local research inputs. If a future change needs
to copy source code, structured data, scripts, or configuration from external
projects or firmware content, update `CREDITS.md`, `NOTICE.md`, and the license
decision before merging.
