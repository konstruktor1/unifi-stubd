# Testing Guide

All Go-dependent checks run through the configured Build Runner. Host Go and
host-installed Go tools are not supported.

## Candidate checks

Run `make check`, `make race`, and the required package target through Build
Runner. The reduced Go suite protects independent wire vectors, bounded
adoption and host-safety behavior, and the local metrics mapping; it does not
recreate a controller journey.

## Product acceptance

For any adoption, metrics, or controller-compatibility change, run:

```sh
BUILD_RUNNER_DIR=/path/to/build-runner make integration-nonunifi-metrics
```

The gate resolves the latest published Stable and Release Candidate at run
start and runs the same product image in both lanes. It proves model `US8`,
adoption, stable connection, observed source link metadata, advancing real
counters, positive rates, and complete cleanup. A fixed controller version
cannot create a release claim.

The manual `Controller Integration` workflow invokes the same target for a
trusted ref and uploads no captures or controller state.

## External compatibility

Firmware research, QEMU/VM runtimes, physical topologies, real hosts, and
deployed controllers belong to `unifi-stubd-labs`. Their evidence does not
expand the supported core product scope without a separate product decision.
