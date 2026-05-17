# unifi-stubd

`unifi-stubd` is a lab-focused UniFi device stub. It makes a Linux host,
Proxmox bridge, firewall VM, or similar non-UniFi system appear as a minimal
UniFi device in a UniFi Network Controller without allowing the controller to
provision the host.

Documentation: [English](docs/en/README.md) | [Deutsch](docs/de/README.md)

Coding-agent instructions live in [AGENTS.md](AGENTS.md). Tool-specific bridge
files only point back to that file. [llms.txt](llms.txt) is a public project
index, not an agent instruction source.

## Status

This project is experimental and intended for isolated lab networks. It is not
affiliated with, endorsed by, or supported by Ubiquiti.

Implemented:

- UniFi discovery packet builder and sender.
- Inform packet encode/decode foundation.
- Minimal fake device payloads with selectable device profiles.
- Experimental stub-only `UGW3` gateway identity profile.
- Experimental stub-only `UXG` Gateway Lite identity profile.
- Experimental stub-only `UXGPRO` 10G gateway identity profile.
- Built-in SSH shim for advanced adoption commands.
- YAML configuration under `/etc/unifi-stubd/config.yaml`.
- OpenRC and systemd service definitions.
- Package builders for Debian, RPM, Arch Linux, and `.tar.gz`.
- Stub-only FreeBSD/OPNsense tarball with rc.d service artifact.

Not goals:

- It is not a UniFi gateway replacement.
- It must not blindly apply controller provisioning to the host.
- It does not reproduce full UniFi DPI, firewall, or routing behavior.

## Quick Start

```sh
git clone https://github.com/konstruktor1/unifi-stubd.git
cd unifi-stubd
make check
go run ./cmd/unifi-stubd -dry-run
```

List built-in profiles:

```sh
go run ./cmd/unifi-stubd -list-profiles
```

The `ugw3` profile reports a legacy UniFi Security Gateway identity with three
1G ports. It is useful for gateway-profile experiments, but it is still a
stub-only identity profile and does not implement routing, DHCP, firewall, DPI,
or WAN health behavior.

The `uxg-lite` profile reports a two-port UniFi Gateway Lite identity with LAN
and WAN roles. It is useful for comparing newer `uxg` gateway payload shape
against `uxgpro` without running a real firmware rootfs.

The `uxgpro` profile reports a UniFi Next-Generation Gateway Pro identity with
the original gateway-style layout: WAN1 on 1G RJ45, LAN on 1G RJ45, WAN2 on
10G SFP+, and LAN2 on 10G SFP+. Like `ugw3`, it is an identity and
status-payload stub only. Use `uplink_port` and `port_overrides` when a lab
uses the SFP+ port as the active WAN.

Profiles describe hardware. Use YAML `port_overrides` to map lab assignments
such as WAN, LAN, or backup WAN onto those profile ports via `role` and
`network_group`.

Send discovery traffic in a lab:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac auto \
  -ip 192.0.2.50 \
  -hostname auto
```

Send discovery plus one inform heartbeat:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac auto \
  -ip 192.0.2.50 \
  -hostname auto \
  -controller http://192.0.2.10:8080/inform \
  -once
```

The `192.0.2.0/24` addresses are documentation examples. Replace them with
addresses from your isolated lab network.

## Configuration

Install a YAML config for service usage:

```sh
sudo install -m 0755 ./unifi-stubd /usr/local/bin/unifi-stubd
sudo install -d -m 0755 /etc/unifi-stubd /var/lib/unifi-stubd
sudo install -m 0600 packaging/linux/etc/unifi-stubd/config.yaml /etc/unifi-stubd/config.yaml
sudo /usr/local/bin/unifi-stubd
```

Runtime layout:

```text
/usr/local/bin/unifi-stubd
/etc/unifi-stubd/config.yaml
/etc/unifi-stubd/ssh_host_rsa_key
/var/lib/unifi-stubd/adoption.env
/var/lib/unifi-stubd/status.json
/var/log/unifi-stubd.log
/var/log/unifi-stubd.err
```

Without arguments, `unifi-stubd` tries to read
`/etc/unifi-stubd/config.yaml`. If that file is absent, it uses safe lab
defaults. If `-config <path>` is set explicitly, a missing file is an error.
CLI flags override YAML values.

The packaged Linux config source is
`packaging/linux/etc/unifi-stubd/config.yaml`. Lab switch identities and
commands live in `lab/`, and installed Linux paths are documented in
`packaging/installed-files.md`.

Local health/status output:

```sh
unifi-stubd -status
unifi-stubd -status-json
```

The status command reads local config and state only. It reports identity,
operation mode, adoption state, observe counters/FDB counts, and the last
inform response without printing the adoption authkey.

Runtime modes are documented in
[English](docs/en/operation-modes.md) and
[Deutsch](docs/de/operation-modes.md). The default `stub` mode remains fully
synthetic. The `observe` mode is read-only and can merge Linux bridge FDB and
sysfs counters into the switch payload when `observe_interface` and/or
`observe_bridge` are configured.

FreeBSD/OPNsense support is documented in
[English](docs/en/freebsd.md) and [Deutsch](docs/de/freebsd.md). It is
currently stub-only: discovery, inform, adoption SSH, profiles, port overrides,
uplink overrides, and configured uplink neighbors are supported; Linux
observation and macvlan modes are not.

## Services

OpenRC:

```sh
sudo install -m 0755 packaging/linux/etc/init.d/unifi-stubd /etc/init.d/unifi-stubd
sudo rc-update add unifi-stubd default
sudo rc-service unifi-stubd restart
```

Systemd:

```sh
sudo install -m 0644 packaging/linux/usr/lib/systemd/system/unifi-stubd.service /etc/systemd/system/unifi-stubd.service
sudo systemctl daemon-reload
sudo systemctl enable --now unifi-stubd.service
```

## Packages

Native packages are built with nFPM, and the `.tar.gz` package is built from
the same staging tree:

```sh
make package
```

Individual formats:

```sh
make package-deb
make package-rpm
make package-arch
make package-tgz
make package-freebsd-tgz
```

FreeBSD/OPNsense package builds default to `amd64`; set
`PKG_FREEBSD_GOARCH=arm64` for ARM FreeBSD hosts.

Common overrides:

```sh
PKG_VERSION=0.1.0 PKG_RELEASE=1 PKG_GOARCH=amd64 \
  PKG_MAINTAINER='Name <email@example.com>' make package
```

Artifacts are written to `dist/packages/`.

## Development

The repository keeps the Go requirement as a minor-version floor:

- `go.mod`: minimum supported module version, currently Go `1.25`.
- `go.work`: repository workspace using the same Go minor version.

Build tools are tracked as Go module tools and run through `go tool`, so no
separate global `golangci-lint` or `nfpm` install is required.

```sh
make lint
make test
make package
```

The lint profile uses `golangci-lint` and a small repository policy check for
project-specific rules such as keeping Go tests under `tests/` and keeping lab
secrets out of commits.

## Security

Run this only in isolated lab or management networks. Adoption keys, PCAPs,
MAC tables, DHCP information, and controller responses can contain sensitive
data. Report private security issues to `info@spinas.org` and see
[SECURITY.md](SECURITY.md) before sharing captures or logs.

## Credits and License

`unifi-stubd` is licensed under
[AGPL-3.0-or-later](LICENSE) so redistributed or network-accessible modified
versions must keep source available under the same copyleft terms. Research
sources, idea provenance, third-party notices, and thanks are documented in
[CREDITS.md](CREDITS.md) and [NOTICE.md](NOTICE.md).
