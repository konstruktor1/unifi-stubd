# Agent Instructions

## Project

`unifi-stubd` is a Go lab tool that makes a Linux host, VM, or bridge appear as
a minimal UniFi switch to a UniFi Network Controller. It is experimental,
unofficial, and intended only for isolated lab or management networks.

The safety boundary is important: the controller must not blindly provision the
host. Adoption commands can be accepted and persisted, but controller-triggered
restart, upgrade, shell, or host-networking changes must stay explicit,
lab-scoped, and reviewable.

## First Places To Read

- `README.md`: user-facing overview, quick start, services, packaging.
- `packaging/linux/etc/unifi-stubd/config.yaml`: packaged Linux service config.
- `lab/stub/`: project-owned controller integration fixtures.
- `docs/en/architecture.md`: architecture and component boundaries.
- `docs/en/protocol-notes.md`: discovery, inform, and adoption protocol notes.
- `CREDITS.md`: research sources, license boundaries, and attribution rules.
- `SECURITY.md`: lab-data and vulnerability reporting policy.

## Known Agent Support

`AGENTS.md` is the canonical instruction file. Compatibility bridge files are
thin adapters for common tools and should not duplicate project rules:

- Claude Code: `CLAUDE.md`
- Gemini CLI: `GEMINI.md`
- GitHub Copilot: `.github/copilot-instructions.md` and `.github/instructions/`
- Cursor: `.cursor/rules/unifi-stubd.mdc`
- Windsurf: `.windsurf/rules/unifi-stubd.md`
- Cline: `.clinerules/unifi-stubd.md`
- Roo Code legacy setups: `.roo/rules/unifi-stubd.md`
- Aider: `.aider.conf.yml` and `CONVENTIONS.md`

## Build And Validation

All Go-dependent work must run through the Build Runner. Treat local Go and
locally installed Go tools as unavailable: do not probe for, install, restore,
or invoke them, including cached `golang.org/toolchain` binaries. Use the
stable project name `unifi-stubd` and the configured provider checkout:

```sh
. ./build-runner/activate.local.sh
PROJECT_NAME=unifi-stubd
export BUILD_RUNNER_CONTROLLER_IMAGE=local/build-runner-controller:unifi-stubd-go
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make check
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make lint
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make test
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make package
```

The product gate has one runner-aware consumer entry point. On the host,
`make integration-nonunifi-metrics` delegates to the committed Build Runner
instance manifest; inside the canonical runner workspace it executes the lab
payload. Do not invoke
`lab/stub/scripts/run-docker-integration.sh` directly on the host. The payload
owns one disposable Compose project per gate and must remove all of its
containers, networks, named volumes, and orphans before returning; do not add
fixed container names, persistent restart policies, or a keep-running bypass.

This applies equally to direct `go`, `gofmt`, `golangci-lint`, `gopls`, and
`nFPM` use. Do not fall back to host tooling when the runner is unavailable;
report the runner blocker instead. Run the Build Runner from this durable
consumer checkout. `/tmp` and `/private/tmp` must not be used as its consumer
project, source tree, workspace, manifest store, artifact store, or evidence
store.

Before planning, delegation, or work that may use either provider, primary and
delegated agents use the reviewed consumer-local activation, run each
provider's project-bound release status, and read its exact current consumer
contract, operational guide, and named release note from the reported provider
commit. Apply current Consumer requirements and every matching Direct-Upgrade
Ledger row. Reread after provider path, version, commit, backend, or adoption
state changes, keep both providers read-only, and stop provider-dependent work
if required material is unavailable.
Delegates report only readiness or concrete blockers; only the primary agent
confirms or acknowledges a release after the complete consumer review.

The adopted Build Runner identity is `6.13.0` at commit
`e2e5907da0d5bd4be6b56aca24698547decf8406`; release adoption is automatic.

Build Runner and Code Checker are independent providers of equal standing for
this consumer. Build Runner supplies execution environments; Code Checker
supplies code checking and Anti-Bloat. An unqualified request to check the
provider means checking both provider relationships independently. Bind Code
Checker explicitly with `CODE_CHECKER_DIR`, follow its own usage instructions,
and keep its configuration and results consumer-owned. Neither provider adopts,
discovers, or replaces the other.
Passing the read-only Code Checker checkout into a Build Runner artifact with
`BUILD_RUNNER_HOST_PATH_CODE_CHECKER_SOURCE="$CODE_CHECKER_DIR"` is transport
only; it does not merge provider identity, authority, or ownership.

The adopted Code Checker provider identity is `0.8.0` at commit
`a30d6a688fcfbb381e0081599902f904e2666507`; its release selection is automatic.
Its function IDs are `code-checker-check`, `code-checker-antibloat`, and
`code-checker-i18n`; this consumer does not subscribe to the optional i18n
function.
`code-checker.toml` subscribes this consumer to Anti-Bloat/Anti-Slop.
`make code-checker` runs the Go-, JavaScript-, and Bash-aware gate through the
Build Runner environment;
`make anti-bloat-changed` and `make anti-bloat-full` request the canonical
scope instruction. A plan alone is not cleanup or quality evidence: the agent
must resolve the consumer-owned scope, apply the policy, and run executable
product gates.

The direct QEMU manager in the `unifi-stubd-labs` companion repository remains
an external-machine compatibility path for test/lab lifecycle. Its wrapper
must warn before every action that the profile has no Provider consumer
attribution, recommend the optional Test Runner conversion procedure, and
leave execution and ownership unchanged. Reevaluate conversion on every use;
never infer or apply adoption without a truthful observed source, a compatible
blueprint/service, reviewed parity, and separate authority.

`make check` is the normal pre-merge command. It verifies the golangci-lint
configuration, runs lint, enforces repository policy, validates packaged
configuration/profile YAML, and runs `go test ./...`.

The repository uses:

- Go modules and a root `go.work`.
- Go `1.25` as the minimum minor-version floor.
- Go tool directives for `golangci-lint` and `nFPM`.
- YAML service configuration under `/etc/unifi-stubd/config.yaml`.
- Packaged Linux file sources under `packaging/linux/`.

## Repository Layout

- `cmd/unifi-stubd/`: command entry point and CLI wiring.
- `internal/config/`: YAML config loading and CLI override model.
- `internal/device/`: fake switch profiles, link speed, and payload data.
- `internal/discovery/`: UniFi discovery TLV packet building.
- `internal/inform/`: inform packet framing, padding, and client logic.
- `internal/adoption/`: adoption state and environment persistence.
- `internal/adoptionssh/`: built-in SSH shim for advanced adoption.
- `internal/adapters/linuxbridge/`: Linux bridge/FDB helpers.
- `tests/`: all Go tests. Do not add `_test.go` files under `internal/`.
- `packaging/linux/`: Linux service units and packaged config sources.
- `packaging/nfpm.yaml`: Debian, RPM, and Arch Linux package metadata.
- `scripts/`: policy and packaging scripts.
- `docs/en/` and `docs/de/`: user-facing docs in English and German.
- `lab/stub/`: the only lab runtime kept in this core repository; firmware
  research, QEMU/VM runtimes, and other Tier 3 assets belong to the companion
  project `unifi-stubd-labs`.

## Companion Data And Tier 3 Test Machines

- On this workstation, the project-owned companion data root is
  `/Volumes/ssd/unifi-stubd-data/`. Treat it as part of the overall
  `unifi-stubd` project context even though its contents stay outside this core
  repository. The volume name is only a filesystem path; do not use "SSD" as
  project terminology.
- `/Volumes/ssd/unifi-stubd-data/tier3/machines/` contains the binding test
  machines for controller-backed product tests. Use these test hosts when a
  test requires a controller; do not substitute a core `lab/stub` fixture, a
  generic runner host, or a production controller.
- `/Volumes/ssd/unifi-stubd-data/tier3/artifacts/` and
  `/Volumes/ssd/unifi-stubd-data/tier3/migration/` contain machine artifacts
  and migrations. They are not the authoritative test-machine inventory.
- Before Tier 3 work, read the companion runner documentation and the machine
  definitions under `tier3/machines/`; do not reconstruct their roles from
  memory.

## Coding Rules

- Follow the existing Go style and package boundaries.
- Keep exported Go identifiers documented in English.
- Keep user-facing docs in English and German when the change affects users.
- Keep tests under `tests/`, grouped by package behavior.
- Keep generated artifacts under `dist/` out of commits.
- Keep firmware images, VM disks, extracted firmware trees, and Tier 3 runtime
  state out of this core repository. They belong to `unifi-stubd-labs` and its
  separately managed data storage.
- Do not commit PCAPs, adoption keys, controller API tokens, private controller
  URLs, real lab addresses, SSH host keys, MAC tables, or client data.
- Use documentation example addresses such as `192.0.2.0/24` instead of real
  site addresses.

## Protocol And Research Boundaries

`CREDITS.md` lists public docs and reverse-engineering projects used for
research. Some research repositories do not publish a license. Treat those as
idea and protocol references only; do not copy source code from them.

If a future change copies code or structured data from any external project,
update `CREDITS.md`, `NOTICE.md`, and the license decision before merging.

## Safe Behavior Expectations

- Discovery and inform traffic should remain opt-in and lab-scoped.
- Adoption SSH behavior should emulate only the minimal commands needed for
  advanced adoption.
- Do not execute arbitrary shell commands received from a controller.
- Do not mutate host networking based on controller provisioning data unless a
  future design explicitly adds a reviewed, local-only adapter.
- Prefer deterministic profile data over controller-derived guesses.
