# Contributing

Thanks for helping make `unifi-stubd` easier to run and reason about.

## Development Loop

Use the Build Runner contract and commands in [`AGENTS.md`](AGENTS.md); host Go
and direct Go-based Make targets are not fallbacks. The branch, gate, and
promotion workflow is owned by
[`docs/en/development.md`](docs/en/development.md).

Keep Go tests under `tests/`. Production packages under `internal/` should not
contain `_test.go` files.

## Pull Requests

- Keep changes focused.
- Open normal feature and refactor pull requests against `dev`.
- Promote `dev` to `main` with a pull request after `dev` is green.
- Use English comments and public-facing documentation.
- Add or update tests when behavior changes.
- Do not commit generated packages from `dist/`.
- Do not commit PCAPs, lab controller addresses, adoption keys, API keys, or
  private MAC tables.

## Docs

When adding user-facing documentation, update both language trees when useful:

- `docs/en/`
- `docs/de/`

## Safety Model

The project must not execute arbitrary controller commands or mutate host
networking from controller provisioning data. Stub behavior should stay explicit
and lab-scoped.
