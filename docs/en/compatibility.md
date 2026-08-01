# Controller Compatibility

Compatibility is evaluated against the latest published UniFi Network Stable
and latest published Release Candidate resolved at run start. Both lanes must
execute the same product bytes. Recorded controller versions are evidence for
that run, never product pins.

## Accepted runs

No run currently qualifies as durable product acceptance.

The repository contains no retained executable record that binds one exact
candidate to the required Stable/RC product result and focused host-safety
checks. The current Release Candidate accepts, connects, and receives real
metrics from the Stub, but its device page is unavailable. This prevents a
complete operator-visible result; it is not evidence of incorrect Stub
behavior.

## Required run record

An accepted row must link durable evidence for:

| Identity | Controller result | Product result |
| --- | --- | --- |
| Exact source and product-binary identity | Resolved Stable and RC versions, with the same bytes used in both | Adoption, connected state, real visible source metrics, and focused host-safety checks |

The complete acceptance conditions are owned by the
[product requirements](requirements.md). Diagnostic and superseded runs remain
lab history and do not appear as compatibility claims.
Package installation is a separate Phase 2 release result.
