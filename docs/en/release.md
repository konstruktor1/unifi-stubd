# Release Checklist

Use this checklist before publishing a tag or package set.

Release artifacts are neutral by design. They must not include real controller
URLs, site IP addresses, MAC addresses, client names, credentials, adoption
state, or host-specific interface mappings. Keep those files in a private
configuration store outside the repository and copy the selected host config
onto the installed system after the package is installed.

## Before Tagging

1. Run `make check`.
2. Run `make package`.
3. Build any cross-architecture release targets, such as `PKG_GOARCH=arm64
   make package` and `PKG_FREEBSD_GOARCH=amd64 make package-freebsd-tgz`.
4. Inspect `dist/packages/` for Debian, RPM, Arch Linux, and `.tar.gz` output.
5. Confirm `go.mod`, `go.work`, and `go.sum` are tidy after dependency or tool updates.
6. Search for private lab data:

   ```sh
   sh scripts/check-policy.sh
   ```

7. Update `CHANGELOG.md`.
8. Update `CREDITS.md` and `NOTICE.md` if new sources, copied code, packages,
   or tools were added.
9. Confirm `packaging/linux/etc/unifi-stubd/config.yaml` and `lab/` contain
   documentation addresses or neutral defaults only.
10. For public releases, sign packages or checksums with the project release key
   once that key exists; do not publish unsigned artifacts as stable releases.

## Tagging

Use semantic version tags:

```sh
git tag -a v0.1.0 -m "unifi-stubd v0.1.0"
git push origin v0.1.0
```

The normal CI workflow validates the tagged commit. The Package Repositories
workflow also runs for `v*` tags and GitHub pre-releases, so the `github-pages`
environment must allow deployments from `main` and `v*` tags.

Create a pre-release for alpha package sets:

```sh
gh release create v0.1.4-alpha --prerelease \
  --title "unifi-stubd v0.1.4-alpha" \
  --notes-file dist/releases/v0.1.4-alpha/release-notes.md \
  dist/releases/v0.1.4-alpha/*
```

## GitHub Pages Package Repositories

Alpha package repositories are published unsigned through GitHub Pages:

```text
https://konstruktor1.github.io/unifi-stubd/
```

Build all package artifacts first, then generate the static repository site:

```sh
PKG_VERSION=0.1.4-alpha PKG_RELEASE=1 PKG_GOARCH=amd64 make package
PKG_VERSION=0.1.4-alpha PKG_RELEASE=1 PKG_GOARCH=arm64 make package
PKG_VERSION=0.1.4-alpha PKG_RELEASE=1 PKG_FREEBSD_GOARCH=amd64 make package-freebsd-tgz
PKG_VERSION=0.1.4-alpha PKG_RELEASE=1 PKG_FREEBSD_GOARCH=arm64 make package-freebsd-tgz
PKG_VERSION=0.1.4-alpha PKG_RELEASE=1 make package-freebsd-pkg-repos
make package-repos
```

The Package Repositories workflow publishes the package repository through
GitHub Pages when a `v*` tag or pre-release is published. To retry or rebuild
the package repository manually, run the workflow from `main`:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.1.4-alpha \
  -f package_release=1
```

`make package-repos` writes `dist/package-site/` with APT, RPM, Arch Linux,
FreeBSD/OPNsense tarball paths, and native FreeBSD `pkg` repository paths when
`dist/freebsd-pkg-repos/repo/` exists. The Package Repositories workflow builds
those native FreeBSD repositories on the self-hosted runner labelled
`freebsd-pkg-builder`, then deploys the combined static site from Ubuntu. Keep
alpha repository instructions visibly unsigned
(`trusted=yes`, `gpgcheck=0`, `SigLevel = Never`) until a project release key
exists.

The generated project page also links to the source repository, releases, wiki,
`CREDITS.md`, and a short research/source-project map. Keep that map as a
summary only; the authoritative attribution matrix stays in `CREDITS.md`.

## Package Metadata

The default package maintainer is `unifi-stubd maintainers <info@spinas.org>`.
Override maintainer metadata for public package builds:

```sh
PKG_MAINTAINER='Name <email@example.com>' PKG_GOARCH=amd64 make package
```

## Host Configurations

Do not place real host configs in this repository. Store them in a private
directory or private deployment repository. A useful local layout is:

```text
unifi-stubd-host-configs/
  hosts/<host>/current-tmp.yaml
  hosts/<host>/installed-config.yaml
  hosts/<host>/process.txt
  hosts/<host>/status-before.json
  hosts/<host>/status-after.json
```

`current-tmp.yaml` records the temporary runtime config. `installed-config.yaml`
is the package-ready config copied to the installed service path. The package
itself remains neutral and does not change VLANs, firewall rules, routes, or
controller network definitions.
