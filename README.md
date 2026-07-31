# unifi-stubd — Add a non-UniFi host to UniFi Network

`unifi-stubd` represents a Linux host or bridge as a minimal UniFi switch. The
first product target is deliberately narrow: add one non-UniFi Linux bridge to
UniFi Network as an `US8`, adopt it, keep it connected, and show real read-only
link and traffic metrics from a named local interface.

The project is experimental, unofficial, and intended only for isolated lab or
management networks. It is not affiliated with or supported by Ubiquiti.

Documentation: [English](docs/en/README.md) · [Deutsch](docs/de/README.md) ·
[Product requirements](docs/en/requirements.md) ·
[Current backlog](docs/en/backlog.md)

## Product outcome

For the supported slice, the Stub:

- appears as an `US8` before adoption and accepts controller adoption;
- remains connected through discovery and Inform;
- reads link state, speed, counters, and rates from the configured Linux
  bridge uplink;
- sends those real values to UniFi Network without inventing missing data; and
- never applies controller provisioning to the host.

Acceptance always resolves the latest published UniFi Network Stable and
latest published Release Candidate at run start. The same Stub binary must pass
both; controller versions are recorded as run evidence, never pinned as product
configuration. See the [requirements](docs/en/requirements.md) for the complete
contract and [compatibility](docs/en/compatibility.md) for accepted runs.

## Safety boundary

Run the Stub only on an isolated lab or management network. Controller input
may update the Stub's bounded adoption state, but it must not:

- execute arbitrary shell commands;
- change interfaces, bridges, VLANs, routes, firewall rules, or users;
- restart services or the host;
- install packages or firmware; or
- turn missing observations into synthetic success values.

Discovery and Inform are opt-in network traffic. Keep real controller URLs,
adoption keys, MAC tables, captures, and client data outside the repository.

## Quick start: US8 on a Linux bridge

The product path below targets Debian, Ubuntu, and Proxmox with systemd. The
example addresses use the documentation-only `192.0.2.0/24` range; replace
them with values from the isolated network. Replace `vmbr0` and `eno1` with the
real bridge and its physical uplink.

### 1. Install the package

The current alpha repository is unsigned and must not be used on a production
network:

```sh
echo 'deb [trusted=yes arch=amd64] https://konstruktor1.github.io/unifi-stubd/apt alpha main' \
  | sudo tee /etc/apt/sources.list.d/unifi-stubd.list
sudo apt update
sudo apt install unifi-stubd
```

Use `arch=arm64` on an ARM host. RPM, Arch, tarball, FreeBSD, and OPNsense
artifacts are documented separately; they are not part of the defined
US8/Linux-bridge acceptance path described here.

### 2. Configure the real metric source

Edit `/etc/unifi-stubd/config.yaml`. Keep the complete packaged file and change
at least these fields:

```yaml
operation_mode: bridge-observe
controller_url: http://192.0.2.10:8080/inform
profile: us8
mac: auto
ip: 192.0.2.50
hostname: auto

bridge_observe:
  bridge: vmbr0
  uplink_interface: eno1

traffic_rates_enabled: true
instance_guard: fail
ssh_listen: ""
```

`uplink_interface` is the named real source for link state, speed, cumulative
RX/TX counters, and current rates. `bridge-observe` reads the bridge, sysfs,
and available FDB data; it does not change host networking.

### 3. Validate and start

```sh
sudo -u unifi-stubd /usr/local/bin/unifi-stubd \
  -validate -config /etc/unifi-stubd/config.yaml
sudo systemctl enable --now unifi-stubd.service
sudo systemctl status unifi-stubd.service
sudo -u unifi-stubd /usr/local/bin/unifi-stubd -status
```

Validation must name a concrete profile, bridge, and existing uplink. A second
live instance is rejected by the default instance guard.

### 4. Adopt and verify the product result

1. Open UniFi Network and find the pending `US8`.
2. Adopt it and wait until it remains connected.
3. Generate bounded bidirectional traffic through the configured uplink.
4. On the US8 device page, confirm the real link state, observed speed, and
   changing RX/TX traffic values.

Controller API readback is useful diagnosis, but it does not prove that an
operator can see the values. Product acceptance requires the visible device
page result on both dynamically resolved controller channels.

## Supported boundary and experimental work

The defined product scope is `US8` on a local Linux bridge with a named real
metric source. Other switch identities, gateway identities, FreeBSD/OPNsense,
active VLAN lifecycle, routing, DHCP, firewall, DPI, and WAN-health emulation
are separate experimental work and do not inherit acceptance from the US8.

The Stub may mirror explicit local metadata into a payload, but it is not a
UniFi gateway replacement and does not run vendor firmware.

## Documentation ownership

Each changing fact has one owner:

| Document | Owns |
| --- | --- |
| [Product requirements](docs/en/requirements.md) | Stable product contract and acceptance conditions |
| [Product backlog](docs/en/backlog.md) | Ordered open outcomes and the next step |
| [Product roadmap](docs/en/roadmap.md) | Product phases only |
| [Controller compatibility](docs/en/compatibility.md) | Concrete accepted Stable/RC runs |
| [Operation modes](docs/en/operation-modes.md) | Configuration and runtime behavior |
| [Architecture](docs/en/architecture.md) | Component and safety boundaries |
| [Protocol notes](docs/en/protocol-notes.md) | Discovery, Inform, and adoption wire behavior |

German equivalents are linked from the [German documentation
index](docs/de/README.md).

## Installation and operation references

- [Packaged Linux configuration](packaging/linux/etc/unifi-stubd/config.yaml)
- [Installed package files](packaging/installed-files.md)
- [FreeBSD and OPNsense](docs/en/freebsd.md)
- [OPNsense read-only generator](docs/en/opnsense-generator.md)
- [Security policy](SECURITY.md)

## Development

Build, format, lint, test, and package commands use the project's configured
Build Runner. Contributors should follow [AGENTS.md](AGENTS.md) and the
[development workflow](docs/en/development.md); the README does not duplicate
runner, CI, packaging, or release procedures.

The Go product, deterministic tests, packaging, and disposable Docker
controller gate live in this repository. Firmware research, QEMU/VM runtimes,
physical test machines, and other Tier 3 assets belong to the companion
`unifi-stubd-labs` project.

## License and sources

`unifi-stubd` is licensed under [AGPL-3.0-or-later](LICENSE). Research
provenance and license boundaries are documented in [CREDITS.md](CREDITS.md)
and [NOTICE.md](NOTICE.md).
