# Installed Package Files

The package builder stages files below `dist/stage/pkgroot/`. Source files that
represent installed Linux config live under `packaging/linux/`. Source files
that represent installed FreeBSD/OPNsense config live under `packaging/freebsd/`.

## Linux Packaged Files

| Linux path | Repository source | Notes |
| --- | --- | --- |
| `/etc/unifi-stubd/config.yaml` | `packaging/linux/etc/unifi-stubd/config.yaml` | Main service config, packaged as config/noreplace |
| `/etc/unifi-stubd/ssh_host_rsa_key` | generated at first SSH adoption start | Host key for the built-in adoption SSH shim |
| `/var/lib/unifi-stubd/adoption.env` | generated at runtime | Persisted adoption state from controller responses |
| `/var/lib/unifi-stubd/status.json` | generated at runtime | Non-sensitive status snapshot for health checks |
| `/usr/local/bin/unifi-stubd` | built from `cmd/unifi-stubd` | Static Linux binary |
| `/lib/systemd/system/unifi-stubd.service` | `packaging/linux/usr/lib/systemd/system/unifi-stubd.service` | Debian systemd unit |
| `/usr/lib/systemd/system/unifi-stubd.service` | `packaging/linux/usr/lib/systemd/system/unifi-stubd.service` | RPM/Arch systemd unit |
| `/etc/init.d/unifi-stubd` | `packaging/linux/etc/init.d/unifi-stubd` | OpenRC service |
| `/usr/share/doc/unifi-stubd/LICENSE` | `LICENSE` | Project license |
| `/usr/share/doc/unifi-stubd/NOTICE.md` | `NOTICE.md` | Third-party notices |
| `/usr/share/doc/unifi-stubd/CREDITS.md` | `CREDITS.md` | Research and attribution |

## FreeBSD/OPNsense Packaged Files

The FreeBSD package path is currently stub-only and is built as a `.tar.gz`
artifact with `make package-freebsd-tgz`.

| FreeBSD path | Repository source | Notes |
| --- | --- | --- |
| `/usr/local/etc/unifi-stubd/config.yaml` | `packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml` | Main service config |
| `/usr/local/etc/unifi-stubd/ssh_host_rsa_key` | generated at first SSH adoption start | Host key for the built-in adoption SSH shim |
| `/var/db/unifi-stubd/adoption.env` | generated at runtime | Persisted adoption state from controller responses |
| `/var/db/unifi-stubd/status.json` | generated at runtime | Non-sensitive status snapshot for health checks |
| `/usr/local/bin/unifi-stubd` | built from `cmd/unifi-stubd` | Static FreeBSD binary |
| `/usr/local/etc/rc.d/unifi-stubd` | `packaging/freebsd/usr/local/etc/rc.d/unifi-stubd` | rc.d service |
| `/usr/local/share/doc/unifi-stubd/LICENSE` | `LICENSE` | Project license |
| `/usr/local/share/doc/unifi-stubd/NOTICE.md` | `NOTICE.md` | Third-party notices |
| `/usr/local/share/doc/unifi-stubd/CREDITS.md` | `CREDITS.md` | Research and attribution |

## Lab Switch Identities

Lab switch identities live under `lab/`. They are not installed by packages
automatically; copy one over `/etc/unifi-stubd/config.yaml` when it matches the
Linux lab you are building, or over
`/usr/local/etc/unifi-stubd/config.yaml` on FreeBSD.
