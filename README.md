# unifi-stubd - UniFi Stub for Proxmox Lab Emulation

`unifi-stubd` is a minimal UniFi network device stub for Proxmox, Linux
bridges, and FreeBSD. It emulates UniFi switches and gateway identities in a
UniFi Network Controller without mutating the host.

Documentation: [English](docs/en/README.md) | [Deutsch](docs/de/README.md) |
[Wiki](docs/wiki/README.md) | [Project Status](docs/en/project-status.md) |
[Projektstand](docs/de/project-status.md)

Coding-agent instructions live in [AGENTS.md](AGENTS.md). Tool-specific bridge
files only point back to that file. [llms.txt](llms.txt) is a public project
index, not an agent instruction source.

## Fake UniFi Device Emulation for Homelabs

`unifi-stubd` speaks the minimal UniFi discovery, inform, and adoption flows
needed to show a fake UniFi device in a UniFi Network Controller. It is meant
for isolated homelab and network-lab environments where a Linux host, Proxmox
bridge, FreeBSD system, or firewall VM should appear as a controller-visible
UniFi switch or gateway stub.

## UniFi Controller Proxmox Switch and Gateway Use Cases

- Show a fake UniFi switch in UniFi Network Controller for lab topology tests.
- Represent Proxmox bridges and Linux bridge members as read-only UniFi switch
  ports.
- Experiment with UGW3, UXG-Lite, UXGPRO, and UCGF gateway identities without
  running UniFi OS.
- Feed passive LLDP, bridge, and port-map observations into payloads without
  controller-driven host networking changes.

## Features

- Emulates minimal UniFi switch and experimental gateway identities for
  isolated labs.
- Sends UniFi discovery and inform traffic for controller visibility and
  adoption testing.
- Persists safe adoption state while rejecting controller-triggered host
  provisioning.
- Supports read-only Linux bridge observation for Proxmox-style lab networks.
- Provides explicit port mapping for lab interfaces without changing host
  networking.
- Includes built-in UniFi device profiles, YAML configuration, and
  Linux/FreeBSD packaging assets.

## Status

This project is experimental and intended for isolated lab networks. It is not
affiliated with, endorsed by, or supported by Ubiquiti.

Implemented:

- UniFi discovery packet builder and sender.
- Inform packet encode/decode foundation.
- Minimal fake device payloads with selectable device profiles.
- Read-only `bridge-observe` mode for Linux bridge FDB and sysfs counters.
- Read-only `port-map` mode for explicit port-to-interface mapping.
- Read-only platform facade for Linux `/proc`, `journalctl`, optional D-Bus,
  FreeBSD syslog, and passive LLDP through `lldpd`.
- Bridge-member role classification for Proxmox-style bridges, including
  remote uplink MAC filtering, disconnected unused ports, and explicit
  `uplink_port` placement for SFP/SFP+ links.
- Experimental stub-only `UGW3` gateway identity profile.
- Experimental stub-only `UXG` Gateway Lite identity profile.
- Experimental stub-only `UXGPRO` 10G gateway identity profile.
- Experimental stub-only `UCGF` Cloud Gateway Fiber identity profile.
- Built-in SSH shim for advanced adoption commands.
- YAML configuration under `/etc/unifi-stubd/config.yaml`.
- OpenRC and systemd service definitions.
- Package builders for Debian, RPM, Arch Linux, and `.tar.gz`.
- Stub-only FreeBSD/OPNsense tarball with rc.d service artifact.
- Docker lab integration smoke test against a real UniFi Network Application
  container, including controller-triggered adoption.

Lab/research status:

- Gateway firmware simulation profiles live under `lab/gateway-profiles/` and
  are intentionally separate from `internal/device` stub profile data.
- The UDM Pro SE VM reference under `lab/gateway-profiles/udm-pro-se-vm/`
  reaches the UDM firmware `systemd` path with a foreign QEMU-virt-capable
  ARM64 kernel, UDM initramfs/rootfs, a lab initramfs, and project-owned
  userspace hardware mocks. The current UTM profile uses Shared/NAT as the
  SFP+ WAN role and Host networking as the 2.5G LAN role; direct guest HTTPS
  works, while native UTM localhost forwarding still needs verification.
- The UDM Pro SE C mock is split into modules under
  `lab/gateway-profiles/udm-pro-se/mock/ldpreload/`, and the VM/rootfs payloads
  injected by the lab initramfs live under
  `lab/gateway-profiles/udm-pro-se-vm/initramfs/`.
- The UDM Pro SE Docker webportal path is documented as a partial UniFi OS
  setup surface. Its project-owned CommonJS helpers are split under
  `lab/gateway-profiles/udm-pro-se/network-app/` and
  `lab/gateway-profiles/udm-pro-se/systemd-dbus/`.
- The QEMU/UTM profile can stage a shared ignored kernel deployment payload
  under `lab/gateway-profiles/udm-pro-se-vm/artifacts/deploy/kernel/`; UTM uses
  it for boot inputs and the Docker firmware profile mounts it read-only for
  comparison.
- The UDM Pro SE VM is a firmware reference for understanding native behavior;
  it is not the Go stub and it is not a supported UniFi OS replacement.

Not goals:

- It is not a UniFi gateway replacement.
- It must not blindly apply controller provisioning to the host.
- It does not reproduce full UniFi DPI, firewall, or routing behavior.

## Install From Package Repositories

Unsigned alpha package repositories are published through GitHub Pages at
`https://konstruktor1.github.io/unifi-stubd/`. Use them only for isolated lab or
management networks. Stable releases should use signed package repositories once
a project release key exists.

Debian, Ubuntu, and Proxmox:

```sh
echo 'deb [trusted=yes arch=amd64] https://konstruktor1.github.io/unifi-stubd/apt alpha main' | sudo tee /etc/apt/sources.list.d/unifi-stubd.list
sudo apt update
sudo apt install unifi-stubd
```

Use `arch=arm64` on ARM hosts.

Fedora, RHEL, and openSUSE-compatible RPM systems:

```ini
[unifi-stubd]
name=unifi-stubd alpha
baseurl=https://konstruktor1.github.io/unifi-stubd/rpm/$basearch
enabled=1
gpgcheck=0
repo_gpgcheck=0
```

Arch Linux and Arch Linux ARM:

```ini
[unifi-stubd]
SigLevel = Never
Server = https://konstruktor1.github.io/unifi-stubd/arch/$arch
```

FreeBSD/OPNsense builds are published as tarballs, not as a FreeBSD `pkg`
repository yet:

```sh
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/amd64/unifi-stubd_0.1.1-alpha-1_freebsd_amd64.tar.gz
sudo tar -xzf unifi-stubd_0.1.1-alpha-1_freebsd_amd64.tar.gz -C /
```

Packages install neutral defaults only. Copy the host-specific config to
`/etc/unifi-stubd/config.yaml` on Linux or
`/usr/local/etc/unifi-stubd/config.yaml` on FreeBSD/OPNsense before starting the
service.

## Installation

### Build From Source

```sh
git clone https://github.com/konstruktor1/unifi-stubd.git
cd unifi-stubd
make check
go build ./cmd/unifi-stubd
```

### Install As A Lab Service

```sh
sudo install -m 0755 ./unifi-stubd /usr/local/bin/unifi-stubd
sudo install -d -m 0755 /etc/unifi-stubd /var/lib/unifi-stubd
sudo install -m 0640 packaging/linux/etc/unifi-stubd/config.yaml /etc/unifi-stubd/config.yaml
```

Review `/etc/unifi-stubd/config.yaml` before starting the daemon. Use only
isolated lab or management networks.

## Usage

### List Available Profiles

```sh
go run ./cmd/unifi-stubd -list-profiles
```

External lab profiles can be loaded with `profile_file` or `profile_dir`.
Validate them before use.

### Validate Configuration

```sh
go run ./cmd/unifi-stubd -profile-template switch > lab-switch.yaml
go run ./cmd/unifi-stubd -profile-validate lab-switch.yaml
go run ./cmd/unifi-stubd -validate -config packaging/linux/etc/unifi-stubd/config.yaml
```

### Run A Dry Test

```sh
go run ./cmd/unifi-stubd -dry-run
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

The `ucg-fiber` profile reports a UniFi Cloud Gateway Fiber identity with four
2.5G RJ45 LAN ports, one 10G RJ45 WAN2 port, one 10G SFP+ WAN port, and one
10G SFP+ LAN port. It uses the Cloud Gateway `udm` device type and is
stub-only; it does not run UniFi OS or any controller applications.

Profiles describe hardware. Use YAML `port_overrides` to map lab assignments
such as WAN, LAN, or backup WAN onto those profile ports via `role` and
`network_group`.

### Send Discovery Traffic

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac auto \
  -ip 192.0.2.50 \
  -hostname auto
```

### Send Discovery And Inform Traffic

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

## Safety: Lab-Only, No Host Mutation

Run `unifi-stubd` only in isolated lab or management networks. The controller
may update local stub adoption state, but controller provisioning must not
mutate host networking, services, packages, firewall rules, routes, or users.

`unifi-stubd` is not a UniFi gateway replacement and does not reproduce full
UniFi DPI, firewall, routing, DHCP, or WAN health behavior.

## Configuration

The packaged Linux config source is
`packaging/linux/etc/unifi-stubd/config.yaml`. Lab switch identities and
commands live in `lab/`, and installed Linux paths are documented in
`packaging/installed-files.md`.

Runtime layout:

```text
/usr/local/bin/unifi-stubd
/etc/unifi-stubd/config.yaml
/var/lib/unifi-stubd/ssh_host_rsa_key
/var/lib/unifi-stubd/adoption.env
/var/lib/unifi-stubd/status.json
/var/log/unifi-stubd.log
/var/log/unifi-stubd.err
```

Without arguments, `unifi-stubd` tries to read
`/etc/unifi-stubd/config.yaml`. If that file is absent, it uses safe lab
defaults. If `-config <path>` is set explicitly, a missing file is an error.
CLI flags override YAML values.

The systemd unit runs as the dedicated `unifi-stubd` user and grants only
`CAP_NET_BIND_SERVICE` so the lab SSH shim can keep UniFi-compatible port 22
without running the daemon as root.

The configuration schema is in `docs/schema/config.schema.json`; the profile
schema is in `docs/schema/profile.schema.json`. Use `management_lan` for switch
management VLAN metadata and for `preexisting-interface` mode when a switch
stub should report and source management traffic from an already created VLAN
interface such as `vmbr0.20`; the daemon still does not create host VLAN
interfaces.

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
synthetic. `bridge-observe` reads Linux bridge FDB and sysfs data without
mutating host networking; `observe` is kept as a migration alias.
`port-map` maps each represented UniFi port to an explicit local interface,
`disabled`, or `unmapped` source.

For Proxmox-style bridge representation, prefer a synthetic locally
administered stub MAC unless you intentionally want to test how the controller
handles a physical host MAC already visible on an upstream UniFi switch. Use
`uplink_neighbor` to provide the upstream switch hint and `uplink_port` when the
represented link is an SFP/SFP+ port instead of the default GE fallback.

Optional passive sources are configured with `lldp_source`, `log_source`,
`proc_source`, and `dbus_enabled`. They feed payload/status metadata only and
do not apply controller provisioning or host-network changes.

The Docker controller smoke test reuses the lab controller and MITM services,
checks `bridge-observe` and `port-map` payloads, triggers adoption through the
controller API, verifies persisted adoption state, and cleans temporary test
devices again:

```sh
make integration-docker
```

FreeBSD/OPNsense support is documented in
[English](docs/en/freebsd.md) and [Deutsch](docs/de/freebsd.md). It is
currently conservative: discovery, inform, adoption SSH, profiles, port
overrides, uplink overrides, configured uplink neighbors, explicit `port-map`
interface reads, syslog status metadata, and `lldpd` LLDP reads are supported.
Full `bridge-observe` parity and macvlan lifecycle work are not implemented
there.

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

Unsigned GitHub Pages package repositories can be built from those artifacts:

```sh
make package-repos
```

The generated site is written to `dist/package-site/` and contains APT, RPM,
Arch Linux, and FreeBSD/OPNsense tarball paths for the alpha channel.

Release packages are intentionally neutral. They install the daemon, service
files, documentation, and the packaged example config only. Keep real controller
URLs, host MACs, lab addresses, client names, and adoption paths in a private
configuration store outside this repository, then copy the host-specific config
to `/etc/unifi-stubd/config.yaml` on Linux or
`/usr/local/etc/unifi-stubd/config.yaml` on FreeBSD/OPNsense after installing
the package.

Installing a package does not create VLANs, change firewall or routing rules,
alter controller network definitions, or mutate host networking. Any lab
topology mapping must be expressed explicitly in the local host config.

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
