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

## Service

Install or extract the tarball on the FreeBSD/OPNsense host, then enable the
service with rc.conf or OPNsense tunables:

```sh
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
