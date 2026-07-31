# Product Requirements

Last updated: 2026-07-31.

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
| `PR-METRICS` | The Stub sends real read-only link state, speed, cumulative RX/TX counters, and current bidirectional rates from the configured local source to the controller. UniFi Network must present the resulting link, speed, and traffic values to the operator; raw cumulative counters need not be rendered verbatim. Missing data stays absent and is never replaced with synthetic values. |
| `PR-ADOPTION` | Adoption persists only the bounded state needed by the Stub. Controller input must not provision or otherwise mutate the host. |
| `PR-COMPAT` | Product acceptance resolves the latest published Stable and latest published Release Candidate at run start. The same product bytes must pass both; no controller version is permanently locked. |
| `PR-SCOPE` | The accepted slice is `US8` on a local Linux bridge. Other profiles, gateways, operating systems, and physical topologies require a separate product decision and evidence. |

## Acceptance conditions

The product candidate is accepted only when both current controller channels
show all of the following for the same binary:

1. The controller reports the dynamically resolved version.
2. The US8 is visible before adoption and accepts controller-triggered adoption.
3. Both controller metric reads report the US8 adopted and connected.
4. The selected uplink identifies the named local source, is up, and reports
   the speed observed from that source.
5. Both root counters advance between controller reads during bounded
   bidirectional traffic and current uplink/port rates are positive.
6. In both controller channels, an operator can see the resulting link, speed,
   and traffic values on the US8 device page. The observation must distinguish
   real values from an empty field, placeholder, or stale synthetic value.

Once per candidate, focused executable checks must also show that adoption
persists only the bounded Stub identity and that controller input cannot execute
host provisioning, shell, restart, or upgrade actions. These are candidate
conditions, not duplicated assertions in each controller lane.

Controller API readback proves the data path but not its presentation. A direct
operator-visible observation proves presentation; both are required. Exact
pixel placement is not a product requirement.

## Acceptance evidence

Four results have distinct meanings:

- controller data-path evidence proves visibility, adoption, connection, and
  real metric values on current Stable and Release Candidate;
- UI evidence proves that an operator can actually see those values;
- focused safety checks prove bounded persistence and the absence of
  controller-triggered host actions;
- artifact binding proves that a package contains the accepted binary.

None substitutes for the others. In particular, package creation does not
prove that the installed service and operator instructions work on a supported
host.
