# Controller Product Gate

The core repository keeps one disposable controller fixture under `lab/stub/`.
It tests the same non-UniFi `US8` product image against the latest published
Stable and Release Candidate. See [stub/README.md](stub/README.md) for the
product boundary, command, assertions, and cleanup contract.

Firmware research, gateway simulations, QEMU/VM runtimes, and machine-backed
compatibility tests belong to `unifi-stubd-labs`. Never commit captures,
controller credentials, adoption keys, private URLs, real MAC tables, or client
data here.
