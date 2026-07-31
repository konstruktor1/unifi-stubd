# CI/CD, Build, Packaging, and Deployment Catalog

This catalog records the repository build and publishing chain. It is based on
the checked-in workflows, Makefile targets, scripts, package metadata, and
release documentation. Items that are not defined in the repository are marked
as open instead of inferred.

## Current Architecture

GitHub Actions orchestrates the pipeline. Project build logic lives in
`Makefile`, `scripts/package.sh`, `scripts/package-freebsd-pkg-repos.sh`, and
`scripts/build-package-repos.sh`. Go tooling is versioned through `go.mod`
tool directives and run via `go tool`. The root `Dockerfile` is a runtime and
lab image path; it is not a CI build container and no GHCR publish workflow is
defined.

The pipeline is split into:

- `CI`: validation, vulnerability scan, SBOM, and a `main`-only Linux package
  smoke test, entirely on GitHub-hosted Ubuntu.
- `Controller Integration`: manually dispatched controller-adoption gate on
  the trusted Docker lab runner, independent of `CI`.
- `Package Repositories`: release-package and persistent SBOM build, GitHub
  pre-release asset publication, and Pages deployment for the Linux-first
  unsigned alpha package repositories.
- `FreeBSD Packaging`: manually dispatched FreeBSD package build. It can
  either keep seven-day run artifacts only or extend an existing `v*`
  pre-release and rebuild Pages with the FreeBSD outputs. This packaging
  pipeline is not a Tier 1, Tier 2, or Tier 3 test classification.

## Build Targets

| Target | Toolchain | Entry point | Output | OS/arch | Build environment | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `make build` | Current GitHub-supported Go 1.25 patch | `go build ./...` | Local binaries in Go build cache | Host OS/arch | Local or GitHub Ubuntu | Defined, not uploaded |
| `make check` | Go 1.25, golangci-lint, shell policy | `lint`, `validate-config`, `test` | Pass/fail gate | Host OS/arch | GitHub Ubuntu and local | Primary pre-merge gate |
| `make coverage` | Go 1.25 coverage instrumentation | External tests under `tests/...` with all production `internal/...` packages instrumented | Diagnostic `dist/cover.out` and `dist/coverage.html`; no numeric merge threshold | Host OS/arch | GitHub Ubuntu and local | CI diagnostic report |
| `make vulncheck` | Go 1.25 `govulncheck` tool | `govulncheck ./...` | Pass/fail gate | Host OS/arch | GitHub Ubuntu | CI security gate |
| `make build-freebsd` | Go 1.25 cross-compile | `cmd/unifi-stubd` | `dist/unifi-stubd_freebsd_<arch>` | FreeBSD `PKG_FREEBSD_GOARCH` | Local only | Kept as local diagnostic target; not release CI |
| `make package` | Go 1.25 plus nFPM | `scripts/package.sh` | `.deb`, `.rpm`, Arch package, Linux `.tar.gz` | Linux `amd64` or `arm64` | GitHub Ubuntu | Release package path |
| `make package-deb` | Go plus nFPM | `scripts/package.sh deb` | Debian package | Linux current `PKG_GOARCH` | Local or GitHub Ubuntu | Defined |
| `make package-rpm` | Go plus nFPM | `scripts/package.sh rpm` | RPM package | Linux current `PKG_GOARCH` | Local or GitHub Ubuntu | Defined |
| `make package-arch` | Go plus nFPM | `scripts/package.sh archlinux` | Arch Linux package | Linux current `PKG_GOARCH` | Local or GitHub Ubuntu | Defined |
| `make package-tgz` | Go plus tar | `scripts/package.sh tgz` | OS-specific tarball | `PKG_GOOS`/`PKG_GOARCH` | Local | Defined |
| `make package-freebsd-tgz` | Go cross-compile plus tar | `scripts/package.sh tgz` | FreeBSD tarball | FreeBSD `amd64` or `arm64` | Local diagnostic path | Defined, not used by release workflow |
| `make package-freebsd-pkg-repos` | Go on FreeBSD builder, FreeBSD `pkg`, tar | `scripts/package-freebsd-pkg-repos.sh` | Native FreeBSD `pkg` repos and FreeBSD tarballs | `FreeBSD:14/15` for `amd64`, `aarch64`, `armv7`; tarballs for `amd64`, `arm64` by default | Manual containerized GitHub Actions runner; optional mapped `jexec` jails | FreeBSD packaging path |
| `make package-repos` | dpkg, createrepo-c, repo-add, tar | `scripts/build-package-repos.sh` | Static package repository site | Repository metadata for available packaged targets | GitHub Ubuntu | Linux-first or combined GitHub Pages input |
| `make integration-nonunifi-metrics` | Build Runner controller instance with Docker CLI/Compose and typed Docker API access | Resolve current Stable and RC, then run the same product image in both disposable lanes | US8 adoption, connection, and real-metrics acceptance | Linux container lab | Build Runner locally or by manual GitHub Actions dispatch | Defined in the standalone controller workflow |
| Root `Dockerfile` | Docker BuildKit, Go, Alpine | `Dockerfile` | Product runtime image containing only the daemon | Linux multi-arch capable | Local/manual | No publish target defined |

Open build target definitions:

- No macOS or Windows build target is defined.
- No top-level `tools/` directory exists; core integration helpers live under
  `lab/stub/tools/`. Tier 3 tooling belongs to `unifi-stubd-labs`.
- No CI container image or GitHub Container Registry publishing path is defined.

## GitHub Actions

### `CI`

| Area | Current value |
| --- | --- |
| File | `.github/workflows/ci.yml` |
| Triggers | `workflow_dispatch`, pushes to `dev` and `main`, `pull_request` |
| Default permissions | `contents: read` |
| Jobs | `check`, `package` |
| Runners | `ubuntu-latest` only |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `anchore/sbom-action@v0`, `actions/upload-artifact@v7` |
| Cache | No explicit cache. `setup-go` defaults apply where provided by the action. |
| Artifacts | `coverage`, `sbom`; `packages` on a `main` push, all retained for seven days |
| Security boundary | No self-hosted runner and no FreeBSD builder access. |

`check` runs `make check`, `make coverage`, `make vulncheck`, generates `dist/sbom.spdx.json`,
uploads the coverage and SBOM outputs for seven days, and reads the Go version
from `go.mod`.

`package` runs only on pushes to `main` after `check`. It reads the centralized
`VERSION`, builds Linux packages, installs the generated Debian package once,
validates the installed config, and uploads `dist/packages/` for seven days.

### `Controller Integration`

| Area | Current value |
| --- | --- |
| File | `.github/workflows/tier2-docker.yml` |
| Trigger | `workflow_dispatch` only |
| Permissions | `contents: read` |
| Job | `controller-adoption` |
| Runner | `[self-hosted, Linux, X64, docker-dmz, docker-engine, unifi-stubd]` |
| Command | `make integration-nonunifi-metrics` |
| Execution boundary | The Make target delegates exclusively to the committed Build Runner instance; no workflow-local Go or Docker toolchain |
| Artifacts | None; captures and controller state are never uploaded |
| Concurrency | One `unifi-stubd-tier2-docker` group, no cancel-in-progress |

The workflow preflight requires the configured Build Runner provider. The
committed consumer instance validates Go, Docker client/server, Compose, and
Buildx before it executes the lab payload. The workflow resolves the latest
published Stable and Release Candidate at run start and has no dependency on
or trigger from the standard `CI`, package
workflows, pull requests, pushes, or schedules.

### `FreeBSD Packaging`

| Area | Current value |
| --- | --- |
| File | `.github/workflows/freebsd-packaging.yml` |
| Trigger | `workflow_dispatch` only, with required boolean `publish` input |
| Default permissions | `contents: read` |
| Runner | `[self-hosted, Linux, X64, docker-dmz, unifi-stubd, freebsd-pkg-builder]` for the build only |
| Build command | `make package-freebsd-pkg-repos` |
| Build-only mode | `publish=false`: one seven-day `freebsd-release-<sha>` run artifact containing the two tarballs, repository archive, and `SHA256SUMS.freebsd` |
| Publication mode | `publish=true`: only on an existing matching `v*` pre-release; uploads FreeBSD assets and `SHA256SUMS.freebsd`, then rebuilds combined Pages |
| Concurrency | One `unifi-stubd-freebsd-packaging` build group; Pages publication also uses the shared `github-pages` group |

The Ubuntu validation job resolves `VERSION`. Every request must use `main` or
the matching `v*` tag. A publication request must use that tag, which must
already have a pre-release
containing the expected Linux packages, persistent SBOM, and `SHA256SUMS`. It
refuses to mutate a final release, an incomplete pre-release, a release with
signed checksum assets, or a legacy `SHA256SUMS` that already covers FreeBSD.
The selected release tag must itself contain this workflow. The containerized
self-hosted GitHub Actions runner runs on `docker.dmz.lan`; `docker-dmz`
identifies that location and does not grant Docker-daemon capability. Its
preflight verifies Linux/X64, SSH/SCP/tar, absence of Docker-daemon access, and
strict SSH/SCP access to the non-root FreeBSD build-host account with native
`pkg`, `tar`, and Go tools before starting the build. It rejects passwordless
sudo on a bare-host proxy. For a disposable, containerized proxy,
container-local passwordless sudo is accepted only when the protected
`freebsd-packaging` environment explicitly sets
`ALLOW_CONTAINER_LOCAL_SUDO=true`. The preflight must recognize a container
marker, confirm an enforced seccomp filter and private container root mount,
and verify that PID 1's root either matches the container root when inspectable
or is inaccessible even to container root because of the sandbox. It must also
find no dangerous root capabilities, exposed block devices, host-root mounts,
or host runtime sockets.

FreeBSD packaging is not part of `CI`, the Tier 2 controller lab, the
automatic tag workflow, or the test-tier classification. A failed or omitted
FreeBSD run therefore does not block Linux packages, the Linux-first Pages
site, or creation of the pre-release.

### `Package Repositories`

| Area | Current value |
| --- | --- |
| File | `.github/workflows/package-pages.yml` |
| Triggers | `workflow_dispatch`, `v*` tag push |
| Default permissions | `contents: read` |
| Deploy permissions | `contents: read`, `pages: write`, `id-token: write` on `build-and-deploy` only |
| Jobs | `build-packages`, `build-sbom`, `build-and-deploy`, `publish-release` |
| Runners | `ubuntu-latest` only |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `anchore/sbom-action@v0`, `actions/upload-artifact@v7`, `actions/download-artifact@v7`, `actions/configure-pages@v6`, `actions/upload-pages-artifact@v5`, `actions/deploy-pages@v5` |
| Artifacts | Seven-day `linux-package-artifacts` and `release-sbom`, Pages artifact, persistent Linux/SBOM release assets |
| Environment | `github-pages` |
| Concurrency | One `github-pages` deployment group, no cancel-in-progress |

Linux packages are built on Ubuntu for `amd64` and `arm64`. The deploy job
builds and publishes a Linux-only `dist/package-site`. On a `v*` tag,
`publish-release` creates or updates the pre-release with Linux packages,
`unifi-stubd.spdx.json`, and `SHA256SUMS`. This workflow never selects a
self-hosted runner and does not wait for FreeBSD packaging.

## OS and Architecture Matrix

| Target | Build path | Test path | Packaging | Publishing |
| --- | --- | --- | --- | --- |
| Linux amd64 | `PKG_GOARCH=amd64 make package` | `make check`; Debian install smoke on `main` | Debian, RPM, Arch, tarball | GitHub Actions artifact and Pages package site |
| Linux arm64 | `PKG_GOARCH=arm64 make package` | `make check` only; no install smoke | Debian, RPM, Arch, tarball | Pages package site |
| FreeBSD 14 amd64 | Manual `FreeBSD Packaging` via FreeBSD builder | Build/package verification only; target-host smoke manual | Native `pkg` repo | Seven-day artifact, or existing pre-release and combined Pages with `publish=true` |
| FreeBSD 14 aarch64 | Same | Build/package verification only | Native `pkg` repo | Same |
| FreeBSD 14 armv7 | Same | Build/package verification only | Native `pkg` repo | Same |
| FreeBSD 15 amd64 | Same | Build/package verification only | Native `pkg` repo | Same |
| FreeBSD 15 aarch64 | Same | Build/package verification only | Native `pkg` repo | Same |
| FreeBSD 15 armv7 | Same | Build/package verification only | Native `pkg` repo | Same |
| FreeBSD/OPNsense tarball amd64 | Manual `FreeBSD Packaging` builder path | Manual Tier 3 target-host smoke | Tarball | Seven-day artifact, or existing pre-release and combined Pages with `publish=true` |
| FreeBSD/OPNsense tarball arm64 | Manual `FreeBSD Packaging` builder path | Manual Tier 3 target-host smoke | Tarball | Same |

Open gaps:

- Linux arm64 packages are built but not installation-smoked in CI.
- FreeBSD/OPNsense runtime behavior still requires manual target-host evidence.
- Stable signing is open; alpha repositories are intentionally unsigned.

## Buildhosts, Runners, and Containers

GitHub-hosted Ubuntu is used for generic validation, Linux packaging, repository
metadata, normal release publication, and Pages deployment. Normal CI and the
automatic tag workflow never select a self-hosted runner.

`docker-dmz` is a location label for containerized self-hosted GitHub Actions
runners on `docker.dmz.lan`; it does not itself mean that a runner can reach a
Docker daemon. The Tier 2 selector is `[self-hosted, Linux, X64, docker-dmz,
docker-engine, unifi-stubd]`. It runs only `lab/stub`; `docker-engine` is the
explicit capability label for Docker daemon access and Docker Compose v2.

The separate FreeBSD packaging runner container is selected by `[self-hosted,
Linux, X64, docker-dmz, unifi-stubd, freebsd-pkg-builder]`. It is a GitHub
Actions SSH/SCP proxy, not a test-tier role: it transfers source through strict
SSH/SCP to a non-root account on the FreeBSD build host and receives the
artifacts back. It does not need or receive the `docker-engine` label. Both
runner containers may share `docker.dmz.lan`; they remain separate
registrations with separate mounts and capabilities.

Passwordless sudo remains forbidden when the packaging proxy runs directly on
a host. It is permitted only inside a disposable, containerized proxy whose
workflow preflight verifies a recognized container marker, an enforced seccomp
filter, a private container root mount, and a PID 1 root that either matches
the container root when inspectable or is inaccessible even to container root
because of the sandbox. It must also verify that no dangerous root
capabilities, exposed block devices, host-root mounts, or host runtime sockets
exist. This exception requires the protected `freebsd-packaging` environment
variable `ALLOW_CONTAINER_LOCAL_SUDO=true`. Protect that environment with
required reviewers and deployment rules for `main` and `v*` only.

The BSD builder is expected to provide:

- SSH alias `unifi-stubd-freebsd-builder`, unless `FREEBSD_PKG_REMOTE` is
  overridden.
- A work directory from `FREEBSD_PKG_REMOTE_DIR`, defaulting to
  `/tmp/unifi-stubd-freebsd-pkg`.
- FreeBSD `pkg`, `tar`, and a Go command named `go`, `go125`, or `go126`.
- Optional `jexec` and `FREEBSD_PKG_BUILD_JAILS` only for custom builders with
  running jails. Poudriere target jails are not assumed to be running `jexec`
  build environments.

Containers are defined for runtime and lab use only:

- Root `Dockerfile`: builds only the daemon into the Alpine product image.
- Lab Dockerfiles build the controller API reader and dynamically resolved
  disposable controller images separately.
- No CI build container, no GHCR image, and no container package release are
  defined.

## Packaging and Publishing

| Format | Source | Build command | Name/output | Validation | Publish target |
| --- | --- | --- | --- | --- | --- |
| Debian | `packaging/nfpm.yaml`, Linux stage tree | `make package` or `make package-deb` | `unifi-stubd_<version>-<release>_<arch>.deb` | Installed and config-validated on `main` for amd64 | Artifact, APT repo |
| RPM | `packaging/nfpm.yaml`, Linux stage tree | `make package` or `make package-rpm` | `unifi-stubd-<version>-<release>.<arch>.rpm` | Repository metadata build | Artifact, RPM repo |
| Arch Linux | `packaging/nfpm.yaml`, Linux stage tree | `make package` or `make package-arch` | `unifi-stubd-<version>-<release>-<arch>.pkg.tar.zst` | Repository metadata build | Artifact, Arch repo |
| Linux tarball | Linux stage tree | `make package` or `make package-tgz` | `unifi-stubd_<version>-<release>_linux_<arch>.tar.gz` | Checksums | Artifact, Pages checksum |
| FreeBSD tarball | FreeBSD stage tree from BSD-built binary | `make package-freebsd-pkg-repos` | `unifi-stubd_<version>-<release>_freebsd_<arch>.tar.gz` | `SHA256SUMS.freebsd` when published | Manual FreeBSD packaging artifact; optional pre-release and combined Pages |
| FreeBSD `pkg` | FreeBSD stage tree and UCL manifest | `make package-freebsd-pkg-repos` | `repo/FreeBSD:*/*` including `packagesite.pkg` | `pkg repo` metadata and checksums | Manual FreeBSD packaging repo archive; optional combined Pages |
| SBOM | Repository tree | `anchore/sbom-action@v0` | `unifi-stubd.spdx.json` | Action success | CI artifact and persistent GitHub Release asset |
| Static package site | Linux packages, optionally plus published FreeBSD output | `make package-repos` | `dist/package-site/` | `checksums.txt`; package metadata tools | Linux-first or combined GitHub Pages |

## Deployment and Release Boundaries

- Pull requests run only generic CI validation. They do not publish packages and
  do not run on the self-hosted BSD builder.
- Pushes to `dev` run generic validation only.
- Pushes to `main` run validation and the Linux package smoke job. They upload
  seven-day run artifacts only.
- `v*` tags build Linux packages and SBOM, create or update the GitHub
  pre-release, attach the Linux release set plus `SHA256SUMS`, and deploy the
  Linux-first Pages repositories.
- Manual dispatch from `main` rebuilds the Pages package repositories without
  creating a new GitHub release.
- Manual dispatch of `Controller Integration` tests only the selected ref
  and does not invoke or modify `CI`, package publication, branch rules, or
  required checks.
- Manual dispatch of `FreeBSD Packaging` with `publish=false` builds
  seven-day FreeBSD run artifacts only. With `publish=true` on an existing
  matching `v*` pre-release, it adds the FreeBSD assets and
  `SHA256SUMS.freebsd`, then replaces the Linux-only Pages site with the
  combined Linux/FreeBSD site.

## Problems Found and Corrections

- CI had no explicit top-level token permissions. It now uses
  `permissions: contents: read`.
- The Pages workflow granted Pages/OIDC write permissions to every job. Those
  permissions now exist only on the deploy job.
- BSD tarballs were previously part of automatic CI and release workflows.
  They now come only from the manually dispatched FreeBSD packaging workflow;
  Linux CI, release assets, and Pages no longer wait for the BSD builder.
- The FreeBSD packaging script previously built binaries locally and used the
  BSD host only for `pkg create`/`pkg repo`. It now sends the tracked source
  tree to the builder and builds every configured FreeBSD ABI there. The
  canonical Poudriere host uses host-side Go cross-compilation because its ARM
  Poudriere targets cannot build the Go toolchain under qemu-user-static.
- Release and development documentation described the old cross-build release
  path. The documentation now describes the builder path and the optional
  `jexec` jail mode.

## Open Items

- Decide whether to publish the private Poudriere host repositories in
  addition to the GitHub Pages package site.
- Decide whether to SHA-pin third-party actions. The current workflows use
  stable major-version tags; GitHub's secure-use guidance notes full-length SHA
  pinning as the immutable option.
- Decide whether to add Dependabot updates for GitHub Actions.
- Decide whether Linux arm64 deserves an installation smoke test.
- Decide whether the FreeBSD packaging publication should require a separate
  Tier 3 target-host install/upgrade smoke test before `publish=true` is
  allowed.
- Decide whether package/checksum signing is required before non-alpha
  publication.
- Decide whether a CI build container is necessary. None is currently required
  by the project structure.

## How to Trigger and Verify

Local gates:

```sh
make check
make vulncheck
make package
make package-freebsd-pkg-repos
make package-repos
```

GitHub gates:

- Pull request: `CI / check`.
- Push to `dev`: `CI / check`.
- Push to `main`: `CI / check` and `CI / package`.
- Linux release and package site: push a `v*` tag or run
  `package-pages.yml` manually from `main`.
- Tier 2 controller gate: run `tier2-docker.yml` manually for the exact ref.
- FreeBSD build only: `gh workflow run freebsd-packaging.yml --ref main -f publish=false`.
- FreeBSD publication: after the Linux pre-release exists, run
  `gh workflow run freebsd-packaging.yml --ref <new-v-tag-with-freebsd-workflow> -f publish=true`.

References checked during the audit:

- GitHub Actions workflow syntax and `permissions`:
  `https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax`
- GitHub Pages custom workflow requirements:
  `https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages`
- GitHub secure-use guidance for self-hosted runners and action pinning:
  `https://docs.github.com/en/actions/reference/security/secure-use`
