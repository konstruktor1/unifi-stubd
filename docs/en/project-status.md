# Project Status

Last updated: 2026-07-29.

## Product outcome

`unifi-stubd` is intended to let an operator add a non-UniFi Linux host or
bridge to UniFi Network as a small switch and see real read-only link and
traffic metrics from a named local interface.

The first supported slice is deliberately narrow: one `US8` representation on
a local Linux bridge. Gateway identities, other switch models, FreeBSD, and
other physical topologies are implemented or researched in places, but they
are not accepted product scope without a separate decision and evidence.

## Evidence status

| Product condition | Status | Evidence or gap |
| --- | --- | --- |
| Visible before adoption, adoptable, and connected | Verified | The same candidate passed current Stable and Release Candidate. |
| Real link, observed speed, counters, and rates reach the controller | Verified | Controller data reads showed the named source, positive rates, and advancing counters in both channels. |
| Operator sees those values on the US8 device page | Open | No direct UI observation has yet been retained. API readback does not prove presentation. |
| Adoption cannot provision or execute actions on the host | Evidence open | Focused tests cover bounded persistence and rejected shell, networking, restart, upgrade, and provisioning actions, but the current product evidence does not bind a safety run to the exact qualified source. |
| Linux package contains the qualified binary | Verified | The run-bound package set is built from the exported candidate rather than a second compile. |
| Installed package starts in the supported US8/bridge/metric configuration | Open | The packaged starter now selects `US8` explicitly, but remains in non-observing `stub` mode without a named real metric source. |
| Operator quick start reaches the visible-metric outcome | Open | Installation, supported configuration, UI confirmation, limits, and recovery still need one coherent path. |

Exact run IDs, controller versions, and binary identity belong in the
[compatibility evidence](compatibility.md), not in this status summary.

## Product boundary

The supported behavior must remain read-only toward the represented host.
Controller input may establish the Stub's bounded adoption state, but it must
not execute arbitrary shell commands or change host networking, services,
firmware, routes, firewall, users, or time.

The Stub exposes only values observed from the configured local source.
Missing observations stay missing rather than becoming synthetic success
values. Discovery and inform traffic remain opt-in and suitable only for an
isolated lab or management network.

## Next product work

1. Complete the supported US8 bridge configuration with a named real local
   metric source; the explicit packaged US8 identity alone is not acceptance
   evidence.
2. Install the run-bound package on one supported Linux host and start it with
   that configuration.
3. Adopt the same candidate on the dynamically resolved Stable and Release
   Candidate and directly confirm that an operator sees real link, speed, and
   bidirectional traffic values.
4. Retain the focused host-safety result for that exact candidate source.
5. Turn that exact installation, configuration, adoption, limitation, and
   recovery path into the quick start.

Only after this path is accepted is the first usable release complete. New
metric families, gateway behavior, and additional profiles remain later
product decisions.

## Evidence hygiene

Firmware images, captures, adoption keys, controller tokens, private
controller URLs, SSH host keys, MAC tables, and client data stay out of Git.
Labs, runners, packages, and tests support delivery; they are not separate
product features or roadmap phases.
