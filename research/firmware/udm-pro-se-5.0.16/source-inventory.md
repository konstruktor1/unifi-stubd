# UDM Pro SE Source Inventory

This inventory separates project-owned notes from vendor firmware artifacts.
It is intentionally structural only.

## Project-Owned Files

The committed project-owned files for this profile are:

```text
research/firmware/udm-pro-se-5.0.16/README.md
research/firmware/udm-pro-se-5.0.16/source-inventory.md
lab/gateway-profiles/udm-pro-se/
```

## Local Vendor Artifacts

The downloaded firmware image is stored under:

```text
research/firmware/udm-pro-se-5.0.16/artifacts/
```

That path is ignored by Git. Do not commit firmware images, extracted rootfs
trees, raw captures, keys, controller tokens, certificates, SSH host keys, or
private lab data.

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

## License Boundary

Use the vendor binaries only as local research inputs. If a future change needs
to copy source code, structured data, scripts, or configuration from external
projects, update `CREDITS.md`, `NOTICE.md`, and the license decision before
merging.
