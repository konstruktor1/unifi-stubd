# FreeBSD and OPNsense

FreeBSD support is stub-only for now. The daemon can be cross-built and
packaged with FreeBSD paths, but passive Linux observation is not implemented
there yet.

Supported in the FreeBSD package:

- `operation_mode: stub`
- UniFi discovery and inform traffic
- Built-in advanced-adoption SSH shim
- Static profile payloads, port overrides, uplink port override, and configured
  uplink neighbor entries
- rc.d service artifact for FreeBSD and OPNsense-style systems

Not supported on FreeBSD yet:

- `operation_mode: observe`
- Linux bridge FDB import
- `/sys/class/net` counters and speed detection
- `operation_mode: macvlan`

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
`operation_mode: stub` on FreeBSD until a native observation adapter exists.
