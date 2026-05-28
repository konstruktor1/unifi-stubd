# FreeBSD and OPNsense

FreeBSD support is still conservative. The daemon can be cross-built and
packaged with FreeBSD paths. Stub mode is the supported baseline; initial
read-only observation helpers are present for bridge FDB parsing and explicit
port-map interface metadata.

Supported in the FreeBSD package:

- `operation_mode: stub`
- UniFi discovery and inform traffic
- Optional advanced-adoption SSH shim, disabled by default with `ssh_listen: ""`
- Static profile payloads, port overrides, uplink port override, and configured
  uplink neighbor entries
- Optional `wan_health.source: ping` for gateway profiles. It runs local
  read-only pings and reports only WAN telemetry.
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
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${ARCH}/unifi-stubd_0.1.3-alpha-1_freebsd_${ARCH}.tar.gz
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep "freebsd/${ARCH}/unifi-stubd_0.1.3-alpha-1_freebsd_${ARCH}.tar.gz" checksums.txt
sha256 unifi-stubd_0.1.3-alpha-1_freebsd_${ARCH}.tar.gz
tar -tzf unifi-stubd_0.1.3-alpha-1_freebsd_${ARCH}.tar.gz
```

The `sha256` output should match the checksum entry. The tarball contains
neutral defaults only. Review or replace
`/usr/local/etc/unifi-stubd/config.yaml` after extraction and before enabling
the service.

## Service

Install or extract the tarball on the FreeBSD/OPNsense host, then enable the
service with rc.conf or OPNsense tunables:

```sh
sudo tar -xzf unifi-stubd_0.1.3-alpha-1_freebsd_${ARCH}.tar.gz -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sysrc unifi_stubd_enable=YES
service unifi-stubd start
```

The packaged config source is
`packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml`. Keep
`operation_mode: stub` for the lowest-risk package default. Use `port-map` only
in isolated labs where every profile port is explicitly configured as
`interface`, `disabled`, or `unmapped`.

For OPNsense gateway labs, keep the UniFi-facing interface name separated from
the FreeBSD interface name:

UXG-Pro profile ports are fixed profile data:

```text
port 1 -> eth0, profile role wan,  1G RJ45
port 2 -> eth1, profile role lan,  1G RJ45
port 3 -> eth2, profile role wan2, 10G SFP+
port 4 -> eth3, profile role lan2, 10G SFP+
```

```yaml
profile: uxgpro
uplink_port: 3
port_overrides:
  - port: 3
    role: wan
    network_group: WAN
    interface: ixl0
    speed: 10000
    media: SFP+
wan_health:
  source: ping
  interval_seconds: 10
  timeout_ms: 1000
  targets:
    - port: 3
      host: 1.1.1.1
```

With `uxgpro`, physical port 3 is controller `ifname: eth2`; `ixl0` is only
reported as `source_interface`. The ping source updates connectivity and
latency telemetry, but it does not run a UniFi speed test, detect the ISP, or
change OPNsense interfaces, routes, firewall rules, or VLANs. The ping follows
the OPNsense host routing table; `targets[].port` selects which WAN telemetry
row is updated, not which source interface the ICMP packet uses.

If the gateway also represents a LAN on another physical port, configure that
LAN explicitly in `port_overrides`. Do not use `management_lan` as a gateway
LAN shortcut. The management or transport address used to reach the controller
can stay in top-level runtime fields, while routed LAN data belongs only to the
LAN port:

```yaml
port_overrides:
  - port: 4
    role: lan
    network_group: LAN
    interface: vtnet0
    ip: 192.0.2.1
    netmask: 255.255.255.0
```

Unused or disabled profile ports should stay `role: unassigned` and should not
carry an `ip`. The payload reports them as physical inventory only, so the
controller does not receive extra LAN/Gateway hints on disconnected ports.

After every config edit, validate the single YAML document, restart the service,
and inspect local status:

```sh
unifi-stubd -validate -config /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
unifi-stubd -status-json
```

For the SFP WAN example above, the controller payload should show WAN on port 3
with `ifname: eth2`, `source_interface: ixl0`, `uplink: eth2`, and WAN health
latency/connectivity when ping succeeds. Host names such as `ixl0`, `igb0`, or
`vtnet0` must not appear in controller `ifname` fields. Provider and ISP fields
remain unset because automatic provider detection is not implemented.

For OPNsense log metadata, keep `log_source: off` unless the service user can
read the selected file. A read-only runtime test on OPNsense 26.1/FreeBSD 14.3
validated `syslog_path: /var/log/system/latest.log` when the binary was run as
root. The default `/var/log/messages` remains the generic FreeBSD path.

## Native Helpers

Go remains the implementation language. C++ or native helper binaries are not
used. If FreeBSD later needs data that Go cannot read cleanly, add a small
reviewed helper only after the concrete OS limitation is proven. The next likely
FreeBSD improvement is an `x/sys/unix` path for `SIOCGIFMEDIA`.
