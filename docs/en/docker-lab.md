# Docker Controller Lab

The Docker lab under `lab/stub/` answers one product question: can the current
`unifi-stubd` binary make a non-UniFi Linux bridge appear as an `US8`, become
adopted, stay connected, and expose live metrics in UniFi Network?

The gate uses one product fixture, `stub-bridge-observe`. The remaining
services are provisioning only: UniFi Network, MongoDB, and a controller API
helper. Payload dumps, local status schemas, gateway commands, synthetic port maps, and
management-LAN variants are not separate product gates.

## Acceptance command

Run the product acceptance on the latest published Stable and latest published
Release Candidate:

```sh
BUILD_RUNNER_DIR=/path/to/build-runner make integration-nonunifi-metrics
```

The matrix asks Ubiquiti's release service for both versions at the start of
every run; neither channel is locked in the repository. It builds the product
once and runs the same bytes in both disposable controller lanes.

The acceptance succeeds only when both lanes prove:

- the actual controller reports the resolved version;
- the US8 is visible before adoption and accepts controller-triggered adoption;
- both metric reads report the US8 adopted and connected;
- bounded bidirectional bridge traffic makes both controller-visible root
  counters advance between reads;
- the controller shows an up `uplink0` whose speed matches the locally observed
  source, with positive current uplink and port rates.

These are observations at the product boundary. The gate intentionally does
not reproduce the product's own JSON structures in assertion helpers.

## Evidence and cleanup

The Stable/RC matrix records ignored evidence below
`dist/controller-channel-matrix/`, including:

- resolved Stable and RC versions and release metadata;
- product image ID and binary SHA-256;
- controller readback logs, final result, and cleanup result.

A claim applies only to those exact bytes and that run. A later product or gate
change requires another run.

Each lane starts with empty controller and database volumes. On success and
failure, the gate removes its containers, networks, named volumes, orphans, and
owned image aliases. It has no keep-running bypass and does not use fixed
container names.

## Boundary

The lab proves container-local Linux bridge/FDB observation and the resulting
controller-visible product behavior. It does not prove a physical topology,
Proxmox host bridge, FreeBSD/OPNsense runtime, or deployed controller. Those
remain separately owned machine-backed compatibility tests.
