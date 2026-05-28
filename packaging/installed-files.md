# Installed Package Files

The package builder stages files below `dist/stage/pkgroot/`. Source files that
represent installed Linux config live under `packaging/linux/`. Source files
that represent installed FreeBSD/OPNsense config live under `packaging/freebsd/`.
These packaged configs are neutral defaults. Keep real host-specific configs in
a private location outside this repository, then copy the selected config to the
installed service path after package installation.

## Linux Packaged Files

Linux packages create a dedicated `unifi-stubd` service user for the systemd
unit. Packaged defaults keep the adoption SSH shim closed. The unit still
grants `CAP_NET_BIND_SERVICE` so isolated labs can explicitly bind a
UniFi-compatible low port without running the daemon as root.

| Linux path | Repository source | Notes |
| --- | --- | --- |
| `/etc/unifi-stubd/config.yaml` | `packaging/linux/etc/unifi-stubd/config.yaml` | Main service config, packaged as config/noreplace |
| `/var/lib/unifi-stubd/ssh_host_rsa_key` | generated at first SSH adoption start | Host key for the optional adoption SSH shim |
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
artifact with `make package-freebsd-tgz`. Release builds publish both
`freebsd_amd64` and `freebsd_arm64` tarballs as GitHub Release assets and under
the GitHub Pages paths `/freebsd/amd64/` and `/freebsd/arm64/`.

| FreeBSD path | Repository source | Notes |
| --- | --- | --- |
| `/usr/local/etc/unifi-stubd/config.yaml` | `packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml` | Main service config |
| `/usr/local/etc/unifi-stubd/ssh_host_rsa_key` | generated at first SSH adoption start | Host key for the optional adoption SSH shim |
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

Real controller URLs, site IP addresses, MAC addresses, client names, adoption
state paths, and private topology mappings must not be committed. Use sanitized
examples in `lab/` for shareable documentation and keep live host configs in a
private deployment directory.
