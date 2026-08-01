# Product Requirements

Last updated: 2026-08-01.

## Product goal

`unifi-stubd` lets a non-UniFi Linux host or bridge appear in UniFi Network as
a small switch. The first supported product slice is an `US8`: an operator can
add it to the controller, adopt it, keep it connected, and see real link and
traffic metrics from the named local interface.

Lab infrastructure, runners, fixtures, internal status schemas, and test
catalogues are delivery mechanisms. They are not product requirements.

## Contract

| ID | Requirement |
| --- | --- |
| `PR-STUB` | One local Linux bridge running the Stub appears as `US8`, can be adopted, and remains connected. |
| `PR-METRICS` | UniFi Network presents real link state, observed speed, and bidirectional traffic from the configured local source. The Stub reads that source without modifying the host and never fabricates missing values. |
| `PR-ADOPTION` | Adoption persists only the bounded state needed by the Stub. Controller input must not provision or otherwise mutate the host. |
| `PR-COMPAT` | The same product candidate works with the latest published Stable and latest published Release Candidate; no controller version is permanently locked. |
| `PR-SCOPE` | The accepted slice is `US8` on a local Linux bridge. Other profiles, gateways, operating systems, and physical topologies require a separate product decision and evidence. |

## Acceptance conditions

The product candidate is accepted only when the same candidate shows all of
the following on both current controller channels:

1. The US8 is visible, can be adopted, and remains connected.
2. An operator can see its real link state, observed speed, and bidirectional
   traffic from the named local source.
3. Missing observations remain absent instead of becoming synthetic values.
4. Controller input cannot provision, restart, upgrade, or execute commands on
   the host.

The operator-visible result is required; controller-internal data alone is not
acceptance. Exact pixel placement is not a product requirement. Packaging and
installed-service acceptance belong to the usable Linux release.
