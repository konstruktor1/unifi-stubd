# Development Workflow

This project uses a staged GitHub Flow. Changes stay small and reviewable, `dev`
is the integration branch, `main` is the tested release base, and packages are
published only from tags, GitHub releases, or an explicit package workflow run.

The model fits the existing repository infrastructure:

- GitHub Actions `CI` runs `make check`, `make coverage`, `make vulncheck`, and
  SBOM generation entirely on GitHub-hosted Linux runners.
- On `main` pushes, `CI` also builds packages and installs the generated Debian
  package once in the Ubuntu runner as a neutral smoke test.
- `Package Repositories` builds Linux packages and the release SBOM on Ubuntu,
  deploys the Linux-only GitHub Pages site, and creates or updates the GitHub
  pre-release for a `v*` tag.
- `Tier 3 FreeBSD Packaging` is a separate manual workflow. It uses the
  self-hosted FreeBSD builder path only when explicitly dispatched and never
  participates in `CI` or the automatic package workflow.
- The `github-pages` environment already allows deployments only from `main`
  and `v*` tags.

## Go Toolchain

Both `go.mod` and `go.work` declare the supported Go `1.25` minor line. Patch
versions and a separate `toolchain` directive are intentionally not pinned.

All GitHub Actions jobs read `go.mod` through `actions/setup-go` with
`check-latest: true`, so GitHub resolves the newest available Go `1.25` patch.
The Docker build follows the same minor line through `golang:1.25-bookworm`.

## Branch Stages

| Stage | Branch | Purpose | Automatic checks | Output |
| --- | --- | --- | --- | --- |
| Change | short-lived `codex/*`, `feat/*`, or `fix/*` branch | One focused change | `CI / check` on pull request | No package publication |
| Integration | `dev` | Batch reviewed work before release promotion | `CI / check` on pull request and push | No package publication |
| Release base | `main` | Stable, tested source for tags and public packages | `CI / check` plus package build and Debian install smoke on push | Package artifacts uploaded to the run |
| Package repository | `v*` tag, GitHub pre-release, or manual workflow from `main` | Publish the standard package set | Linux package matrix and release SBOM | Linux-only GitHub Pages and, for a tag, a GitHub pre-release |
| FreeBSD extension | Manual Tier 3 dispatch | Build or publish FreeBSD/OPNsense package evidence | Self-hosted FreeBSD builder path | Run evidence only, or FreeBSD release assets and combined Pages |

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
3. For controller-facing changes, run the standardized Docker gate from
   [Docker Controller Lab](docker-lab.md) on the exact `dev` commit being
   promoted and record the evidence.
4. Review the diff as a release-candidate change set, not as a single feature.
5. Merge to `main`.
6. The `main` push runs `CI / check`; the package job then builds Linux package
   formats and installs the generated Debian package once in GitHub Actions.
7. Do not deploy package repositories from `dev`.

Direct pushes to `main` should be limited to explicit emergency or automation
cases. The normal route is pull request review into `main`.

## Release And Package Publishing

Version numbers come from `v*` tags or an explicit package workflow input. Do
keep `VERSION`, the changelog section, and the release tag aligned.

The normal alpha release path is:

```sh
git switch main
git pull --ff-only origin main
git tag -a v0.2.5-alpha -m "unifi-stubd v0.2.5-alpha"
git push origin v0.2.5-alpha
```

Publishing a `v*` tag starts `Package Repositories`. Linux packages, the SPDX
SBOM, the Linux-only Pages site, and the GitHub pre-release are published first.
Manual Linux package and Pages builds can run from `main`:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.5-alpha \
  -f package_release=1
```

If `version` is omitted in a manual run, the workflow reads `VERSION`.
This manual event does not create or update a GitHub release. Rerun a failed
tag-triggered workflow run when the pre-release itself must be repaired.

FreeBSD packaging is an independent Tier 3 operation. The `publish` input
defaults to `false`; that mode builds the selected ref and retains only
seven-day Actions artifacts and the job summary as evidence:

```sh
gh workflow run tier3-freebsd.yml --ref main -f publish=false
```

To extend an existing matching `v*` pre-release and deploy combined Linux and
FreeBSD Pages, dispatch the same tag with explicit publication enabled:

```sh
gh workflow run tier3-freebsd.yml --ref <v-tag-with-tier3-workflow> -f publish=true
```

The publishing mode adds the FreeBSD tarballs, native `pkg` repository archive,
and `SHA256SUMS.freebsd`. It does not replace the automatic Linux
`SHA256SUMS`. Tier 3 is never a trigger, `needs` dependency, or required result
of `CI` or `Package Repositories`.
Legacy tags that do not contain the Tier 3 workflow, or whose `SHA256SUMS`
already includes FreeBSD artifacts, are not eligible for this split extension.

## Hotfixes

1. Branch from `main`.
2. Open a pull request to `main`.
3. Run the same `main` gates.
4. Tag or publish only after the `main` CI run is green.
5. Bring the hotfix back to `dev` before normal development continues.

## Gate Selection

| Change type | Required local gate | Extra gate |
| --- | --- | --- |
| Go code, config schema, profile data | `make check`, `make coverage`, `git diff --check` | Targeted `go test ./tests/...` when useful |
| Inform, adoption, controller payload, profile rendering | `make check` | Standardized Docker gate with `make integration-docker` before `dev` to `main` |
| Packaged config, service files, package metadata | `make check`, `make package` | GitHub `main` package install smoke |
| FreeBSD or OPNsense runtime behavior | `make check` | Manual `Tier 3 FreeBSD Packaging` run, plus FreeBSD/OPNsense smoke with temporary state when behavior changed |
| Release notes, package publication | `make check` | Linux-first `Package Repositories`; optional manual Tier 3 publication only after the matching pre-release exists |

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
- `tier3-freebsd` environment: require approval, allow only `main` and `v*`,
  and grant no publication permissions to the self-hosted build job.
- Runner roles: remove `docker-dmz` and Docker daemon access from the old
  FreeBSD proxy. Bare-host passwordless sudo is forbidden. A disposable,
  containerized proxy may retain container-local passwordless sudo only when
  the protected `tier3-freebsd` environment explicitly sets
  `ALLOW_CONTAINER_LOCAL_SUDO=true` and the workflow verifies its container
  marker, an enforced seccomp filter, a private container root mount, the same
  root filesystem for PID 1, and the absence of dangerous root capabilities,
  exposed block devices, host-root mounts, and host runtime sockets. Tier 2
  requires a separate Docker runner; no runner carries both Tier 2 and Tier 3
  role labels.

The package job is intentionally post-merge on `main`, because it validates the
exact committed release base and uploads package artifacts for that run.

## References

- GitHub Flow: https://docs.github.com/en/get-started/using-github/github-flow
- GitHub Actions events: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
- Protected branches and required checks: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- Deployment environments: https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments
- Short-lived branches in trunk-based development: https://trunkbaseddevelopment.com/short-lived-feature-branches/
