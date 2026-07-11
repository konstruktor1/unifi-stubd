# Prioritized Backlog

Last updated: 2026-07-11.

This backlog contains the next concrete product work for `unifi-stubd`. The
roadmap remains the long-term feature list. Items here are ordered by delivery
priority and should preserve the core safety boundary: controller input must
not mutate host networking, services, users, packages, routes, or firewalls.

Effort uses relative sizes: `S` for a focused change, `M` for a multi-module
change, and `L` for work that includes controller integration and fixtures.

| ID | Priority | Work item | Acceptance criteria | Effort |
| --- | --- | --- | --- | --- |
| B01 | P0 | **Completed 2026-07-10:** redact adoption SSH logs | Adoption SSH logs use fixed command labels. A real SSH regression test proves keys and passwords stay out of logs, responses, and authentication errors while the key remains persisted; the existing status regression remains green. | S |
| B02 | P0 | **Completed 2026-07-10:** update the Go build line | GitHub and Docker builds use the current Go 1.25 patch without a committed patch or `toolchain` pin. `make check`, the race detector, `make vulncheck`, and a package build pass on the current Go 1.25 release. | S |
| B03 | P1 | **Completed 2026-07-10:** refresh interface observations correctly | Every runtime port build starts from unchanged operator configuration and reads a fresh MAC, IPv4, IPv6, link-state, speed, media, and counter snapshot. Explicit values retain priority, and complete zero-counter resets reach the payload. | M |
| B04 | P1 | **Completed 2026-07-10:** harden persistence and process shutdown | `SIGTERM` and `SIGINT` stop the daemon cleanly and release runtime resources. Adoption and status files use same-directory temporary files, file and directory sync, and atomic rename. Interrupted-write regressions pass. | M |
| B05 | P1 | **Completed 2026-07-10:** restore a real controller integration gate | The pinned Docker controller gate bootstraps empty volumes and verifies discovery, Pending Adoption, sanitized `setparam`, persisted auth state, at least two AES-GCM informs, and three stable connected samples for switch and gateway fixtures. | L |
| B06 | P1 | **Completed 2026-07-10:** complete the safe controller-command lifecycle | Only an argument-free `wg-genkeypair` with a request ID is handled. The X25519 key pair and request ID remain in memory, the correlated result is retried until one successful inform, and logs/status expose only sanitized counters. The pinned controller gate proves exactly one handled command, one delivered result, no duplicate, no pending result, and no repeating task loop after a stability window. | M/L |
| B07 | P2 | **Completed 2026-07-10:** add heartbeat and failure policy | Positive controller interval hints drive the dynamic timer after clamping to operator-configured bounds. Transport failures double the effective delay with bounded symmetric jitter; a completed HTTP exchange resets the failure sequence. Unit and live-daemon tests cover clamping, cap, jitter, reset, status, and invalid configuration. The pinned controller gate proves its returned positive hint becomes the bounded effective and next interval with zero pending failures or backoff. | M |
| B08 | P2 | **Completed 2026-07-10:** repair the coverage target | `make coverage` instruments all production `internal/...` packages from the external `tests/...` suites, reports the aggregate result, and enforces an initial 60 percent minimum locally and in CI. | S |
| B09 | P2 | **Completed 2026-07-11:** make profile maturity evidence-based | Experimental gateway profiles report `experimental`. `tested` requires at least one concrete entry in `validated_controller_versions`; `us8` and `uxg-lite` record their 10.3.58 and 10.4.57 controller evidence without promoting the gateway profile. | S |
| B10 | P2 | **Completed 2026-07-11:** restore release hygiene | `VERSION`, changelog, and install docs agree on 0.2.4-alpha. Tag builds attach packages, checksums, and a persistent SPDX SBOM to the GitHub pre-release. The release checklist covers signing and repository metadata. | M |

## Delivery Order

1. Security blockers B01 and B02 were completed on 2026-07-10.
2. Runtime correctness: B03 and B04 were completed on 2026-07-10.
3. Controller compatibility: B05, B06, and B07 were completed on 2026-07-10.
4. Quality and release gates: B08, B09, and B10 are complete.

## Alpha Release Gate

Do not create another alpha tag until B01 through B06 are complete and these
commands pass on the release candidate:

```sh
make check
go test -race ./...
make vulncheck
make integration-docker
make package
```

## Beta Entry Criteria

- The controller gate passes against at least two pinned UniFi Network versions.
- Linux and FreeBSD package smoke tests pass, including upgrade and rollback.
- No controller command remains in a repeating task loop.
- Dynamic interface changes are reflected without restarting the daemon.
