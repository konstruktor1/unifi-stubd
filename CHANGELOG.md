# Changelog

All notable changes to this project will be documented in this file.

The format follows the spirit of Keep a Changelog, and this project uses
semantic version tags once releases are published.

## [Unreleased]

### Fixed

- Documented the required `github-pages` environment policy for automatic
  package-repository deployments from `v*` release tags.

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
