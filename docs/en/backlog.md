# Product Backlog

Last updated: 2026-08-01.

The backlog is the only owner of current priority and next work. It is ordered
by usable product outcome; runners, fixtures, receipts, and test catalogues are
delivery mechanisms, not backlog outcomes.

| Order | Product outcome | Status | Done when |
| ---: | --- | --- | --- |
| 1 | Adopt the non-UniFi `US8` and show its real metrics. | Blocked upstream | The same candidate remains connected and visibly presents the named source's real link, speed, and bidirectional traffic on dynamically resolved Stable and RC. |
| 2 | Keep the host outside controller control. | Open | Adoption retains only Stub identity, and controller input cannot mutate the host. |
| 3 | Deliver the supported Linux installation. | Open | The accepted Stub installs and starts in `bridge-observe` mode with a named real source on one supported Linux host. |
| 4 | Publish the operator path. | After 1–3 | The quick start covers installation, configuration, adoption, visible metrics, limitations, and recovery without lab internals. |
| 5 | Add another representation only for a concrete need. | Future | A named operator outcome, local read-only source, safety boundary, and separate Stable/RC acceptance exist before implementation. |

Current status: Stable has the complete operator-visible result. The current
RC UI returns 404 for a required asset; no incorrect Stub behavior was observed,
and the strict RC result remains blocked upstream. See [Compatibility](compatibility.md).

## Next step

Rerun the unmodified Stable/RC product gate when Ubiquiti publishes a new RC.
Until then, complete the focused host-safety proof for outcome 2. Do not patch
the controller artifact, weaken the visible-metrics requirement, or change the
Stub without evidence of incorrect product behavior. Installation work begins
only after outcomes 1 and 2 are complete.
