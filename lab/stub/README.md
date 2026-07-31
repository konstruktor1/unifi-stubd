# US8 Controller Product Gate

This disposable lab answers one product question: can the current
`unifi-stubd` binary make a non-UniFi Linux bridge appear as an `US8`, become
adopted, stay connected, and expose real metrics in UniFi Network?

Run the complete acceptance only through the project Build Runner:

```sh
BUILD_RUNNER_DIR=/path/to/build-runner make integration-nonunifi-metrics
```

The matrix resolves the latest published Stable and Release Candidate at run
start and executes the same product image in both lanes. There is no
fixed-version acceptance fallback.

Each lane must show through the controller API that:

- the actual controller version is the resolved channel version;
- the device is model `US8`, is visible before adoption, is adopted, and is
  connected in both metric reads;
- `uplink0` is up and its controller-visible speed equals the locally observed
  source speed;
- bounded bidirectional traffic makes both root counters advance between
  controller reads and exposes positive current uplink and port rates.

The lab has one product service, `stub-bridge-observe`. UniFi Network, MongoDB,
and the controller API reader are provisioning. The gate connects the product
directly to the controller and does not mirror product payloads or status JSON
in assertion helpers.

Every lane starts with empty controller and database volumes and removes its
containers, networks, volumes, orphans, and owned image aliases on exit.
Evidence below ignored `dist/controller-channel-matrix/` binds the product
bytes, dynamically resolved controller versions, result, and cleanup.

The fixture proves container-local Linux observation, not a physical topology,
Proxmox host, FreeBSD/OPNsense runtime, or deployed controller. Those are
separately owned compatibility tests. Never commit captures, credentials,
adoption keys, private URLs, real MAC tables, or client data.
