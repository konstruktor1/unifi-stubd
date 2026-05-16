# Security Policy

`unifi-stubd` is intended for isolated lab and management networks only. Do not
run it on untrusted networks or production VLANs.

## Supported Versions

Only the current `main` branch is considered supported while the project is
pre-1.0.

## Reporting a Vulnerability

Please do not open public issues for secrets, credential exposure, or
controller-impacting vulnerabilities. Report privately to `info@spinas.org`,
or use GitHub private vulnerability reporting if it is enabled for the
repository.

Include:

- A short description of the issue.
- A minimal reproduction.
- Affected commit, tag, or package version.
- Whether logs, PCAPs, MAC addresses, or controller responses were involved.

## Handling Lab Data

Before sharing logs or captures, remove:

- UniFi `authkey` values.
- Controller URLs that identify a private site.
- MAC tables, DHCP leases, client names, and NetFlow/DPI records.
- SSH host keys or adoption credentials.

More notes: [English](docs/en/security.md) | [Deutsch](docs/de/security.md).
