# Controller Compatibility

Compatibility is evaluated against the latest published UniFi Network Stable
and latest published Release Candidate resolved at run start. Both lanes must
execute the same product bytes. Recorded controller versions are evidence for
that run, never product pins.

## Accepted runs

No run currently qualifies as durable product acceptance.

The repository contains no retained executable record that binds one exact
candidate to the required Stable/RC product result and focused host-safety
checks.

## Diagnostic current-channel finding

On 2026-08-01, the same Stub candidate passed the complete visible product path
on Stable 10.5.67. RC 10.5.66 accepted, adopted, connected, and received real
10-GbE and bidirectional traffic metrics. Its UI requested
`swai.b54a...js`; a direct controller request to that path returned 404. The RC
result is therefore blocked at the controller UI, while no incorrect Stub
behavior was observed. The upstream context is recorded in the
[10.5.66 release discussion](https://community.ui.com/releases/UniFi-Network-Application-10-5-66/9fd6a940-8f43-4071-a36a-28edcc7e5032)
and [10.5.67 release](https://community.ui.com/releases/UniFi-Network-Application-10-5-67/375288b9-a4b4-46f1-a19d-5c787d342c2b).

## Required run record

An accepted row must link durable evidence for:

| Identity | Controller result | Product result |
| --- | --- | --- |
| Exact source and product-binary identity | Resolved Stable and RC versions, with the same bytes used in both | Adoption, connected state, real visible source metrics, and focused host-safety checks |

The complete acceptance conditions are owned by the
[product requirements](requirements.md). Diagnostic and superseded runs remain
lab history and do not appear as compatibility claims.
Package installation is a separate Phase 2 release result.
