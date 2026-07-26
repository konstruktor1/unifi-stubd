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
BUILD_RUNNER_DIR=${BUILD_RUNNER_DIR:?set BUILD_RUNNER_DIR to the provider checkout}
PROJECT_NAME=unifi-stubd
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make check
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make lint
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make test
"$BUILD_RUNNER_DIR/build-runner" --project-name "$PROJECT_NAME" run --tier 1 --service go -- make package
```

The Docker-lab Make target is a runner-aware consumer entry point. On the
host, `make integration-docker` delegates to the committed Build Runner
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

Before planning, delegating, or executing Build Runner work, every primary or
delegated agent must read the current provider contract at
`$BUILD_RUNNER_DIR/snippets/AGENTS.build-runner.md`. Provider documentation and
release notes referenced there are part of that contract; remembered or copied
instructions are not a substitute. Until the exact provider version and commit
are acknowledged, reread the current contract immediately before every runner
invocation and warn the user that the provider identity is pending. Treat the
provider checkout as read-only from this consumer project.

The approved provider identity for this checkout is Build Runner `6.2.1` at
commit `260efe778819a906eade3f1afe3b73c7282da89b`. Read the regular checked-out
provider contract; the former 6.1.3 Git-object reading exception is revoked.
Do not modify the provider checkout.

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

## Coding Rules

- Follow the existing Go style and package boundaries.
- Keep exported Go identifiers documented in English.
- Keep user-facing docs in English and German when the change affects users.
- Keep tests under `tests/`, grouped by package behavior.
- Keep generated artifacts under `dist/` out of commits.
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
