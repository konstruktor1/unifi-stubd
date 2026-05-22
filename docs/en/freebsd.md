# FreeBSD and OPNsense

FreeBSD support is still conservative. The daemon can be cross-built and
packaged with FreeBSD paths. Stub mode is the supported baseline; initial
read-only observation helpers are present for bridge FDB parsing and explicit
port-map interface metadata.

Supported in the FreeBSD package:

- `operation_mode: stub`
- UniFi discovery and inform traffic
- Built-in advanced-adoption SSH shim
- Static profile payloads, port overrides, uplink port override, and configured
  uplink neighbor entries
- `operation_mode: port-map` for explicit read-only port states. Interface
  metadata is copied from existing OS interfaces when available.
- FreeBSD bridge FDB parsing for `ifconfig <bridge> addr`
  with the same bridge-member classification used on Linux: `bridge0` is
  backplane metadata, configured or single physical-looking members are uplink
  candidates, and `tap*`/`epair*`/`vnet*` members are access ports.
- `log_source: syslog` for read-only status/log metadata from a configured file,
  defaulting to `/var/log/messages`; OPNsense commonly uses
  `/var/log/system/latest.log`, which is root-only by default
- `lldp_source: lldpd` when `lldpcli` is installed and reachable
- optional D-Bus availability checks only; D-Bus is not required
- rc.d service artifact for FreeBSD and OPNsense-style systems

Not supported on FreeBSD yet:

- Full `operation_mode: bridge-observe` parity with Linux
- `/sys/class/net` counters and speed detection
- Linux `/proc` support
- `operation_mode: macvlan`
- native `SIOCGIFMEDIA`/netlink-style subscriptions

## Build

Cross-build the FreeBSD binary:

```sh
make build-freebsd
```

Build the FreeBSD/OPNsense stub package tarball:

```sh
make package-freebsd-tgz
```

The FreeBSD targets default to `amd64`, matching common OPNsense installs. Use
`PKG_FREEBSD_GOARCH=arm64` when targeting an ARM FreeBSD host. The tarball is
written to `dist/packages/` and stages this layout:

```text
/usr/local/bin/unifi-stubd
/usr/local/etc/unifi-stubd/config.yaml
/usr/local/etc/unifi-stubd/ssh_host_rsa_key
/usr/local/etc/rc.d/unifi-stubd
/var/db/unifi-stubd/adoption.env
/var/db/unifi-stubd/status.json
/var/log/unifi-stubd.log
```

## Install Published Tarballs

Published alpha tarballs are available from GitHub Releases and from the
GitHub Pages package site. The Pages site provides stable URLs for package
manager documentation and direct FreeBSD/OPNsense downloads:

```text
https://konstruktor1.github.io/unifi-stubd/freebsd/amd64/
https://konstruktor1.github.io/unifi-stubd/freebsd/arm64/
```

Fetch the tarball for the host architecture. Use `amd64` for typical
OPNsense/x86_64 installations and `arm64` for ARM FreeBSD hosts:

```sh
ARCH=amd64 # or arm64
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${ARCH}/unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep "freebsd/${ARCH}/unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz" checksums.txt
sha256 unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz
tar -tzf unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz
```

The `sha256` output should match the checksum entry. The tarball contains
neutral defaults only. Review or replace
`/usr/local/etc/unifi-stubd/config.yaml` after extraction and before enabling
the service.

## Service

Install or extract the tarball on the FreeBSD/OPNsense host, then enable the
service with rc.conf or OPNsense tunables:

```sh
sudo tar -xzf unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sysrc unifi_stubd_enable=YES
service unifi-stubd start
```

The packaged config source is
`packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml`. Keep
`operation_mode: stub` for the lowest-risk package default. Use `port-map` only
in isolated labs where every profile port is explicitly configured as
`interface`, `disabled`, or `unmapped`.

For OPNsense log metadata, keep `log_source: off` unless the service user can
read the selected file. A read-only runtime test on OPNsense 26.1/FreeBSD 14.3
validated `syslog_path: /var/log/system/latest.log` when the binary was run as
root. The default `/var/log/messages` remains the generic FreeBSD path.

## Native Helpers

Go remains the implementation language. C++ or native helper binaries are not
used. If FreeBSD later needs data that Go cannot read cleanly, add a small
reviewed helper only after the concrete OS limitation is proven. The next likely
FreeBSD improvement is an `x/sys/unix` path for `SIOCGIFMEDIA`.
