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

The CI path is split into:

- `CI`: validation, vulnerability scan, SBOM, trusted FreeBSD artifact check,
  and a `main`-only Linux package smoke test.
- `Package Repositories`: release-package build and GitHub Pages deployment for
  unsigned alpha package repositories.

## Build Targets

| Target | Toolchain | Entry point | Output | OS/arch | Build environment | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `make build` | Go 1.25 floor | `go build ./...` | Local binaries in Go build cache | Host OS/arch | Local or GitHub Ubuntu | Defined, not uploaded |
| `make check` | Go, golangci-lint, shell policy | `lint`, `validate-config`, `test` | Pass/fail gate | Host OS/arch | GitHub Ubuntu and local | Primary pre-merge gate |
| `make vulncheck` | Go `govulncheck` tool | `govulncheck ./...` | Pass/fail gate | Host OS/arch | GitHub Ubuntu | CI security gate |
| `make build-freebsd` | Go cross-compile | `cmd/unifi-stubd` | `dist/unifi-stubd_freebsd_<arch>` | FreeBSD `PKG_FREEBSD_GOARCH` | Local only | Kept as local diagnostic target; not release CI |
| `make package` | Go plus nFPM | `scripts/package.sh` | `.deb`, `.rpm`, Arch package, Linux `.tar.gz` | Linux `amd64` or `arm64` | GitHub Ubuntu | Release package path |
| `make package-deb` | Go plus nFPM | `scripts/package.sh deb` | Debian package | Linux current `PKG_GOARCH` | Local or GitHub Ubuntu | Defined |
| `make package-rpm` | Go plus nFPM | `scripts/package.sh rpm` | RPM package | Linux current `PKG_GOARCH` | Local or GitHub Ubuntu | Defined |
| `make package-arch` | Go plus nFPM | `scripts/package.sh archlinux` | Arch Linux package | Linux current `PKG_GOARCH` | Local or GitHub Ubuntu | Defined |
| `make package-tgz` | Go plus tar | `scripts/package.sh tgz` | OS-specific tarball | `PKG_GOOS`/`PKG_GOARCH` | Local | Defined |
| `make package-freebsd-tgz` | Go cross-compile plus tar | `scripts/package.sh tgz` | FreeBSD tarball | FreeBSD `amd64` or `arm64` | Local diagnostic path | Defined, not used by release workflow |
| `make package-freebsd-pkg-repos` | Go on FreeBSD builder, FreeBSD `pkg`, tar | `scripts/package-freebsd-pkg-repos.sh` | Native FreeBSD `pkg` repos and published FreeBSD tarballs | `FreeBSD:14/15` for `amd64`, `aarch64`, `armv7`; tarballs for `amd64`, `arm64` by default | Self-hosted FreeBSD builder; optional mapped `jexec` jails | Release BSD path |
| `make package-repos` | dpkg, createrepo-c, repo-add, tar | `scripts/build-package-repos.sh` | Static package repository site | Repository metadata for packaged targets | GitHub Ubuntu | GitHub Pages input |
| `make integration-docker` | Docker Compose and Go helpers | `lab/stub/scripts/run-docker-integration.sh` | Pass/fail smoke test | Linux container lab | Local/manual | Defined, not automatic in CI |
| Root `Dockerfile` | Docker BuildKit, Go, Alpine | `Dockerfile` | Runtime/lab image with daemon and inform proxy | Linux multi-arch capable | Local/manual | No publish target defined |

Open build target definitions:

- No macOS or Windows build target is defined.
- No top-level `tools/` directory exists; helper tools live under `lab/stub/tools/`
  and selected lab profile directories.
- No CI container image or GitHub Container Registry publishing path is defined.

## GitHub Actions

### `CI`

| Area | Current value |
| --- | --- |
| File | `.github/workflows/ci.yml` |
| Triggers | `workflow_dispatch`, pushes to `dev` and `main`, `pull_request` |
| Default permissions | `contents: read` |
| Jobs | `check`, `freebsd-build`, `package` |
| Runners | `ubuntu-latest`; self-hosted `[self-hosted, Linux, X64, freebsd-pkg-builder]` for trusted FreeBSD builds |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `anchore/sbom-action@v0`, `actions/upload-artifact@v7` |
| Cache | No explicit cache. `setup-go` defaults apply where provided by the action. |
| Artifacts | `sbom`, `packages` on `main` push |
| Security boundary | Self-hosted FreeBSD job is skipped for pull requests. |

`check` runs `make check`, `make vulncheck`, generates `dist/sbom.spdx.json`,
uploads it for seven days, and uses Go stable on `main` while non-main refs use
the module version from `go.mod`.

`freebsd-build` runs only outside pull requests. It calls
`make package-freebsd-pkg-repos` on the self-hosted BSD package builder. The
canonical builder uses FreeBSD host-side Go cross-compilation plus FreeBSD
`pkg`; `FREEBSD_PKG_BUILD_JAILS` is optional and only for custom builders with
running `jexec` jails.

`package` runs only on pushes to `main` after `check` and `freebsd-build`.
It resolves the latest `v[0-9]*` tag, builds Linux packages, installs the
generated Debian package once, validates the installed config, and uploads
`dist/packages/`.

### `Package Repositories`

| Area | Current value |
| --- | --- |
| File | `.github/workflows/package-pages.yml` |
| Triggers | `workflow_dispatch`, `v*` tag push, published/prereleased GitHub Release |
| Default permissions | `contents: read` |
| Deploy permissions | `contents: read`, `pages: write`, `id-token: write` on `build-and-deploy` only |
| Jobs | `build-packages`, `build-freebsd-artifacts`, `build-and-deploy` |
| Runners | Ubuntu for Linux packages and Pages site; self-hosted builder for BSD artifacts |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `actions/upload-artifact@v7`, `actions/download-artifact@v7`, `actions/configure-pages@v6`, `actions/upload-pages-artifact@v5`, `actions/deploy-pages@v5` |
| Artifacts | `linux-package-artifacts`, `freebsd-package-artifacts`, `freebsd-pkg-repos`, Pages artifact |
| Environment | `github-pages` |
| Concurrency | One `github-pages` deployment group, no cancel-in-progress |

Linux packages are built on Ubuntu for `amd64` and `arm64`. FreeBSD tarballs
and native FreeBSD `pkg` repositories are built through the self-hosted builder
path. The deploy job only combines already built artifacts into
`dist/package-site` and deploys that static site.

## OS and Architecture Matrix

| Target | Build path | Test path | Packaging | Publishing |
| --- | --- | --- | --- | --- |
| Linux amd64 | `PKG_GOARCH=amd64 make package` | `make check`; Debian install smoke on `main` | Debian, RPM, Arch, tarball | GitHub Actions artifact and Pages package site |
| Linux arm64 | `PKG_GOARCH=arm64 make package` | `make check` only; no install smoke | Debian, RPM, Arch, tarball | Pages package site |
| FreeBSD 14 amd64 | `make package-freebsd-pkg-repos` via FreeBSD builder | Build/package verification only; target-host smoke manual | Native `pkg` repo | Pages package site |
| FreeBSD 14 aarch64 | Same | Build/package verification only | Native `pkg` repo | Pages package site |
| FreeBSD 14 armv7 | Same | Build/package verification only | Native `pkg` repo | Pages package site |
| FreeBSD 15 amd64 | Same | Build/package verification only | Native `pkg` repo | Pages package site |
| FreeBSD 15 aarch64 | Same | Build/package verification only | Native `pkg` repo | Pages package site |
| FreeBSD 15 armv7 | Same | Build/package verification only | Native `pkg` repo | Pages package site |
| FreeBSD/OPNsense tarball amd64 | Produced from the BSD builder path | Manual target-host smoke | Tarball | Pages package site and manual release assets |
| FreeBSD/OPNsense tarball arm64 | Produced from the BSD builder path | Manual target-host smoke | Tarball | Pages package site and manual release assets |

Open gaps:

- Linux arm64 packages are built but not installation-smoked in CI.
- FreeBSD/OPNsense runtime behavior still requires manual target-host evidence.
- Stable signing is open; alpha repositories are intentionally unsigned.

## Buildhosts, Runners, and Containers

GitHub-hosted Ubuntu is used for generic validation, Linux packaging, repository
metadata, and Pages deployment. The self-hosted `freebsd-pkg-builder` runner is
used only for trusted BSD work and is not used for pull requests. This follows
GitHub's self-hosted-runner risk model: untrusted pull request code must not run
on persistent project infrastructure.

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

- Root `Dockerfile`: builds the daemon and inform proxy into an Alpine runtime
  image.
- Lab gateway profile Dockerfiles: firmware and controller-lab research paths.
- No CI build container, no GHCR image, and no container package release are
  defined.

## Packaging and Publishing

| Format | Source | Build command | Name/output | Validation | Publish target |
| --- | --- | --- | --- | --- | --- |
| Debian | `packaging/nfpm.yaml`, Linux stage tree | `make package` or `make package-deb` | `unifi-stubd_<version>-<release>_<arch>.deb` | Installed and config-validated on `main` for amd64 | Artifact, APT repo |
| RPM | `packaging/nfpm.yaml`, Linux stage tree | `make package` or `make package-rpm` | `unifi-stubd-<version>-<release>.<arch>.rpm` | Repository metadata build | Artifact, RPM repo |
| Arch Linux | `packaging/nfpm.yaml`, Linux stage tree | `make package` or `make package-arch` | `unifi-stubd-<version>-<release>-<arch>.pkg.tar.zst` | Repository metadata build | Artifact, Arch repo |
| Linux tarball | Linux stage tree | `make package` or `make package-tgz` | `unifi-stubd_<version>-<release>_linux_<arch>.tar.gz` | Checksums | Artifact, Pages checksum |
| FreeBSD tarball | FreeBSD stage tree from BSD-built binary | `make package-freebsd-pkg-repos` | `unifi-stubd_<version>-<release>_freebsd_<arch>.tar.gz` | Checksums | Artifact, Pages freebsd paths |
| FreeBSD `pkg` | FreeBSD stage tree and UCL manifest | `make package-freebsd-pkg-repos` | `repo/FreeBSD:*/*` including `packagesite.pkg` | `pkg repo` metadata and checksums | Pages freebsd/pkg paths |
| SBOM | Repository tree | `anchore/sbom-action@v0` | `dist/sbom.spdx.json` | Action success | CI artifact |
| Static package site | Package artifacts and FreeBSD pkg repo output | `make package-repos` | `dist/package-site/` | `checksums.txt`; package metadata tools | GitHub Pages |

## Deployment and Release Boundaries

- Pull requests run only generic CI validation. They do not publish packages and
  do not run on the self-hosted BSD builder.
- Pushes to `dev` run validation plus trusted BSD artifact checks, but do not
  publish package repositories.
- Pushes to `main` run validation, trusted BSD artifact checks, and the Linux
  package smoke job. They upload run artifacts only.
- `v*` tags, GitHub published/prereleased releases, and manual dispatch from
  `main` build and deploy the GitHub Pages package repositories.
- The checked-in workflows do not upload GitHub Release assets. Release asset
  upload remains a manual `gh release create ... dist/releases/...` process in
  the release checklist.

## Problems Found and Corrections

- CI had no explicit top-level token permissions. It now uses
  `permissions: contents: read`.
- The Pages workflow granted Pages/OIDC write permissions to every job. Those
  permissions now exist only on the deploy job.
- BSD tarballs were previously built on Ubuntu in the release workflow. BSD
  release artifacts now come from `make package-freebsd-pkg-repos`.
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
- Decide whether GitHub Release asset upload should become automated. It is
  currently documented as manual.
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
- Trusted push to `dev`: `CI / check` and `CI / freebsd-build`.
- Push to `main`: `CI / check`, `CI / freebsd-build`, and `CI / package`.
- Release package site: push a `v*` tag, publish/pre-release a GitHub Release,
  or run `package-pages.yml` manually from `main`.

References checked during the audit:

- GitHub Actions workflow syntax and `permissions`:
  `https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax`
- GitHub Pages custom workflow requirements:
  `https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages`
- GitHub secure-use guidance for self-hosted runners and action pinning:
  `https://docs.github.com/en/actions/reference/security/secure-use`
