# Development Workflow

This project uses a staged GitHub Flow. Changes stay small and reviewable, `dev`
is the integration branch, `main` is the tested release base, and packages are
published only from tags, GitHub releases, or an explicit package workflow run.

The model fits the existing repository infrastructure:

- GitHub Actions `CI` runs `make check`, `make vulncheck`, SBOM generation, and
  FreeBSD cross-build checks.
- On `main` pushes, `CI` also builds packages and installs the generated Debian
  package once in the Ubuntu runner as a neutral smoke test.
- `Package Repositories` builds Linux packages, FreeBSD tarballs, repository
  metadata, and native FreeBSD pkg repositories, then deploys GitHub Pages.
- The `github-pages` environment already allows deployments only from `main`
  and `v*` tags.

## Branch Stages

| Stage | Branch | Purpose | Automatic checks | Output |
| --- | --- | --- | --- | --- |
| Change | short-lived `codex/*`, `feat/*`, or `fix/*` branch | One focused change | `CI / check` on pull request | No package publication |
| Integration | `dev` | Batch reviewed work before release promotion | `CI / check` on pull request and push | No package publication |
| Release base | `main` | Stable, tested source for tags and public packages | `CI / check` plus package build and Debian install smoke on push | Package artifacts uploaded to the run |
| Package repository | `v*` tag, GitHub pre-release, or manual workflow from `main` | Published package repository set | Package matrix plus FreeBSD pkg repo build | GitHub Pages package repositories |

Short-lived branches should normally live only a few days. If a change grows
large, split it into smaller pull requests that can each pass the gates on its
own.

## Normal Development

1. Start from current `dev`.

   ```sh
   git switch dev
   git pull --ff-only origin dev
   git switch -c codex/<topic>
   ```

2. Implement the change locally and run the matching gate.

   ```sh
   make check
   git diff --check
   ```

3. Push the topic branch and open a pull request to `dev`.
4. Merge to `dev` only after the pull request has the required checks and review.
5. Keep `dev` green. Fix broken `dev` immediately before adding unrelated work.

## Promotion To Main

1. Open a pull request from `dev` to `main`.
2. The pull request must pass `CI / check`.
3. Review the diff as a release-candidate change set, not as a single feature.
4. Merge to `main`.
5. The `main` push runs `CI / check`, then the package job builds all package
   formats and installs the generated Debian package once in GitHub Actions.
6. Do not deploy package repositories from `dev`.

Direct pushes to `main` should be limited to explicit emergency or automation
cases. The normal route is pull request review into `main`.

## Release And Package Publishing

Version numbers come from `v*` tags or an explicit package workflow input. Do
not rely on the Makefile default for release builds.

The normal alpha release path is:

```sh
git switch main
git pull --ff-only origin main
git tag -a v0.1.9-alpha -m "unifi-stubd v0.1.9-alpha"
git push origin v0.1.9-alpha
```

Publishing a `v*` tag or GitHub pre-release starts `Package Repositories`.
Manual retries run from `main`:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.1.9-alpha \
  -f package_release=1
```

If `version` is omitted in a manual run, the workflow resolves the latest
reachable `v[0-9]*` tag and strips the leading `v`.

## Hotfixes

1. Branch from `main`.
2. Open a pull request to `main`.
3. Run the same `main` gates.
4. Tag or publish only after the `main` CI run is green.
5. Bring the hotfix back to `dev` before normal development continues.

## Gate Selection

| Change type | Required local gate | Extra gate |
| --- | --- | --- |
| Go code, config schema, profile data | `make check`, `git diff --check` | Targeted `go test ./tests/...` when useful |
| Inform, adoption, controller payload, profile rendering | `make check` | `make integration-docker` |
| Packaged config, service files, package metadata | `make check`, `make package` | GitHub `main` package install smoke |
| FreeBSD or OPNsense runtime behavior | `make check` | FreeBSD/OPNsense smoke with temporary state only |
| Release notes, package publication | `make check` | `Package Repositories` workflow from tag, release, or `main` dispatch |

Target-host package installation is not a default development gate. Use it only
when explicitly testing a rollout, and keep host-specific configs outside this
repository.

## Recommended GitHub Controls

The repository currently uses a ruleset for the default branch that blocks
deletion and non-fast-forward updates. Keep that, and use these controls as the
target policy:

- `main`: require pull requests, require `CI / check`, block deletion, block
  non-fast-forward updates, and keep the Pages environment limited to `main` and
  `v*` tags.
- `dev`: require pull requests and `CI / check` once the branch is used as the
  integration stage.
- `github-pages` environment: allow deployment from `main` and `v*` only.

The package job is intentionally post-merge on `main`, because it validates the
exact committed release base and uploads package artifacts for that run.

## References

- GitHub Flow: https://docs.github.com/en/get-started/using-github/github-flow
- GitHub Actions events: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
- Protected branches and required checks: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- Deployment environments: https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments
- Short-lived branches in trunk-based development: https://trunkbaseddevelopment.com/short-lived-feature-branches/
