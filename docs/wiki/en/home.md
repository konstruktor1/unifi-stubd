# unifi-stubd Wiki

This wiki is the practical navigation layer for `unifi-stubd`. Use it when you
need to understand what the project is, where a change belongs, how to operate
the stub safely, and which tests prove a change.

## Project In One Paragraph

`unifi-stubd` is an experimental lab tool that makes a Linux host, Proxmox
bridge, FreeBSD/OPNsense host, or VM appear as a minimal UniFi device to a
UniFi Network Controller. It can emulate switch-shaped devices and experimental
gateway identities, but it must not become a controller-managed host agent. The
controller may update local adoption state; it may not provision the host.

## Start Here

| Goal | Page |
| --- | --- |
| Install or run the stub safely | [Operator Guide](operator-guide.md) |
| Understand code ownership and data flow | [Architecture Map](architecture-map.md) |
| Plan or run tests | [Testing Guide](testing-guide.md) |
| Understand current architectural decisions | [Decisions](decisions.md) |
| Read full user documentation | [English Documentation](../../en/README.md) |
| Read architecture reference | [Architecture](../../en/architecture.md) |
| Read operation-mode details | [Operation Modes](../../en/operation-modes.md) |
| Read project status | [Project Status](../../en/project-status.md) |
| Read roadmap | [Roadmap](../../en/roadmap.md) |

## Safety Boundary

The project has one non-negotiable boundary:

- controller responses may update local stub state;
- read-only observation may enrich payload and status;
- controller provisioning must not mutate host networking or execute arbitrary
  shell commands.

This applies to `stub`, `bridge-observe`, `port-map`, management LAN metadata,
passive LLDP, logs, D-Bus capability checks, and future platform adapters.

## Current Product Shape

Supported and active:

- switch-style discovery and inform payloads;
- advanced-adoption SSH shim with limited command handling;
- adoption-state persistence and restore-default/forget reset behavior;
- data-driven built-in and external profiles;
- `bridge-observe` for Proxmox/Linux bridge representation;
- `port-map` for explicit port-to-interface mapping;
- Linux platform facade for sysfs, procfs, journalctl, D-Bus probe, and lldpd;
- conservative FreeBSD/OPNsense support;
- Docker controller integration smoke tests.

Experimental:

- gateway identity profiles;
- full topology representation;
- management LAN modeling beyond metadata and preexisting interface binding;
- FreeBSD bridge-observe parity.

Not goals:

- production gateway replacement;
- full UniFi OS replacement;
- blind controller provisioning;
- automatic host VLAN creation in the current release.

## Repository Map

| Path | Purpose |
| --- | --- |
| `cmd/unifi-stubd/` | CLI, config layering, validation, daemon orchestration |
| `internal/config/` | YAML schema and defaults |
| `internal/device/` | profiles, registry, identities, and resolved ports |
| `internal/device/payload/` | switch/gateway JSON payload renderer |
| `internal/observe/` | read-only observation model |
| `internal/platform/` | OS facade for read-only host integrations |
| `internal/inform/` | inform packet crypto, padding, HTTP response handling |
| `internal/adoption/` | adoption response parsing and local state |
| `internal/adoptionssh/` | minimal SSH compatibility shim |
| `tests/` | all Go tests |
| `docs/en/`, `docs/de/` | detailed user and project documentation |
| `docs/wiki/` | this navigation and runbook layer |

## Maintenance Rule

When a change affects behavior, update the detailed documentation first and then
update the wiki only if the user/operator navigation changes. Avoid copying
full sections from the reference docs into the wiki; link to them and summarize
the decision.
