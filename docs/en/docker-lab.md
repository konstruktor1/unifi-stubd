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

The entry point runs the outer matrix through the declared Build Runner Tier 2
instance described in [CI/CD](ci-cd.md). Its validation requires Docker API
access to the selected isolated lab daemon:

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

Core owns product qualification on the assigned Linux ARM64 Test Runner machine
through `make qualify-candidate` and `make qualify-candidate-matrix`. The
[qualification and controller contract](../../lab/stub/TIER3-CONTROLLER.md)
defines the required native context, explicit data locator, candidate/SSH inputs,
release binding, handshake and cleanup. Core uses public Test Runner lifecycle
commands through its existing native Build Runner binding, then runs the bounded
guest fixture and strict SSH inform tunnel. Labs owns controller preparation,
fresh per-lane provenance and the authoritative machine definitions. This lane
proves the exact guest executable hash and live controller behavior; a
Stable-only result leaves RC and combined acceptance open.

For the separate UniFi OS controller integration, the API helper accepts
`--controller-mode unifi-os`, `--controller-url` and `--auth-file`. The origin
must be reachable from the helper and contain no credentials, path, query or
fragment. The private regular JSON file contains only `username` and `password`;
keep it outside Git with mode `0600`. This mode uses the UniFi OS login,
session cookie and rotating CSRF headers, then routes Network requests through
`/proxy/network`. Controller setup belongs to the companion; the helper refuses
its `bootstrap` action in this mode.

The Make handshake supports schema 1 for the classic disposable controller and
schema 2 for a Labs-owned UniFi OS controller with Core-owned API/browser
helpers. The 2026-10-26 assigned-machine Stable 11.0.97 result, including visible port
metrics and cleanup, is recorded in [STUB-02](backlog.md). RC remains unqualified;
the schema support alone does not establish acceptance for another controller
instance or channel.

The lab proves container-local Linux bridge/FDB observation and the resulting
controller-visible product behavior. It does not prove a physical topology,
Proxmox host bridge, FreeBSD/OPNsense runtime, or deployed controller. Those
remain separately owned machine-backed compatibility tests.
