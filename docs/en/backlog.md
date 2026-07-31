# Product Backlog

Last updated: 2026-07-29.

The backlog is ordered by usable product outcome. Provisioning and validation
mechanics are not separate product work.

| Order | Outcome | Status | Done when |
| ---: | --- | --- | --- |
| 1 | Prove the non-UniFi `US8` data path with real metrics. | Verified | The current compatibility evidence shows the same candidate visible, adopted, connected, and supplying advancing real metric values on current Stable and RC. |
| 2 | Prove that an operator sees those metrics in UniFi Network. | Next | The US8 device page on current Stable and RC visibly shows link, observed speed, and bidirectional traffic values from the real source. API readback alone is insufficient. |
| 3 | Preserve the host safety boundary. | Implemented; evidence open | Focused tests cover bounded Stub-owned adoption state and rejected shell, networking, restart, upgrade, and provisioning actions. A retained run must bind those checks to the exact qualified candidate source. |
| 4 | Deliver an installable Linux package for the supported slice. | In validation | The package contains the data-path-qualified binary. Its starter config selects `US8` explicitly, but a supported bridge configuration with a named real metric source must still be installed and started successfully on one Linux host. |
| 5 | Make the supported operator path usable. | After 2 and 4 | The quick start leads from installation and local metric source through adoption, visible controller metrics, limitations, and recovery without lab internals. |
| 6 | Add another representation only for a concrete product need. | Future | A named user outcome, local read-only data source, safety boundary, and separate Stable/RC acceptance exist before implementation. |

## Next step

Finish items 2, 3, and 4. First complete the supported US8 bridge configuration
with its named real local metric source; the explicit packaged US8 identity
alone is not that configuration. Install the run-bound package on one supported
Linux host, start it with the supported configuration, and confirm in both
current controller channels that an operator sees the real metrics. Retain a
focused safety result bound to that exact candidate source. Turn that exact
path into the quick start. No new metric family, gateway behavior, or profile
precedes this usable release.
