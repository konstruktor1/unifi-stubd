# FreeBSD and OPNsense

FreeBSD support is an experimental, conservative package baseline outside the
accepted product slice. The daemon can be cross-built and packaged with
FreeBSD paths. Stub mode is implemented; initial read-only observation helpers
are present for bridge FDB parsing and explicit port-map interface metadata.

Implemented in the experimental FreeBSD package:

- `operation_mode: stub`
- UniFi discovery and inform traffic
- Optional advanced-adoption SSH shim, disabled by default with `ssh_listen: ""`
- Static profile payloads, port overrides, uplink port override, and configured
  uplink neighbor entries
- Optional `wan_health.source: ping` for gateway profiles. It runs local
  read-only pings and reports only local WAN diagnostic status.
- `operation_mode: port-map` for explicit read-only port states. Interface
  metadata is copied from existing OS interfaces when available.
- FreeBSD bridge FDB parsing for `ifconfig <bridge> addr`
  with the same bridge-member classification used on Linux: `bridge0` is
  backplane metadata, configured or single physical-looking members are uplink
  candidates, and `tap*`/`epair*`/`vnet*` members are access ports.
- `lldp_source: lldpd` when `lldpcli` is installed and reachable
- rc.d service artifact for FreeBSD and OPNsense-style systems

Not supported on FreeBSD yet:

- Full `operation_mode: bridge-observe` parity with Linux
- `/sys/class/net` counters and speed detection
- Linux `/proc` support
- `operation_mode: macvlan`
- native `SIOCGIFMEDIA`/netlink-style subscriptions

## Build

Run `build-freebsd` and `package-freebsd-tgz` through the configured Build
Runner. [`AGENTS.md`](../../AGENTS.md) owns the invocation; do not use host Go
as a fallback.

`package-freebsd-pkg-repos` currently delegates compilation to a direct
FreeBSD host-Go builder without Provider consumer attribution. It remains an
external-machine compatibility path, not Build Runner validation or release
evidence. Agent-run package publication is blocked until an authorized VM
Runner route or reviewed split-build route replaces that compilation step.

The tarball target defaults to `amd64`, matching common OPNsense installs. Use
`PKG_FREEBSD_GOARCH=arm64` when targeting an ARM FreeBSD host. Native `pkg`
repositories include `FreeBSD:14` and `FreeBSD:15` builds for `amd64`,
`aarch64`, and `armv7`. The tarball is written to `dist/packages/` and stages
this layout:

The compatibility target sends the source tree to the configured FreeBSD
builder, builds each configured FreeBSD ABI there, and writes the
native `pkg` repositories plus the published `amd64` and `arm64` tarballs. Set
`FREEBSD_PKG_BUILD_JAILS` only for custom builders with running `jexec` jails.
The current compatibility Poudriere buildhost builds FreeBSD binaries on the
FreeBSD host with `go`, `go125`, or `go126`, then creates ABI-specific `pkg`
repositories with the FreeBSD `pkg` tool.

GitHub Actions publication, runner-boundary, retention, and legacy-tag rules
are owned by the [release checklist](release.md#manual-freebsd-packaging-extension).
This page owns target build and installation facts only.

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

Published alpha tarballs become available from GitHub Releases and the GitHub
Pages package site only after the optional manual `FreeBSD Packaging`
publication. A normal tag can therefore have Linux packages and an SBOM while
FreeBSD output is still absent. The Pages site provides stable URLs once that
workflow has completed:

```text
https://konstruktor1.github.io/unifi-stubd/freebsd/amd64/
https://konstruktor1.github.io/unifi-stubd/freebsd/arm64/
```

Fetch the tarball for the host architecture. Use `amd64` for typical
OPNsense/x86_64 installations and `arm64` for ARM FreeBSD hosts:

```sh
ARCH=amd64 # or arm64
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${ARCH}/unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep "freebsd/${ARCH}/unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz" checksums.txt
sha256 unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
tar -tzf unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
```

The `sha256` output should match the checksum entry. The tarball contains a
non-observing starter config that explicitly selects `US8`. Review or replace
`/usr/local/etc/unifi-stubd/config.yaml` after extraction and before enabling
the service.

## Install Published pkg Repositories

Native alpha `pkg` repositories are grouped by FreeBSD ABI. Check the ABI on the
target host first:

```sh
pkg config ABI
```

Then configure the matching repository path, for example for `FreeBSD:14:amd64`:

```sh
sudo mkdir -p /usr/local/etc/pkg/repos
sudo tee /usr/local/etc/pkg/repos/unifi-stubd.conf >/dev/null <<'PKGCONF'
unifi-stubd: {
  url: "https://konstruktor1.github.io/unifi-stubd/freebsd/pkg/FreeBSD:14:amd64",
  enabled: yes,
  signature_type: none
}
PKGCONF
sudo pkg update -r unifi-stubd
sudo pkg install unifi-stubd
```

These repositories are unsigned alpha artifacts. Keep them scoped to isolated
lab or management networks. The native `pkg` package marks
`/usr/local/etc/unifi-stubd/config.yaml` as a package config file, so upgrades
preserve local edits and write `config.yaml.pkgnew` when a new packaged default
cannot be merged automatically. The package `post-install` hook also runs after
upgrades and starts a conservative config migration for known legacy aliases
such as `controller`, `operation_mode: observe`, and top-level
`observe_bridge`/`observe_interface`. The migrator validates the rewritten YAML
before replacing the file and writes a timestamped `.bak.*` backup next to the
config. Ambiguous conflicts are reported but do not abort the package upgrade.

To inspect the migration before writing anything:

```sh
unifi-stubd -config /usr/local/etc/unifi-stubd/config.yaml -config-migrate-dry-run
```

To run it manually:

```sh
sudo unifi-stubd -config /usr/local/etc/unifi-stubd/config.yaml -config-migrate
```

## Service

Install or extract the tarball on the FreeBSD/OPNsense host, then enable the
service with rc.conf or OPNsense tunables:

```sh
sudo tar -xzf unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sysrc unifi_stubd_enable=YES
service unifi-stubd start
```

The packaged config source is
`packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml`. Keep
`operation_mode: stub` for the lowest-risk package default. Use `port-map` only
in isolated labs where every profile port is explicitly configured as
`interface`, `disabled`, or `unmapped`.

Use the [OPNsense generator how-to](opnsense-generator.md) for source mapping,
generation, and review.

After every config edit, validate the single YAML document, restart the service,
and inspect local status:

```sh
unifi-stubd -validate -config /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
unifi-stubd -status-json
```
