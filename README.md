# unifi-stubd — Add Non-UniFi Hosts To UniFi Network

`unifi-stubd` lets a Proxmox or Linux-bridge host appear in UniFi Network, be
adopted, remain connected, and expose real local read-only link and traffic
metrics without letting the controller mutate the host. Minimal UniFi switch
identities are the representation mechanism, not the product outcome. The
FreeBSD/OPNsense package path remains experimental and stub-only.

Documentation: [English](docs/en/README.md) | [Deutsch](docs/de/README.md) |
[Wiki](docs/wiki/README.md) | [Project Status](docs/en/project-status.md) |
[Projektstand](docs/de/project-status.md) |
[Product Requirements](docs/en/requirements.md) |
[Produktanforderungen](docs/de/requirements.md) |
[Prioritized Backlog](docs/en/backlog.md) |
[Priorisierter Backlog](docs/de/backlog.md) |
[Development Workflow](docs/en/development.md) |
[Entwicklungsworkflow](docs/de/development.md)

Coding-agent instructions live in [AGENTS.md](AGENTS.md). Tool-specific bridge
files only point back to that file. [llms.txt](llms.txt) is a public project
index, not an agent instruction source.

## Non-UniFi Host Representation For Labs

`unifi-stubd` speaks the minimal UniFi discovery, inform, and adoption flows
needed to show a fake UniFi device in a UniFi Network Controller. It is meant
for isolated homelab and network-lab environments where a Linux host, Proxmox
bridge, FreeBSD system, or firewall VM should appear as a controller-visible
UniFi switch or gateway stub. Its first product outcome is to adopt one such
non-UniFi host and expose real read-only link, counter, and traffic metrics
from a locally selected interface; it does not import remote third-party
devices through UniFi's native third-party-device feature.

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
- Reports selected-uplink link state, speed, cumulative counters, and measured
  RX/TX rates when a local interface source is configured.
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
- Optional SSH shim for advanced adoption commands.
- YAML configuration under `/etc/unifi-stubd/config.yaml`.
- OpenRC and systemd service definitions.
- Package builders for Debian, RPM, Arch Linux, and `.tar.gz`.
- Stub-only FreeBSD/OPNsense tarball with rc.d service artifact.
- Current Stable/Release-Candidate compatibility gate with dynamically resolved
  controller releases and one exact Stub product image in both lanes.

Repository boundary:

- This core repository contains the Go product, packaging, deterministic tests,
  and the project-owned Docker controller gate under `lab/stub`.
- Firmware research containers, QEMU/VM runtimes, physical-infrastructure
  tests, and other Tier 3 assets are maintained in the companion project
  `unifi-stubd-labs`. They are not build, package, Main CI, or Tier 2
  dependencies of this repository.
- Built-in gateway profiles remain synthetic identity and payload stubs. They
  do not run vendor firmware and are not UniFi gateway replacements.

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

FreeBSD/OPNsense outputs are a separate manual GitHub Actions follow-up to the
Linux-first release. A `v*` tag publishes Linux packages, the SBOM, checksums,
and the Linux package site without waiting for the FreeBSD builder. When
`FreeBSD Packaging` is explicitly published for that existing pre-release, it
adds `amd64` and `arm64` tarballs, `SHA256SUMS.freebsd`, and unsigned native
FreeBSD `pkg`
repositories; those outputs can therefore appear later or remain absent:

```sh
ARCH=amd64 # or arm64
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${ARCH}/unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep "freebsd/${ARCH}/unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz" checksums.txt
sha256 unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
sudo tar -xzf unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sudo sysrc unifi_stubd_enable=YES
sudo service unifi-stubd start
```

Native FreeBSD `pkg` repository URLs are grouped by `pkg config ABI`, for
example:

```sh
pkg config ABI
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

Packages install a non-observing starter config that explicitly selects `US8`.
On Linux, set the host-specific bridge and real metric source in
`/etc/unifi-stubd/config.yaml` before starting the supported product slice.
On FreeBSD/OPNsense, review
`/usr/local/etc/unifi-stubd/config.yaml`; that path remains experimental and
stub-only rather than part of the accepted Linux/US8 metric slice.

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
go run ./cmd/unifi-stubd -profile us8 -dry-run
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

Gateway profiles do not invent firmware WAN-health or successful speed-test
telemetry from static hints or pings. Only the exact built-in `uxgpro` /
`UXGPRO` / `5.0.16.30689` DT-001 binding retains the captured all-idle
`speedtest-status` block and empty `speedtest-status-udapi`; other gateway
bindings omit those blocks.

The `ucg-fiber` profile reports a UniFi Cloud Gateway Fiber identity with four
2.5G RJ45 LAN ports, one 10G RJ45 WAN2 port, one 10G SFP+ WAN port, and one
10G SFP+ LAN port. It uses the Cloud Gateway `udm` device type and is
stub-only; it does not run UniFi OS or any controller applications.

Profiles describe hardware. Use YAML `port_overrides` to map lab assignments
such as WAN, LAN, or backup WAN onto those profile ports via `role`,
`network_group`, and optional local WAN diagnostic hints.

### Send Discovery Traffic

```sh
go run ./cmd/unifi-stubd \
  -profile us8 \
  -mac auto \
  -ip 192.0.2.50 \
  -hostname auto
```

### Send Discovery And Inform Traffic

```sh
go run ./cmd/unifi-stubd \
  -profile us8 \
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
UniFi DPI, firewall, routing, DHCP, speed-test, or provider-detection behavior.
Static WAN hints and local ping samples are local diagnostics; they are not
reported as firmware WAN-health or speed-test measurements.

## Configuration

The packaged Linux config source is
`packaging/linux/etc/unifi-stubd/config.yaml`. Controller Docker fixtures and
sanitized host examples live under `lab/stub/`, and installed Linux paths are
documented in `packaging/installed-files.md`.

Runtime layout:

```text
/usr/local/bin/unifi-stubd
/etc/unifi-stubd/config.yaml
/var/lib/unifi-stubd/ssh_host_rsa_key
/var/lib/unifi-stubd/adoption.env
/var/lib/unifi-stubd/status.json
/var/lib/unifi-stubd/instance.lock
/var/log/unifi-stubd.log
/var/log/unifi-stubd.err
```

Without arguments, `unifi-stubd` tries to read
`/etc/unifi-stubd/config.yaml`. If that file is absent, it uses safe
non-device runtime defaults, but refuses to start until a concrete profile is
selected with `-profile`. If `-config <path>` is set explicitly, a missing
file is an error. CLI flags override YAML values.

YAML is the canonical runtime format. JSON is also accepted because JSON is a
valid YAML subset, but packaged examples use block-style YAML for readability.
Only the local operator-edited YAML file is trusted for runtime choices:
controller `setparam` data can update local adoption state, but it is not
applied to host networking.

`interval_seconds` is the local base heartbeat. A positive controller interval
hint is used only when the selected built-in profile, model, and firmware
version have an exact reviewed timing binding. It is then clamped from the
stricter of the firmware and operator minimums to the lower of the operator
maximum and 300 seconds. Without that exact binding the hint is ignored and the
local base heartbeat remains active. Inform heartbeat failures use exponential
backoff up to `inform_backoff_max_seconds` with the locally configured
`inform_backoff_jitter_percent`. Transport, response-decode, JSON, dispatcher,
and non-200 results advance that bounded failure sequence. Only an empty HTTP
200 or a 200 carrying an accepted controller object resets it. Each heartbeat
uses exactly one locally selected CBC or GCM request. These settings alter only
discovery/inform timing.

Live daemon starts use `instance_guard: fail` by default. A second live process
on the same host is rejected before SSH, discovery, or inform traffic starts.
Intentional multi-stub labs can opt in with `instance_guard: warn` or
`instance_guard: off` and a reviewed `instance_guard_path`.

Keep the config as one YAML document. Block-style YAML is recommended because
new sections such as `wan_health` can be added clearly. If a file is written as
JSON-compatible YAML, every setting must live inside that one JSON object; do
not append block YAML after the closing `}`. After editing, validate and inspect
the local status before restarting a service:

```sh
unifi-stubd -validate -config /etc/unifi-stubd/config.yaml
unifi-stubd -status-json
```

The systemd unit runs as the dedicated `unifi-stubd` user and grants only
`CAP_NET_BIND_SERVICE`. Packaged defaults keep the advanced-adoption SSH shim
closed; the capability is present only so isolated labs can opt into a
UniFi-compatible low port without running the daemon as root.

The configuration schema is in `docs/schema/config.schema.json`; the profile
schema is in `docs/schema/profile.schema.json`. Use `management_lan` for switch
management VLAN metadata and for `preexisting-interface` mode when a switch
stub should report and source management traffic from an already created VLAN
interface such as `vmbr0.20`; the daemon still does not create host VLAN
interfaces.

Gateway YAML has three separate concepts that should not be mixed:

- Profile data describes the represented hardware: port count, media, speed,
  default role, and generated controller interface names such as `eth0`,
  `eth1`, or `eth2`.
- `port_overrides` describes local lab assignment and telemetry for one
  physical profile port: `role`, `network_group`, optional assignment IDs,
  optional VLAN display metadata, optional host `interface`, optional IPv4/IPv6
  addresses, and optional local WAN diagnostic hints.
- `source_interface` is diagnostic metadata copied from the host interface
  named by `port_overrides[].interface`. Host names such as `ixl0`, `vtnet0`,
  or `eth0` from the local OS must stay in `source_interface`; gateway `ifname`
  fields stay controller-facing profile names.

The gateway may still know the management or transport path used to reach the
controller through top-level identity and runtime fields such as `ip`,
`controller_url`, `discovery_interface`, and `discovery_targets`. Controller
adoption state may persist an `inform_url` internally, but that path is not a
WAN or LAN assignment. Do not use
`management_lan` to describe gateway data-plane ports; use `port_overrides` for
WAN, LAN, WAN2, and LAN2. A controller-management address must not be copied
into gateway WAN/LAN rows unless the operator explicitly assigns that address
to a gateway port.

Gateway assignment fields such as `network_group`, `portconf_id`,
`networkconf_id`, `native_networkconf_id`, `network_name`, and `vlan` are
mirrored into the payload so the controller can display the intended lab
assignment. They do not create VLANs, edit controller network profiles, change
host interfaces, or apply controller provisioning.

Ports with `role: unassigned`, disabled ports, or disconnected ports without an
explicit gateway role are reported as physical inventory only. They must not
inherit the LAN IP as a fallback. This keeps unused profile ports from looking
like additional gateway interfaces in controller web or mobile views.

For `uxgpro`, the default profile ports are:

```text
port 1 -> eth0, profile role wan,  1G RJ45
port 2 -> eth1, profile role lan,  1G RJ45
port 3 -> eth2, profile role wan2, 10G SFP+
port 4 -> eth3, profile role lan2, 10G SFP+
```

If an OPNsense lab uses port 3 as the active WAN, set `uplink_port: 3` and give
port 3 `role: wan`, `network_group: WAN`, and `interface: ixl0`. The controller
still sees `ifname: eth2` because that is the profile interface for physical
port 3; `ixl0` is reported only as `source_interface`. Non-link-local IPv6
CIDR addresses observed on `ixl0`, or configured with `port_overrides[].ipv6`,
are mirrored into the gateway WAN/network payload fields for that port.

OPNsense API data can also be used through the separate companion generator
without changing the running daemon. `unifi-stubd-opnsense` reads an existing
`unifi-stubd` config plus a separate OPNsense source file, performs read-only
GET requests against OPNsense, and prints a generated YAML document for review:

```sh
go run ./cmd/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source lab/stub/configs/hosts/opnsense-api-source.example.yaml \
  > generated.yaml
```

Use `-out generated.yaml` only when the generated file should be written by the
tool. API keys are loaded from configured files or environment variables, are
not part of the generated config, and are not logged. The generator maps
OPNsense interfaces such as `ixl0` to existing `port_overrides`; the daemon
then consumes the resulting normal YAML exactly as before.
For an on-box OPNsense setup with commands for API keys, source YAML,
generation, validation, and service restart, see the
[OPNsense API Generator How-to](docs/en/opnsense-generator.md). Field behavior,
merge rules, endpoints, and troubleshooting are documented in the
[OPNsense API Generator Reference](docs/en/opnsense-generator-reference.md).

WAN diagnostics are also explicit YAML. `source: off` runs no active probe,
`source: static` uses the `wan_*` hints from `port_overrides`, and
`source: ping` runs the local OS `ping` command against configured targets.
Ping targets must point at ports explicitly marked `role: wan` or `role: wan2`
in `port_overrides`; profile defaults alone do not enable active probes. Ping
results update only local status. They do not claim firmware availability,
downtime, or speed-test results and do not discover ISP/provider names. The
probe follows the host's normal routing table; it is not bound to a port source
interface and it does not rewrite WAN/LAN assignments.

Local health/status output:

```sh
unifi-stubd -status
unifi-stubd -status-json
```

The status command reads local config and state only. It reports identity,
operation mode, adoption state, observe counters/FDB counts, direct
infrastructure edges from the last inform, and the controller response without
printing the adoption authkey or complete client MAC tables.

Runtime modes are documented in
[English](docs/en/operation-modes.md) and
[Deutsch](docs/de/operation-modes.md). The default `stub` mode remains fully
synthetic. `bridge-observe` reads Linux bridge FDB and sysfs data without
mutating host networking; `observe` is kept as a migration alias.
`port-map` maps each represented UniFi port to an explicit local interface,
`disabled`, or `unmapped` source.

For Proxmox-style bridge representation, real device and interface MACs remain
valid identities. Use `uplink_neighbor` to provide the direct upstream device,
`remote_port` when its port is known, and `uplink_port` when the represented
link is an SFP/SFP+ port instead of the default GE fallback. FDB visibility is
kept separate from topology direction.

Optional passive sources are configured with `lldp_source`, `log_source`,
`proc_source`, and `dbus_enabled`. They feed payload/status metadata only and
do not apply controller provisioning or host-network changes.

The current product qualification dynamically resolves the latest published
Stable and latest published Release Candidate, builds the Stub once, and runs
those exact product bytes through adoption, connected-state, and controller-
visible metrics checks in both disposable lanes:

```sh
BUILD_RUNNER_DIR=/path/to/build-runner make integration-nonunifi-metrics
```

The experimental FreeBSD/OPNsense package baseline is documented in
[English](docs/en/freebsd.md) and [Deutsch](docs/de/freebsd.md). It is
currently conservative: discovery, inform, optional adoption SSH, profiles,
port overrides, uplink overrides, configured uplink neighbors, explicit
`port-map` interface reads, syslog status metadata, and `lldpd` LLDP reads are
implemented but are not part of the accepted product slice.
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
make package-freebsd-pkg-repos
```

The standalone `make package-freebsd-tgz` target defaults to `amd64`; set
`PKG_FREEBSD_GOARCH=arm64` for ARM FreeBSD hosts. FreeBSD release artifacts are
a manual GitHub Actions build through the configured FreeBSD builder and
include native `pkg` repositories for `FreeBSD:14` and `FreeBSD:15` on
`amd64`, `aarch64`, and `armv7`.
`make package-freebsd-pkg-repos` sends the source tree to that builder, builds
each configured FreeBSD ABI there, and writes the native `pkg` repos plus the
`amd64` and `arm64` tarballs. The builder must provide FreeBSD
`pkg`, `tar`, and a Go command named `go`, `go125`, or `go126`. Set
`FREEBSD_PKG_BUILD_JAILS` only for custom builders with running `jexec` jails;
Poudriere target jails are not long-running `jexec` build environments.

The `FreeBSD Packaging` workflow is manual-only. Its containerized self-hosted
GitHub Actions runner runs on `docker.dmz.lan` and is selected by
`[self-hosted, Linux, X64, docker-dmz, unifi-stubd,
freebsd-pkg-builder]`. It transfers the source through strict SSH/SCP to a
non-root FreeBSD build-host account. Build seven-day Actions evidence from
`main` or the matching version tag without publishing it:

```sh
gh workflow run freebsd-packaging.yml --ref main -f publish=false
```

After the automatic Linux workflow has created a matching `v*` pre-release,
publish the FreeBSD assets and rebuild Pages as a combined Linux/FreeBSD site:

```sh
gh workflow run freebsd-packaging.yml --ref <new-v-tag-with-freebsd-workflow> -f publish=true
```

Publication is rejected for a branch, a missing or final release, an
incomplete Linux pre-release, a release with signed checksum assets, or a
legacy `SHA256SUMS` that already includes FreeBSD. The selected tag must contain
this workflow. FreeBSD packaging has no trigger or dependency from normal CI,
Tier 2, pushes, tags, or schedules.
The runner container is a dedicated GitHub Actions SSH proxy to the FreeBSD
package host, not a general lab runner. The `docker-dmz` label identifies its
location on `docker.dmz.lan`; it does not imply access to a Docker daemon. A
bare-host installation must not grant passwordless sudo. A disposable,
containerized proxy may retain container-local passwordless sudo only when the
protected `freebsd-packaging` environment explicitly sets
`ALLOW_CONTAINER_LOCAL_SUDO=true` and the workflow preflight recognizes its
container marker, confirms an enforced seccomp filter and private container
root mount, and verifies that PID 1's root either matches the container root
when inspectable or is inaccessible even to container root because of the
sandbox. It must also find no dangerous root capabilities, exposed block
devices, host-root mounts, or host runtime sockets. This narrow exception does
not give the proxy Docker-engine capability. Tier 2 uses a separate runner
container selected by `[self-hosted, Linux, X64, docker-dmz, docker-engine,
unifi-stubd]`; both containers may share `docker.dmz.lan`. Protect the
`freebsd-packaging` environment with required reviewers and allow only `main`
and `v*` deployment refs.

Common overrides:

```sh
PKG_VERSION=0.2.5-alpha PKG_RELEASE=1 PKG_GOARCH=amd64 \
  PKG_MAINTAINER='Name <email@example.com>' make package
```

Artifacts are written to `dist/packages/`.

Unsigned GitHub Pages package repositories can be built from those artifacts:

```sh
make package-repos
```

The generated site is written to `dist/package-site/`. The automatic package
workflow publishes APT, RPM, Arch Linux, and Linux tarball paths first. A
successful `FreeBSD Packaging` publication rebuilds that site with the
FreeBSD/OPNsense tarball paths and native FreeBSD `pkg` repository paths added.

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

GitHub Actions reads `go.mod` with `check-latest: true`, so CI uses the newest
available Go `1.25` patch without committing a patch or `toolchain` pin.

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

The staged development path is documented in
[docs/en/development.md](docs/en/development.md) and
[docs/de/development.md](docs/de/development.md): topic branches target `dev`,
`main` is the tested release base, and package repositories publish only from
tags, GitHub releases, or an explicit workflow run from `main`. The normal
path is Linux-first; `FreeBSD Packaging` is an independent manual GitHub
Actions follow-up for an exact ref or existing pre-release. Packaging is not a
Tier 1, Tier 2, or Tier 3 test classification.

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
