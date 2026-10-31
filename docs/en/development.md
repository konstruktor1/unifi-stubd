# Development Workflow

This project uses a staged GitHub Flow. Changes stay small and reviewable, `dev`
is the integration branch, `main` is the tested release base, and packages are
published only from tags, GitHub releases, or an explicit package workflow run.

The workflow topology separates CI, Linux package publication, and the manual
FreeBSD extension. General checks and Linux package builds use GitHub-hosted
Ubuntu workers with Build Runner's ephemeral containers; package installation
smoke also uses a disposable GitHub-hosted worker. The existing self-hosted
runner serves only manual FreeBSD packaging. Provider-source preparation and
unresolved handoffs are maintained in
[CI/CD](ci-cd.md#open-provider-handoffs).

## Go Toolchain

Both `go.mod` and `go.work` declare `go 1.27.0`, the current stable minor series.
The separate linter module uses the same floor. `golang.org/x/crypto v0.56.0`
fixes the reported SSH vulnerabilities.
Build Runner selects the executing compiler independently; the module and
workspace have no `toolchain` directive.

Agent-run Go commands and Go-based Make targets use the configured Build
Runner. The prepared consumer runtime owns the invocation; host Go is not a
fallback, including in checked-in workflows.

## Branch Stages

| Stage | Branch | Purpose | Intended automatic checks | Output |
| --- | --- | --- | --- | --- |
| Change | short-lived `codex/*`, `feat/*`, or `fix/*` repository branch | One focused change | `CI / check` on branch push | No package publication |
| Integration | `dev` | Batch reviewed work before release promotion | `CI / check` on branch push | No package publication |
| Release base | `main` | Stable, tested source for tags and public packages | `CI / check` plus package build and Debian install smoke on push | Package artifacts uploaded to the run |
| Package repository | `v*` tag, GitHub pre-release, or manual workflow from `main` | Publish the standard package set | Linux package matrix and release SBOM | Linux-only GitHub Pages and, for a tag, a GitHub pre-release |
| FreeBSD extension | Manual `FreeBSD Packaging` dispatch | Compatibility artifacts pending an approved build route | Containerized self-hosted runner and direct FreeBSD builder | Non-evidentiary run artifacts; publication remains blocked |

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

2. Implement the change and inspect the patch itself on the host.

   ```sh
   git diff --check
   ```

3. Run the complete `make verify` gate from the prepared consumer and any
   additional product acceptance required for the change. Record a blocked
   check with its owner and next public command; it does not satisfy the merge
   requirement. Push the topic branch and open a pull request to `dev`.
4. Merge to `dev` only after the pull request has the required checks and review.
5. Keep `dev` green. Fix broken `dev` immediately before adding unrelated work.

## Promotion To Main

1. Open a pull request from `dev` to `main`.
2. The pull request must pass `CI / check` with both provider bindings prepared
   on the CI worker. Missing runtime preparation is a failure, not evidence.
3. For controller-facing changes, run the standardized Docker gate from
   [Docker Controller Lab](docker-lab.md) on the exact `dev` commit being
   promoted and record the evidence.
4. Review the diff as a release-candidate change set, not as a single feature.
5. Merge to `main`.
6. After the approved route is installed, the `main` push runs `CI / check` and
   the package job builds the Linux formats and performs the Debian install smoke.
7. Do not deploy package repositories from `dev`.

Direct pushes to `main` should be limited to explicit emergency or automation
cases. The normal route is pull request review into `main`.

## Release And Package Publishing

The [release checklist](release.md) owns tags, manual workflow inputs, the
FreeBSD extension, checksums, signing, and Pages publication. This page owns
branch promotion and gate selection only.

## Hotfixes

1. Branch from `main`.
2. Open a pull request to `main`.
3. Run the same `main` gates.
4. Tag or publish only after the exact tree passes the complete gate; the
   current `main` CI run is not a substitute.
5. Bring the hotfix back to `dev` before normal development continues.

## Gate Selection

| Change type | Required product check | Additional product or delivery evidence |
| --- | --- | --- |
| Go code, config schema, profile data | `make verify`, after host-side `git diff --check` | Any additional product acceptance required by the change |
| US8 inform, adoption, or controller-visible metrics | `make verify` | Stable/RC controller acceptance; resolve the [runtime handoff](ci-cd.md#open-provider-handoffs) first |
| Packaged config, service files, package metadata | `make verify` on the host; `make package` inside Build Runner | `main` package install smoke on the disposable worker |
| FreeBSD or OPNsense runtime behavior | `make verify` | Authorized FreeBSD build route plus a Tier 3 runtime smoke; the direct host-Go builder is compatibility-only and blocked as evidence |
| Release notes, package publication | `make verify` | Linux-first `Package Repositories`; FreeBSD publication is unavailable |

Target-host package installation is not a default development gate. Use it only
when explicitly testing a rollout, and keep host-specific configs outside this
repository.

## Recommended GitHub Controls

Configure and verify these controls in GitHub; repository files alone do not
establish their active settings:

- `main`: require pull requests, require `CI / check`, block deletion, block
  non-fast-forward updates, and keep the Pages environment limited to `main` and
  `v*` tags.
- `dev`: require pull requests and `CI / check` once the branch is used as the
  integration stage.
- `github-pages` environment: allow deployment from `main` and `v*` only.
- `freebsd-packaging` environment: require approval, allow only `main` and `v*`,
  and grant no publication permissions to the self-hosted build job.
- Runner roles and FreeBSD proxy-safety controls are enforced by the
  workflows and the [release checklist](release.md); they are not duplicated
  here.

## References

- GitHub Flow: https://docs.github.com/en/get-started/using-github/github-flow
- GitHub Actions events: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
- Protected branches and required checks: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- Deployment environments: https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments
- Short-lived branches in trunk-based development: https://trunkbaseddevelopment.com/short-lived-feature-branches/
