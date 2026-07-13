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
   make package`.
4. Inspect `dist/packages/` for the Linux Debian, RPM, Arch Linux, and `.tar.gz`
   output.
5. Confirm `go.mod`, `go.work`, and `go.sum` are tidy after dependency or tool updates.
6. Search for private lab data:

   ```sh
   sh scripts/check-policy.sh
   ```

7. Confirm `VERSION`, the newest dated `CHANGELOG.md` section, and the intended
   `v<version>` tag all match.
8. Update `CREDITS.md` and `NOTICE.md` if new sources, copied code, packages,
   or tools were added.
9. Confirm `packaging/linux/etc/unifi-stubd/config.yaml` and `lab/` contain
   documentation addresses or neutral defaults only.
10. Verify repository metadata with `gh repo view --json
    description,homepageUrl,licenseInfo,repositoryTopics` and compare package
    name, license, maintainer, and project URL with `packaging/nfpm.yaml`.
11. Verify signing status. Alpha releases may publish the generated unsigned
    `SHA256SUMS` only while they remain marked as pre-releases. Stable releases
    require a signed tag and a detached signature for `SHA256SUMS`.

FreeBSD package generation is not a pre-tag gate. It belongs exclusively to
the manual `FreeBSD Packaging` workflow and may extend the pre-release only
after the automatic Linux publication has completed. This packaging pipeline
is not a Tier 1, Tier 2, or Tier 3 test classification.

## Tagging

Use semantic version tags:

```sh
git tag -a v0.2.5-alpha -m "unifi-stubd v0.2.5-alpha"
git push origin v0.2.5-alpha
```

Tag only commits that already passed the `main` CI run. The Package
Repositories workflow runs for `v*` tags, so the `github-pages` environment
must allow deployments from `main` and `v*` tags.

The tag push owns the first alpha publication. The Package Repositories
workflow creates the GitHub pre-release, attaches Linux packages,
`SHA256SUMS`, and `unifi-stubd.spdx.json`, and deploys the Linux-only Pages
site. It never starts or waits for FreeBSD packaging. Verify the persistent
Linux release assets after the workflow completes:

```sh
gh release view v0.2.5-alpha --json isPrerelease,assets,url
```

## Manual FreeBSD Packaging Extension

`FreeBSD Packaging` is manual-only and has no trigger or `needs`
relationship with `CI` or `Package Repositories`. Its `publish` input defaults
to `false`; that mode builds FreeBSD tarballs and native `pkg` repositories for
the selected ref and retains the Actions artifacts for seven days as build
evidence:

```sh
gh workflow run freebsd-packaging.yml --ref main -f publish=false
```

To publish the result, dispatch the exact release tag with explicit
confirmation:

```sh
gh workflow run freebsd-packaging.yml --ref <v-tag-with-freebsd-workflow> -f publish=true
```

Publication is accepted only when the selected ref is a `v*` tag, the tag
without `v` matches `VERSION`, and a matching GitHub pre-release already
contains the complete Linux package set, SPDX SBOM, and `SHA256SUMS`. The
workflow also refuses to mutate a release that already contains a signed
`SHA256SUMS` or `SHA256SUMS.freebsd`.
It also rejects legacy releases whose `SHA256SUMS` already covers FreeBSD
artifacts. The selected tag must contain this workflow, so the split begins
with the first new release tag created after this implementation lands.

A successful publishing run adds these persistent assets and then deploys a
combined Linux and FreeBSD Pages site:

- `unifi-stubd_<version>-1_freebsd_amd64.tar.gz`
- `unifi-stubd_<version>-1_freebsd_arm64.tar.gz`
- `freebsd-pkg-repos.tar.gz`
- `SHA256SUMS.freebsd`

`SHA256SUMS.freebsd` covers only the three FreeBSD artifacts. The automatic
Linux and SBOM checksum file remains unchanged. Rerunning `FreeBSD Packaging`
for the same pre-release is idempotent because the four FreeBSD assets are
replaced by name.

## Signing And Repository Metadata

When a release signing key is configured, create a signed tag with `git tag
-s`, download the release `SHA256SUMS`, and publish a detached signature such
as `SHA256SUMS.asc`. If FreeBSD assets are present, sign
`SHA256SUMS.freebsd` separately as well. Verify the tag and every published
checksum signature before promotion. Without a configured key, keep the
release alpha, pre-release, and explicitly unsigned; do not promote it to
stable. Add FreeBSD artifacts before signing because the FreeBSD packaging
workflow will not modify a release with signed checksum assets.

Repository description, homepage, SPDX license, and topics must remain
populated and consistent with `README.md`, `LICENSE`, and package metadata.
The release target must be the tested commit on `main`, and the tag SHA must
match the release target SHA.

## GitHub Pages Package Repositories

Alpha package repositories are published unsigned through GitHub Pages:

```text
https://konstruktor1.github.io/unifi-stubd/
```

Build all package artifacts first, then generate the static repository site:

```sh
PKG_VERSION=0.2.5-alpha PKG_RELEASE=1 PKG_GOARCH=amd64 make package
PKG_VERSION=0.2.5-alpha PKG_RELEASE=1 PKG_GOARCH=arm64 make package
PKG_VERSION=0.2.5-alpha PKG_RELEASE=1 make package-freebsd-pkg-repos
make package-repos
```

The Package Repositories workflow publishes a Linux-only package repository
through GitHub Pages when a `v*` tag is pushed. Until FreeBSD packaging is
published, the FreeBSD section states that the manual artifacts are
unavailable. To retry or
rebuild this Linux-only repository without creating another release, run the
workflow from `main`:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.5-alpha \
  -f package_release=1
```

If `version` is omitted in a manual run, the workflow reads `VERSION`.

A later Linux-only Package Repositories run replaces the combined Pages site
with the Linux-only form. Rerun `freebsd-packaging.yml` with `publish=true` on
the matching tag to restore the combined site from the persistent release
assets.

The SPDX JSON SBOM is attached to the GitHub release and therefore does not
expire with the seven-day Actions artifact retention. `SHA256SUMS` covers the
automatic Linux and SBOM assets; `SHA256SUMS.freebsd` covers the separately
published FreeBSD assets.

`make package-repos` writes `dist/package-site/` with APT, RPM, Arch Linux,
FreeBSD/OPNsense tarball paths, and native FreeBSD `pkg` repository paths when
`dist/freebsd-pkg-repos/repo/` exists. Its default mode requires the FreeBSD
artifacts; the automatic workflow explicitly allows their absence for the
Linux-first deployment. Only the manual `FreeBSD Packaging` workflow uses the
containerized self-hosted GitHub Actions runner on `docker.dmz.lan`, selected
by `[self-hosted, Linux, X64, docker-dmz, unifi-stubd,
freebsd-pkg-builder]`. It uses strict SSH/SCP to a non-root FreeBSD build-host
account. The canonical builder uses FreeBSD host-side Go cross-compilation plus
FreeBSD `pkg`;
`FREEBSD_PKG_BUILD_JAILS` is optional for custom builders with running `jexec`
jails. The combined static site is assembled and deployed from Ubuntu after
the FreeBSD build. Keep alpha repository instructions visibly unsigned
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
