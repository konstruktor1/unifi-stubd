# Contributing

Thanks for helping make `unifi-stubd` easier to run and reason about.

## Development Loop

The repository commits both `go.mod` and `go.work`:

- `go.mod` sets the minimum supported Go minor version.
- `go.work` selects the repository workspace and uses the same Go minor version.
- Build tools are declared in `go.mod` and executed with `go tool`.

```sh
make check
make package
```

Keep Go tests under `tests/`. Production packages under `internal/` should not
contain `_test.go` files.

Use the staged GitHub Flow described in
[`docs/en/development.md`](docs/en/development.md):

- short-lived topic branches target `dev`;
- `dev` is the integration branch and must stay green;
- `main` is the tested release base;
- package repositories are published only from `v*` tags, GitHub releases, or
  an explicit `Package Repositories` workflow run from `main`.

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
