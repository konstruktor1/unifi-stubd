# Docker Controller Lab

The Docker lab under `lab/stub/` answers one product question: can the current
`unifi-stubd` binary make a non-UniFi Linux bridge appear as an `US8`, become
adopted, stay connected, and expose live metrics in UniFi Network?

The gate uses one product fixture, `stub-bridge-observe`. The remaining
services are provisioning only: UniFi Network, MongoDB, a controller API
helper, and a browser helper. Payload dumps, local status schemas, gateway
commands, synthetic port maps, and management-LAN variants are not separate
controller gates.

## Controller gate

Run the controller-facing gate on the latest published Stable and latest published
Release Candidate:

```sh
make integration-nonunifi-metrics
```

The matrix asks Ubiquiti's release service for both versions at the start of
every run; neither channel is locked in the repository. It builds the product
once and runs the same bytes in both disposable controller lanes.

The gate succeeds only when both lanes prove:

- the actual controller reports the resolved version;
- the US8 is visible before adoption and accepts controller-triggered adoption;
- while bounded bidirectional bridge traffic runs, the device view shows the
  observed speed and positive `Down` and `Up` traffic values.

These are observations at the product boundary. The gate intentionally does
not reproduce the product's own JSON structures in assertion helpers.
Full product acceptance also requires the separate host-safety and
no-fabrication evidence defined in the [requirements](requirements.md).

## Evidence and cleanup

The Stable/RC matrix keeps uncommitted run artifacts below
`dist/controller-channel-matrix/`:

- resolved Stable and RC versions and release metadata;
- product-gate logs for both lanes.

Each lane starts with empty controller and database volumes. On success and
failure, the gate removes its containers, networks, named volumes, orphans, and
owned image aliases. It has no keep-running bypass and does not use fixed
container names.

## Boundary

The lab proves container-local Linux bridge/FDB observation and the resulting
controller-visible product behavior. It does not prove a physical topology,
Proxmox host bridge, FreeBSD/OPNsense runtime, or deployed controller. Those
remain separately owned machine-backed compatibility tests.
