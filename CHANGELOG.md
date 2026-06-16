# Changelog

All notable changes to this project will be documented in this file.

The format follows the spirit of Keep a Changelog, and this project uses
semantic version tags once releases are published.

## [Unreleased]

## [0.2.0-alpha] - 2026-06-16

### Changed

- Promoted the current alpha release line to `0.2.0-alpha` and refreshed
  release, development, README, and FreeBSD/OPNsense package examples so
  operator-facing instructions point at the current alpha artifact names.

### Fixed

- Rebuild discovery announcements from the current adoption state on every
  heartbeat, so an adopted stub no longer keeps advertising factory/default
  state after the controller has issued and persisted an inform key.

## [0.1.12-alpha] - 2026-06-16

### Fixed

- Rebuild discovery announcements from the current adoption state on every
  heartbeat, so an adopted stub no longer keeps advertising factory/default
  state after the controller has issued and persisted an inform key.

## [0.1.11-alpha] - 2026-06-05

### Fixed

- Removed the unsupported native FreeBSD `pkg` `post-upgrade` script key. The
  config migrator remains in `post-install`, which `pkg` also runs after
  upgrades.

## [0.1.10-alpha] - 2026-06-05

### Added

- Added `-config-migrate` and `-config-migrate-dry-run` for conservative YAML
  config normalization. The migrator handles known legacy aliases, validates the
  rewritten config before writing, and keeps timestamped `.bak.*` backups.
- Added native FreeBSD `pkg` install and upgrade hooks that run the safe config
  migrator without aborting package upgrades on manual conflicts.

### Changed

- Replaced the Docker lab's Python/mitmproxy assertion layer with Go tools and a
  small Go inform proxy.
- Simplified Docker integration coverage to deterministic payload checks and
  captured inform request events, avoiding fragile controller-version and
  container-state assumptions.

### Fixed

- Marked the native FreeBSD/OPNsense package config as package-managed config so
  `pkg` preserves local edits and writes `.pkgnew` on unmergeable default
  changes.

## [0.1.9-alpha] - 2026-06-05

### Added

- Added explicit gateway WAN ping health for locally configured `wan`/`wan2`
  ports.

### Changed

- Refactored project-wide Go naming and low-value comments without changing
  public CLI, YAML, JSON, payload, or protocol names.
- Documented the staged development workflow from topic branches through `dev`,
  `main`, and release/package publishing.
- Tightened CI package validation so `main` builds packages from the latest
  `v*` tag version and installs the generated Debian package in GitHub Actions.
- Updated `dev` CI to use the latest Go patch release for the module minor
  version.

## [0.1.8-alpha] - 2026-06-01

### Fixed

- Repacked native FreeBSD `pkg` repository packages after `pkg create` so the
  final package `+MANIFEST` keeps simple checksum-only file entries. This
  avoids the OPNsense `pkg` 2.3.1 extraction crash when migrating hosts with
  tarball-installed, unregistered files.

## [0.1.7-alpha] - 2026-06-01

### Fixed

- Wrote simple checksum-only file entries into the native FreeBSD `pkg` build
  manifest. This was superseded by the final archive repack in `0.1.8-alpha`
  because `pkg create` normalized those entries back into file objects.

## [0.1.6-alpha] - 2026-06-01

### Fixed

- Adjusted native FreeBSD `pkg` plist paths as part of tarball-to-package
  migration hardening. This was superseded by the package repack format in
  `0.1.8-alpha`.

## [0.1.5-alpha] - 2026-05-29

### Added

- Added a host-global `instance_guard` with `fail`, `warn`, and `off` modes so
  packaged daemon starts reject duplicate live instances before SSH, discovery,
  or inform traffic starts.

### Fixed

- Render gateway-native `rx_rate` and `tx_rate` telemetry as bit/s and include
  matching gateway root rate summaries while leaving `*_bytes-r` fields in
  byte/s.
- Record the latest sanitized inform traffic fields in runtime status with
  explicit byte/s and bit/s units for controller/UI correlation checks.

## [0.1.4-alpha] - 2026-05-29

### Changed

- Kept packaged advanced-adoption SSH closed by default and clarified that
  `ssh_listen` is an explicit lab-only opt-in.
- Tightened gateway profile documentation so management reachability and
  gateway WAN/LAN data-plane assignments stay separate.
- Removed gateway payload fallbacks that made disconnected or unassigned ports
  look like routed LAN interfaces.
- Disabled Dependabot version updates for now to keep alpha dependency changes
  manual and reviewable.

### Fixed

- Added gateway-native `rx_rate` and `tx_rate` telemetry beside explicit byte
  rates so UniFi Network's gateway dashboard aggregation receives live WAN
  throughput instead of rendering `0 bps`.
- Added total `bytes-r` rate fields for explicit interface-rate rows while
  keeping raw gateway inform payloads read-only and deterministic.

## [0.1.3-alpha] - 2026-05-27

### Added

- Added read-only WAN health telemetry for gateway profiles with `off`,
  `static`, and `ping` sources.
- Added UXG-Pro OPNsense SFP WAN/LAN lab coverage, including profile-derived
  controller `ifname` handling and host `source_interface` diagnostics.
- Added FreeBSD/OPNsense package-config guidance for gateway port mapping and
  WAN health.

### Changed

- Refactored device profiles, port overrides, payload construction, status
  output, serve orchestration, adoption helpers, inform framing, and platform
  observation code into smaller modules.
- Split gateway payload construction into identity, network, table, neighbor,
  health, and type modules.
- Expanded gateway YAML, schema, and operation-mode documentation to clarify
  `ifname` versus `source_interface`, VLAN/network metadata, and read-only
  safety boundaries.

### Fixed

- Documented the required `github-pages` environment policy for automatic
  package-repository deployments from `v*` release tags.
- Fixed gateway Docker/lab payload shape so gateway stubs do not emit
  switch-style fields that confuse UniFi Network inform handling.
- Fixed UXG-Pro OPNsense SFP lab expectations so host interface names such as
  `ixl0` and `vtnet0` cannot leak into controller `ifname` fields.

## [0.1.2-alpha] - 2026-05-22

### Added

- Expanded source and test comments around daemon orchestration, adoption and
  inform safety boundaries, UniFi payload mapping, host observation, platform
  integrations, and release-facing fixtures.
- Added package-site source attribution and expanded roadmap documentation.
- Documented FreeBSD and OPNsense tarball availability in user-facing docs.

### Fixed

- Fixed GitHub Actions `govulncheck` execution.
- Updated CI workflow setup so release checks use the latest stable Go where
  appropriate while keeping the repository Go version floor explicit.

## [0.1.1-alpha] - 2026-05-22

### Added

- Read-only traffic-rate reporting from mapped interfaces for lab UI activity.
- Gateway WAN/LAN port reporting for UXG-shaped lab stubs.
- Sanitized SFP+ lab examples for Proxmox/Linux bridge and OPNsense-style
  gateway topologies.
- Release documentation for neutral package artifacts and external private
  host configuration storage.

### Fixed

- Regenerated nFPM metadata per package build so cross-architecture release
  packages use the requested target architecture.
- Removed the OpenRC init script from nFPM Linux packages so systemd enablement
  works cleanly on Debian, RPM, and Arch-based installs.
- Fixed Arch Linux pre-release package metadata so `repo-add` can index alpha
  packages for the GitHub Pages package repository.
- Updated `golang.org/x/crypto` to the SSH vulnerability fix release required
  by `govulncheck`.
- Avoided a gateway `internet` payload shape that could make UniFi Network mark
  an adopted lab gateway as failed during inform processing.

## [0.1.0-alpha] - 2026-05-20

### Added

- Minimal UniFi discovery sender.
- Inform packet encode/decode foundation.
- Fake switch profiles and port payloads.
- Built-in adoption SSH shim.
- YAML runtime configuration.
- OpenRC and systemd service files.
- Debian, RPM, Arch Linux, and `.tar.gz` package builders.
- Repository hygiene for linting, tests, docs, and CI.
