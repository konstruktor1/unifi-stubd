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
stable project name `unifi-stubd` and the prepared consumer runtime entry:

```sh
PROJECT_NAME=unifi-stubd
export BUILD_RUNNER_CONTROLLER_IMAGE=local/build-runner-controller:unifi-stubd-go
./build-runner/runtime/entry build-runner --project-name "$PROJECT_NAME" run --tier 1 --service go -- make check
./build-runner/runtime/entry build-runner --project-name "$PROJECT_NAME" run --tier 1 --service go -- make lint
./build-runner/runtime/entry build-runner --project-name "$PROJECT_NAME" run --tier 1 --service go -- make test
./build-runner/runtime/entry build-runner --project-name "$PROJECT_NAME" run --tier 1 --service go -- make package
```

The controller-facing Stable/RC gate has one runner-aware consumer entry point. On the host,
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

Before planning, delegation, or work that may use Build Runner, primary and
delegated agents resolve the provider root through the reviewed consumer-local
runtime entry and invoke direct `bin/provider status build-runner --json` from
the physical consumer root. Treat its lifecycle result as authoritative, read
its ordered context, current contract, operational guide, and named release
note from the reported provider commits, and use only the primary for
`provider update build-runner` after that review.
Before planning, delegation, or work that may use Code Checker, use the
reviewed stable `code-checker` command with this exact consumer root and run
project-bound release status first. Treat a successful response from that
reviewed command as authoritative. Bind its reported provider checkout,
launcher, exact consumer project, and separate consumer Git root; read the
current contract, guide, and named release note from the reported commit.
Do not implement, copy, or test another status or route validator, pin an
expected Code Checker version or commit, or repeat provider-owned schema,
canonicality, launcher, boundary, or adoption consistency checks.
Never discover, construct, repair, or replace either provider route. Apply
current Consumer requirements and every matching Direct-Upgrade Ledger row.
Reread after provider path, version, commit, backend, or adoption state
changes, keep both providers read-only, and stop provider-dependent work if
required material is unavailable.
Delegates report only readiness or concrete blockers; only the primary agent
confirms or acknowledges a release after the complete consumer review.

Build Runner release adoption is automatic; project-bound release status is
the current identity authority.

Build Runner and Code Checker are independent providers of equal standing for
this consumer. Build Runner supplies execution environments; Code Checker
supplies code checking and Anti-Bloat. An unqualified request to check the
provider means checking both provider relationships independently. Bind Code
Checker only from its authoritative project-bound release status and keep its
configuration and results consumer-owned. Neither provider adopts, discovers,
or replaces the other.
Passing the read-only Code Checker checkout into a Build Runner artifact with
`BUILD_RUNNER_HOST_PATH_CODE_CHECKER_SOURCE` is transport only; its value must
be the provider checkout reported by Code Checker status. It does not merge
provider identity, authority, or ownership.
Consumer Make targets receive the four reported route fields as
`CODE_CHECKER_PROVIDER_CHECKOUT`, `CODE_CHECKER_PROVIDER_LAUNCHER`,
`CODE_CHECKER_CONSUMER_PROJECT`, and `CODE_CHECKER_CONSUMER_GIT_ROOT`; they do
not infer replacements.

Code Checker's function IDs are `code-checker-check`, `code-checker-antibloat`, and
`code-checker-i18n`; this consumer does not subscribe to the optional i18n
function.
`code-checker.toml` subscribes this consumer to Anti-Bloat/Anti-Slop.
`make code-checker` is the single complete pre-merge gate. It runs the
consumer's product and security checks through Build Runner, then the Go-,
JavaScript-, and Bash-aware Code Checker component;
`make anti-bloat-changed` and `make anti-bloat-full` issue the canonical work
orders; the changed scope requires the accepted task base through
`CODE_CHECKER_ANTI_BLOAT_BASE`. A work order alone is not cleanup or quality
evidence: the agent must resolve the consumer-owned scope, apply the policy,
and run the one canonical consumer gate named by the work order.

The direct QEMU manager in the `unifi-stubd-labs` companion repository remains
an external-machine compatibility path for test/lab lifecycle. Its wrapper
must warn before every action that the profile has no Provider consumer
attribution, recommend the optional Test Runner conversion procedure, and
leave execution and ownership unchanged. Reevaluate conversion on every use;
never infer or apply adoption without a truthful observed source, a compatible
blueprint/service, reviewed parity, and separate authority.

The direct FreeBSD packaging builder is also an external-machine compatibility
path without Provider consumer attribution. It cannot supply Go build or
release evidence until an authorized VM Runner or reviewed split-build route
owns its compilation step.

`make check` is the runner-local product subgate. It verifies the
golangci-lint configuration, runs lint, enforces repository policy, validates
packaged configuration/profile YAML, and runs `go test ./...`; it is not the
complete pre-merge gate by itself.

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
