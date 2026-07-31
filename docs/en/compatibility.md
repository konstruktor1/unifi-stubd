# Controller Compatibility

Compatibility is evaluated against the latest published UniFi Network Stable
and latest published Release Candidate resolved at run start. Both lanes must
execute the same product bytes. Recorded controller versions are evidence for
that run, never product pins.

## Accepted runs

No run currently qualifies as durable product acceptance.

The repository contains no retained executable record that binds one exact
candidate to all required controller data-path, operator-visible UI, host-
safety, package, and installed-service outcomes. Session IDs, generated
`KEY=value` receipts, hashes repeated only in prose, and an API result without
the UI observation do not establish those outcomes.

## Required run record

An accepted row must link durable evidence for:

| Identity | Controller result | Product result |
| --- | --- | --- |
| Exact source commit and product-binary digest | Resolved Stable and RC versions, with the same bytes used in both | Visibility, adoption, connected state, real source metrics, direct UI observation, focused host-safety checks, and installed-package startup |

The complete acceptance conditions are owned by the
[product requirements](requirements.md). Diagnostic and superseded runs remain
lab history and do not appear as compatibility claims.
