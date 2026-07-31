# Stub Lab Plan

## Scope

This repository contains only the lab fixtures that directly validate the
`unifi-stubd` build. Firmware research, gateway simulations, UTM/QEMU virtual
machines, and machine-backed runtime tooling live in the separate
`unifi-stubd-labs` repository.

The in-repository lab is `lab/stub`. It is the controller product gate for US8
discovery, adoption, connected metric reads, and read-only bridge observation.
It does not depend on the companion repository or on an external lab-data
volume.

## Safety Boundary

- Use only an isolated lab or management network.
- Use documentation addresses such as `192.0.2.0/24` in committed examples.
- Never commit controller tokens, adoption keys, captures, certificates, real
  MAC tables, or private controller addresses.
- Controller commands must not trigger arbitrary shell execution, upgrades,
  restarts, or host-network changes.

## Validation Entry Points

Repository checks and packaging use the configured Build Runner contract; this
document defines no direct host-side Go or Compose invocation.

Current product qualification enters through
`make integration-nonunifi-metrics`. It researches and resolves the latest
published Stable and latest published Release Candidate at matrix start and
runs the same candidate in both lanes. Failure to resolve either published
lane blocks qualification; there is no fixed-version fallback. Companion QEMU
machines are separate compatibility paths, not an alternative release entry.

## Docker Controller Regression Fixture

The repeatable regression fixture consists of:

```text
lab/stub/compose.yaml
lab/stub/compose.tests.yaml
lab/stub/mongo-init-unifi.sh
lab/stub/configs/
lab/stub/scripts/
lab/stub/tools/
```

The complete Compose project is created and removed atomically by the runner
payload. Containers have no fixed host-global names or restart policy; all
named volumes and networks are removed at the end of every run. This fixture
provides focused controller regression evidence for the US8 adoption,
connection, and metrics outcome.

## Exact-Candidate Release Qualification

The matrix binds the exact same candidate bytes, source snapshot, profile, and
assertions to both controller lanes. For each lane it records:

- the official channel, release, resolved version, release source, and official
  artifact SHA;
- the resolved immutable runtime digest and the actual API-reported controller
  version;
- the candidate identity and SHA, controller-visible product telemetry, and
  complete cleanup evidence.

Each lane must prove the product outcome:

- the non-UniFi host or bridge appears as the configured evidence-bound profile,
  is adopted, and remains connected;
- controller readback shows the expected identity and selected-uplink link
  state and source-observed speed;
- both cumulative counters advance between controller reads during bounded
  bidirectional traffic and current RX/TX rates are positive;
- controller, database, Stub state, and run-specific resources are
  completely cleaned up.

A candidate-bound safety check must separately prove that controller input
cannot provision or mutate the host; this is not duplicated in each controller
lane. That retained candidate binding is currently still open.

The release matrix passes only when both the researched Stable and researched
Release Candidate lanes pass with the same candidate bytes. It does not prove
physical Layer 2 behavior, real Proxmox bridges, FreeBSD or OPNsense runtime
behavior, firmware behavior, or physical UniFi devices. Those remain separate
machine-backed concerns in `unifi-stubd-labs`.
