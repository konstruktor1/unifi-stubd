# Changelog

All notable changes to this project will be documented in this file.

The format follows the spirit of Keep a Changelog, and this project uses
semantic version tags once releases are published.

## [Unreleased]

No unreleased changes yet.

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
