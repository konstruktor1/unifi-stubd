# Decisions

This page records architectural decisions that should remain visible while the
project evolves. It is a compact index; the detailed rationale lives in the
normal documentation and code comments.

## Controller Provisioning Is Not Applied

Decision: controller provisioning data is parsed, summarized, or persisted only
when safe for future inform traffic. It is not applied to the host.

Reason: the project is a lab stub, not a managed host agent. Applying unknown
controller commands to a Linux or FreeBSD host would violate the safety boundary.

## Profiles Are Data

Decision: device shape belongs in profile data, not model-name branches.

Reason: external profiles and future compatibility work need a predictable
data-driven path. Payload rendering should use `payload.kind`, ports, media,
roles, network groups, and profile defaults.

## YAML Extends Merges Before Typed Decode

Decision: external profile inheritance is resolved at YAML mapping level and
decoded once into the canonical model.

Reason: this preserves explicit zero-value overrides and avoids hand-written
field copy cascades.

## Bridge Observe Is Role-Based

Decision: bridge FDB rows are classified as bridge, uplink, access, unknown, or
ignored before payload mapping.

Reason: a Proxmox bridge contains local participants and remote infrastructure.
The upstream switch and its clients must not be rendered as direct local
access-port clients.

## Prefer Synthetic MACs For Representation Tests

Decision: representation tests should use a locally administered synthetic stub
MAC unless the physical-MAC heuristic is the test target.

Reason: UniFi Network can prefer what a real upstream UniFi switch reports about
a physical host MAC and reverse the displayed topology edge.

## Platform Integrations Are Optional

Decision: LLDP, journalctl, syslog, procfs, and D-Bus are optional read-only
sources behind `internal/platform`.

Reason: Linux and FreeBSD environments vary. Missing optional tools should be
visible in status, not installed or treated as fatal by default.

## Management LAN Does Not Create VLANs

Decision: management LAN support is metadata-only or bound to a preexisting
interface in the current release.

Reason: active VLAN lifecycle is host mutation and needs a separate reviewed
design.

## Tests Live Under `tests/`

Decision: Go test files live under `tests/`, not beside internal packages.

Reason: this is a project policy enforced by `make check` and keeps production
package directories focused on runtime code.

