# US8 Controller Product Gate

This fixture implements the acceptance described in
[Docker Controller Lab](../../docs/en/docker-lab.md). It tests one product
service, `stub-bridge-observe`; UniFi Network, MongoDB, and its API and browser
helpers are provisioning only.

Run it only through the project Build Runner entry point:

```sh
. ./build-runner/activate.local.sh
make integration-nonunifi-metrics
```

The matrix resolves current Stable and Release Candidate versions, runs the
same product image in both, and removes every disposable resource. It proves
container-local Linux behavior only; physical and machine-backed compatibility
belongs to `unifi-stubd-labs`.
