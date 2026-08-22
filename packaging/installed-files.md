# Installed Package Files

Installed configuration sources live under `packaging/linux/` and
`packaging/freebsd/`.
These packaged configs explicitly select the small `US8` starter profile but
remain non-observing until a local source is configured. Keep real host-specific
configs in a private location outside this repository, then copy the selected
config to the installed service path after package installation.

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

The FreeBSD package path is currently stub-only. `make package-freebsd-tgz`
builds a runner-produced diagnostic tarball for one `PKG_FREEBSD_GOARCH`.
Release builds are intended to use `make package-freebsd-pkg-repos`, which
writes native FreeBSD `pkg` repositories and both `freebsd_amd64` and
`freebsd_arm64` tarballs. Its current direct FreeBSD host-Go compilation is an
external-machine compatibility path, not Build Runner release evidence;
publication remains
blocked until an authorized build route replaces that step. GitHub Release
assets remain a manual release-checklist step.

| FreeBSD path | Repository source | Notes |
| --- | --- | --- |
| `/usr/local/etc/unifi-stubd/config.yaml` | `packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml` | Main service config; native `pkg` marks it as config so upgrades preserve local edits and write `.pkgnew` on conflicts; the `post-install` hook runs after install and upgrade, calls `unifi-stubd -config-migrate`, and keeps `.bak.*` backups for safe legacy alias normalization |
| `/usr/local/etc/unifi-stubd/ssh_host_rsa_key` | generated at first SSH adoption start | Host key for the optional adoption SSH shim |
| `/var/db/unifi-stubd/adoption.env` | generated at runtime | Persisted adoption state from controller responses |
| `/var/db/unifi-stubd/status.json` | generated at runtime | Non-sensitive status snapshot for health checks |
| `/usr/local/bin/unifi-stubd` | built from `cmd/unifi-stubd` | Static FreeBSD binary |
| `/usr/local/etc/rc.d/unifi-stubd` | `packaging/freebsd/usr/local/etc/rc.d/unifi-stubd` | rc.d service |
| `/usr/local/share/doc/unifi-stubd/LICENSE` | `LICENSE` | Project license |
| `/usr/local/share/doc/unifi-stubd/NOTICE.md` | `NOTICE.md` | Third-party notices |
| `/usr/local/share/doc/unifi-stubd/CREDITS.md` | `CREDITS.md` | Research and attribution |

## Integration And Deployment Examples

Sanitized Tier 2 integration and host examples live under `lab/stub/`. They
are test inputs and are not installed by packages. Start deployment-specific
configs from the packaged example for the target operating system.

Real controller URLs, site IP addresses, MAC addresses, client names, adoption
state paths, and private topology mappings must not be committed. Keep live
host configs in a private deployment directory. Firmware, VM, and other Tier 3
assets belong to the companion project `unifi-stubd-labs`.
