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
- `lab/`: lab switch identities, command snippets, and payload fixtures.
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

Use the repository targets instead of inventing ad-hoc commands:

```sh
make check
make lint
make test
make package
```

`make check` is the normal pre-merge command. It verifies the golangci-lint
configuration, runs lint, enforces repository policy, and runs `go test ./...`.

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
