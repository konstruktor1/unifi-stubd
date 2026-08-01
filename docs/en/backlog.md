# Product Backlog

Last updated: 2026-08-01.

The backlog is the only owner of current priority and next work. It is ordered
by usable product outcome; runners, fixtures, receipts, and test catalogues are
delivery mechanisms, not backlog outcomes.

| Order | Product outcome | Status | Done when |
| ---: | --- | --- | --- |
| 1 | Adopt the non-UniFi `US8` and show its real metrics. | In progress | The same candidate remains connected and visibly presents the named source's real link, speed, and bidirectional traffic on dynamically resolved Stable and RC. |
| 2 | Keep the host outside controller control. | Open | Adoption retains only Stub identity, and controller input cannot mutate the host. |
| 3 | Deliver the supported Linux installation. | Open | The accepted Stub installs and starts in `bridge-observe` mode with a named real source on one supported Linux host. |
| 4 | Publish the operator path. | After 1–3 | The quick start covers installation, configuration, adoption, visible metrics, limitations, and recovery without lab internals. |
| 5 | Add another representation only for a concrete need. | Future | A named operator outcome, local read-only source, safety boundary, and separate Stable/RC acceptance exist before implementation. |

Current status: The Stable product result is complete. On the current Release
Candidate, the Stub is visible, adopted, connected, and supplies real metrics,
but the controller device page is unavailable. The operator-visible RC result
therefore remains open.

## Next step

Complete the missing operator-visible result on the current Release Candidate
and confirm the current Stable result for the same product candidate. If the
Release Candidate continues to accept, connect, and report the Stub while its
device page remains unavailable, record an RC compatibility limitation. Do not
change the Stub without evidence of incorrect product behavior. Phase 2 begins
only after Phase 1 is complete.
