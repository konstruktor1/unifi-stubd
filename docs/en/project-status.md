# Project Status

Last updated: 2026-05-20.

This page describes the `unifi-stubd` product line: the Go daemon, its public
configuration surface, payload model, safety boundaries, packaging, and
validation status. Firmware wrappers and UDM Pro SE VM experiments are tracked
separately in [Lab Project Status](lab-project-status.md).

## Product Definition

`unifi-stubd` is a lab-focused UniFi device stub. It makes a Linux or FreeBSD
host, VM, container, or bridge appear to a UniFi Network Controller as a
minimal UniFi device for controlled experiments.

The product is not a gateway replacement and does not apply controller
provisioning to the host. It owns only the stub identity, payload generation,
adoption-state persistence, read-only host observation, packaging, and local
test harnesses.

## Current Capabilities

The daemon currently provides:

- Discovery and inform framing for synthetic UniFi device identities.
- Built-in data-driven switch profiles and experimental gateway identity
  profiles.
- External YAML profile loading, validation, export, and templates.
- YAML service configuration plus CLI overrides and a `-validate` path.
- `stub`, `bridge-observe`, and `port-map` operation modes.
- Read-only Linux observation through sysfs/procfs, bridge FDB data,
  journalctl checks, optional D-Bus availability checks, and passive LLDP
  through `lldpd`.
- Read-only FreeBSD observation through ifconfig/netstat/syslog-oriented
  adapters.
- Adoption-state persistence for controller inform responses.
- A constrained adoption SSH shim for advanced adoption compatibility.
- Docker controller integration tests against the project-owned UniFi Network
  Application lab.
- Linux and FreeBSD package/build targets.

## Operation Modes

`stub` is fully synthetic. It uses the selected profile and local configuration
to render a deterministic controller-facing device.

`bridge-observe` represents a bridge-like host setup, for example a Proxmox
bridge. The bridge is the observation boundary; learned participant MACs are
projected into the virtual port table without mutating host networking.
Bridge members are classified before projection: the bridge device is treated
as backplane metadata, VM/container member interfaces become access ports, and
the configured physical uplink is kept separate from local participants. MACs
learned on the physical uplink are treated as remote behind the real upstream
switch and are filtered out of local access-port MAC tables.

`port-map` maps UniFi profile ports to explicit OS interfaces, or declares them
as `disabled` or `unmapped`. Mapped ports inherit read-only physical state such
as MAC address, link state, speed/media where available, addresses, counters,
and LLDP neighbors.

For Proxmox-style bridge representation, topology is controller-derived. A
configured `uplink_neighbor` can make UniFi Network report the upstream device
and last connection on the stub's uplink port. If the stub reports the physical
host or bridge MAC, a real UniFi upstream switch may already observe that MAC on
its own port; UniFi Network can then prefer that real observation and render the
link direction incorrectly. Pure representation tests should therefore prefer a
synthetic locally administered stub MAC unless the physical-MAC heuristic is the
explicit test target.

## Configuration Surface

Configuration is intentionally explicit:

- Device identity comes from profile data plus CLI/config overrides.
- Hardware shape comes from built-in or external YAML profiles.
- Lab-specific port assignments come from `port_overrides`, observation mode
  configuration, and `port_mappings`.
- Switch management LAN intent is modeled through `management_lan`.
- Legacy public `management_vlan` configuration has been removed. Controller-
  facing payload fields still use UniFi-compatible names where required.

`-validate` checks the full runtime configuration without starting the daemon.
`-profile-validate` checks profile files or directories in isolation.

## Safety Boundary

The safety boundary remains the core product rule:

- Controller-triggered restart, upgrade, shell, firewall, route, user, and host
  network changes are not blindly executed.
- Controller provisioning data may be parsed and summarized, but it stays
  metadata unless a future reviewed local adapter explicitly implements a
  narrow action.
- The SSH shim recognizes only the small command shape needed for adoption and
  local stub reset behavior.
- Discovery and inform traffic remain opt-in and belong only in isolated lab or
  management networks.

## Validation Status

Current automated validation includes:

- `go test ./...`
- `make check`
- package build targets through `make package`
- configuration/profile validation for packaged Linux and FreeBSD configs
- Docker integration tests for dry-run payloads, inform MITM capture,
  controller pending state, controller-triggered adoption, and persisted local
  adoption state

The Docker integration path also covers `bridge-observe`, `port-map`, gateway
payload rendering, and the current `management_lan.mode:
preexisting-interface` switch path.

Manual real-host validation on 2026-05-20 additionally confirmed that:

- a Linux bridge can be represented as a 48-port switch profile with VM/container
  participants on normal access ports and the physical uplink on a dedicated
  SFP+ profile port via `uplink_port`;
- unused bridge ports should be reported disconnected rather than synthetic-up;
- explicit `uplink_neighbor` metadata is enough for UniFi Network to calculate a
  topology edge when it does not conflict with the controller's real switch view;
- using the host's physical bridge/NIC MAC can reverse the displayed topology
  direction when the real upstream switch already reports that MAC.

## Packaging Status

The repository contains package definitions for:

- Linux service packaging with systemd/OpenRC-oriented configuration.
- FreeBSD/OPNsense-oriented configuration and tarball output.
- Non-root Linux service execution with documented capability handling for the
  adoption SSH compatibility port.

Packaged defaults are lab-oriented and must still be reviewed before use in any
shared management network.

## Known Product Limits

- Gateway profiles are identity and payload stubs, not full gateway behavior.
- External profiles are data-driven but still require validation against a
  concrete UniFi Network version.
- LLDP support is passive and currently depends on `lldpd` output.
- LLDP is not required for adoption or manual topology hints, but without it
  `uplink_neighbor` remains manual and topology direction is still subject to
  UniFi Network's own device/MAC heuristics.
- Linux `/proc` supplements sysfs and bridge data; it is not a complete
  replacement for OS-specific interface APIs.
- FreeBSD support is intentionally conservative and still lacks richer media
  detail compared with future native ioctl/netlink-style adapters.
- Multi-device simulation under one process is not the default design yet.

## Next Product Work

Near-term product work should focus on:

- Adding more pinned UniFi Network versions to the Docker compatibility matrix.
- More profile fixtures for custom switch and gateway payloads.
- More structured status for passive LLDP, log readers, and platform
  capabilities.
- First-class topology metadata for uplink neighbor remote-port reporting,
  including an explicit configured remote port and LLDP-derived fallback.
- Better guidance and tooling for choosing synthetic versus physical stub MACs
  in bridge-observe deployments.
- Better FreeBSD interface media/counter detail.
- CI coverage for package artifacts, SBOM, and dependency scanning.
- Release signing once package CI is stable.

Keep firmware images, captures, adoption keys, controller tokens, private
controller URLs, SSH host keys, MAC tables, and client data out of Git.
