# Controller Compatibility

The current product candidate is controller-data-qualified for the defined
`US8` Linux-bridge slice. Run `matrix-20260729102907-1` passed that data-path
gate on 2026-07-29. Direct operator-visible UI evidence is still open, so this
is not yet full product acceptance.

Every compatibility claim resolves the latest published Stable and latest
published Release Candidate at run start and executes the same product bytes in
both lanes. Recorded versions are run evidence, never product pins.

## Current controller-data evidence

| Recorded run | Controller channels | Qualified product identity | Observed product outcome |
| --- | --- | --- | --- |
| `matrix-20260729102907-1`, 2026-07-29 | Stable `10.5.67`, RC `10.5.66` | `0.2.5-alpha`, Linux/arm64; image `sha256:e32bbf6f114f2d02d91c4cecc4d592f6708f7683cd5cf97776a312101792a9e5`; binary `sha256:e6e6868577061e8fdd6fe52ccd5a5181d084a5de8f363acd66b321249713afa1` | The same `US8` product image was visible before adoption, was adopted, remained connected, and returned `uplink0` up at the locally observed 10,000 Mb/s, positive uplink/port RX and TX rates, and advancing root counters through controller data reads in both lanes. Cleanup passed. UI result: not tested. |

The run binds product source manifest
`sha256:8f086331798ebe872c48def723d52b8d5ec325727ba5a6a5216d3c59969e4477`
and gate manifest
`sha256:7a644c182680eb2efe510c734f0e2eae6b02860c530e9d08236329902b314c94`.
The latter contains 17 explicit executed gate inputs; captures, generated
evidence, and unrelated local files are excluded. The resolved controller
runtime base is
`lscr.io/linuxserver/unifi-network-application@sha256:cebacc22bd529909a85efd91003c20c1d77bab7d79850689e2c0da720677bf4c`;
the Stable and RC controller image IDs are respectively
`sha256:c926c226dd434cafb7ee932946afc0fa79069c6d9fa54f2ebbe289aac47e5084`
and `sha256:926afa0f0c5d06fca4f6fad77f18e766415462f757beab23504878e4dcfce45b`.
It ran as `make integration-nonunifi-metrics` through Build Runner `6.9.1` at
commit `0e52df6641a073a93fd225b23c6164cc75050a33`. The runner job used Go
`1.26.5` on Linux/arm64; the exported product binary records its actual
compiler as Go `1.25.12`. The outer invocation was observed as session
`a837a9247a466ddcaf2388e7c4346a71` in the provider's expiring session cache.
The job did not receive that outer ID or project name, so `final.env` records
both as `unknown` rather than fabricating a durable binding.

The qualified binary was exported only after both lanes and cleanup passed. The
run-bound Linux package set was built from that file rather than from a second
compile. Independent extraction confirmed the same binary SHA-256 in the raw
file, TGZ, and DEB. Both extracted package configs explicitly select
`profile: us8`. The complete DEB/RPM/Arch/TGZ set was published atomically in
Build Runner session `a353f7d825e2405feb56ddf7ded528d4`; all four checksums,
the TGZ/DEB binary identity, and both US8 configs passed in session
`db226b891c0a0e7f984da5358d987914`. RPM and Arch were produced from the same
SHA-checked staging input but their embedded binaries were not independently
extracted in that runner. This proves artifact lineage; it does not prove
rendered UI values or installed-service behavior. Those remain separate
acceptance steps.

## Superseded evidence

Earlier July 2026 runs are diagnostic history only. They lacked at least one
of the current source, binary, toolchain, runtime-digest, or dynamic-channel
bindings and therefore make no present compatibility claim.

Firmware, gateway, physical-topology, and private-controller research belongs
to the companion lab record and does not expand this product matrix. A later
compatibility claim must repeat the dynamic Stable/RC data-path gate and direct
UI observation for its exact product bytes; the versions above remain evidence
for this run, not pins.
