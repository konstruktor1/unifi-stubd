# UCG-Fiber Source Inventory

This inventory separates project-owned helper code from observed vendor
firmware paths. It is intentionally structural only.

## Project-Owned Helpers

The project-owned simulation files are kept under:

```text
lab/gateway-profiles/ucg-fiber/
```

That directory contains the Docker wrapper, compose file, process startup
script, LD_PRELOAD shim, how-to, and firmware inventory notes.

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
- `/usr/share/ubios-udapi-server/config-board/ucg-fiber-a6a8.json`
- `/lib/modules/5.4.213-ui-ipq9574/`

These files are not copied into Git. Keep all extracted vendor content under
ignored `artifacts/`, `rootfs/`, or Docker volume storage.

## License Boundary

Use the vendor binaries only as local research inputs. If a future change needs
to copy source code, structured data, scripts, or configuration from external
projects or firmware content, update `CREDITS.md`, `NOTICE.md`, and the license
decision before merging.
