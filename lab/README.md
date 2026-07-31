# Controller Product Gate

The core repository keeps one disposable controller fixture under `lab/stub/`.
It proves the current product outcome for a non-UniFi Linux bridge represented
as an `US8`: discovery, adoption, stable connection, and real controller-visible
traffic metrics from the named local source.

Run it only through the Build Runner entry point:

```sh
BUILD_RUNNER_DIR=/path/to/build-runner make integration-nonunifi-metrics
```

The gate resolves the latest published Stable and Release Candidate at run
start and uses the same product image in both lanes. It does not expose a
fixed-version acceptance path.

The fixture contains one product service, `stub-bridge-observe`; the controller,
database, and API reader are provisioning. Every lane uses
empty state and removes its containers, networks, volumes, orphans, and owned
image aliases on exit.

Firmware research, gateway simulations, QEMU/VM runtimes, and machine-backed
compatibility tests belong to `unifi-stubd-labs`. Never commit captures,
controller credentials, adoption keys, private URLs, real MAC tables, or client
data here.
