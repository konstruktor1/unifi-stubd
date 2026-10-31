# Release Checklist

Use this checklist before publishing a tag or package set.

Release artifacts are neutral by design. They must not include real controller
URLs, site IP addresses, MAC addresses, client names, credentials, adoption
state, or host-specific interface mappings. Keep those files in a private
configuration store outside the repository and copy the selected host config
onto the installed system after the package is installed.

## Before Tagging

1. Run the complete host-side gate `make verify` successfully, then the required
   package targets through their configured routes. `make check` is the Go
   product gate inside the prepared runner, not the complete release gate;
   [`AGENTS.md`](../../AGENTS.md) documents the validation entry points. The
   complete gate includes the private lab-data policy check.
2. Build the required `amd64` and `arm64` release artifacts.
3. Inspect `dist/packages/` for the Linux Debian, RPM, Arch Linux, and `.tar.gz`
   output.
4. After dependency or tool updates, review `go.mod`, `go.work`, and `go.sum`,
   the isolated linter module `golangci-lint.mod` and `golangci-lint.sum`, and
   the root npm manifest and lockfile, `package.json` and `package-lock.json`.
5. Confirm `VERSION`, the newest dated `CHANGELOG.md` section, and the intended
   `v<version>` tag all match.
6. Update `CREDITS.md` and `NOTICE.md` if new sources, copied code, packages,
   or tools were added.
7. Confirm `packaging/linux/etc/unifi-stubd/config.yaml` and `lab/` contain
   documentation addresses or neutral defaults only.
8. Verify repository metadata with `gh repo view --json
   description,homepageUrl,licenseInfo,repositoryTopics` and compare package
   name, license, maintainer, and project URL with `packaging/nfpm.yaml`.
9. Verify signing status. Alpha releases may publish the generated unsigned
   `SHA256SUMS` only while they remain marked as pre-releases. Stable releases
   require a signed tag and a detached signature for `SHA256SUMS`.

FreeBSD package generation is not a pre-tag gate. The manual `FreeBSD
Packaging` workflow produces compatibility artifacts only; publication remains
disabled until an authorized build route exists. This packaging pipeline is
not a Tier 1, Tier 2, or Tier 3 test classification.

## Tagging

Use semantic version tags:

```sh
git tag -a v0.2.5-alpha -m "unifi-stubd v0.2.5-alpha"
git push origin v0.2.5-alpha
```

Tag only commits whose exact tree passed `make verify`. General CI prepares
current BR/CT sources and native bindings on GitHub-hosted Ubuntu workers, then
uses Build Runner containers. Private Forgejo read access and delivery of current
provider changes must be available and a successful CI run observed before a new
tag is published; see the [open provider handoffs](ci-cd.md#open-provider-handoffs).
The self-hosted runner remains exclusive to manual FreeBSD packaging.
The Package Repositories workflow runs for
`v*` tags, so the `github-pages` environment must allow deployments from `main`
and `v*` tags.

The tag push owns the first alpha publication. The Package Repositories
workflow creates the GitHub pre-release, attaches Linux packages,
`SHA256SUMS`, and `unifi-stubd.spdx.json`, and deploys the Linux-only Pages
site. It never starts or waits for FreeBSD packaging. Verify the persistent
Linux release assets after the workflow completes:

```sh
gh release view v0.2.5-alpha --json isPrerelease,assets,url
```

## Manual FreeBSD Compatibility Packaging

`FreeBSD Packaging` is manual-only and has no trigger or `needs`
relationship with `CI` or `Package Repositories`. It retains its FreeBSD
tarballs, native `pkg` repositories, and `SHA256SUMS.freebsd` as run artifacts
for seven days. Its direct host-Go builder is an external-machine compatibility
path and those artifacts are not compliant build evidence:

```sh
gh workflow run freebsd-packaging.yml --ref main
```

The selected ref must be `main` or a `v*` tag matching `VERSION`. The workflow
has no publication input, release upload, or Pages deployment. Do not publish
these artifacts manually. The FreeBSD build owner must provide an authorized
VM Runner or reviewed split-build route for compilation before the release
owner restores publication after Linux publication. That route must preserve
the matching-tag, complete Linux pre-release, and unsigned-checksum safeguards.

## Signing And Repository Metadata

When a release signing key is configured, create a signed tag with `git tag
-s`, download the release `SHA256SUMS`, and publish a detached signature such
as `SHA256SUMS.asc`. If FreeBSD assets are present, sign
`SHA256SUMS.freebsd` separately as well. Verify the tag and every published
checksum signature before promotion. Without a configured key, keep the
release alpha, pre-release, and explicitly unsigned; do not promote it to
stable. Signed checksum assets must never be replaced or extended by packaging
automation.

Repository description, homepage, SPDX license, and topics must remain
populated and consistent with `README.md`, `LICENSE`, and package metadata.
The release target must be the tested commit on `main`, and the tag SHA must
match the release target SHA.

## GitHub Pages Package Repositories

Alpha package repositories are published unsigned through GitHub Pages:

```text
https://konstruktor1.github.io/unifi-stubd/
```

The Package Repositories workflow publishes a Linux-only package repository
through GitHub Pages when a `v*` tag is pushed. While FreeBSD publication remains
blocked, the FreeBSD section states that the manual artifacts are unavailable.
To retry or rebuild this Linux-only repository without creating another
release, run the workflow from `main`:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.5-alpha \
  -f package_release=1
```

If `version` is omitted in a manual run, the workflow reads `VERSION`.

A Package Repositories run replaces any existing combined Pages site with the
Linux-only form. Restoring a combined site requires the authorized FreeBSD
build and publication route described above.

The SPDX JSON SBOM is attached to the GitHub release and therefore does not
expire with the seven-day Actions artifact retention. `SHA256SUMS` covers the
automatic Linux and SBOM assets. `SHA256SUMS.freebsd` accompanies the temporary
FreeBSD run artifacts.

`make package-repos` writes `dist/package-site/` with APT, RPM, Arch Linux,
FreeBSD/OPNsense tarball paths, and native FreeBSD `pkg` repository paths when
`dist/freebsd-pkg-repos/repo/` exists. Its default mode requires the FreeBSD
artifacts; the automatic workflow explicitly allows their absence for the
Linux-first deployment. Only the manual `FreeBSD Packaging` workflow uses the
isolated packaging-proxy role and strict SSH/SCP to a non-root FreeBSD-builder
role. Its current direct host-Go compilation is compatibility-only and blocked
as release evidence; FreeBSD `pkg` still owns target repository assembly.
`FREEBSD_PKG_BUILD_JAILS` is optional for custom builders with running `jexec`
jails. The FreeBSD workflow does not assemble or deploy a combined site. Keep
alpha repository instructions visibly unsigned
(`trusted=yes`, `gpgcheck=0`, `SigLevel = Never`) until a project release key
exists.

The generated project page also links to the source repository, releases,
`CREDITS.md`, and a short research/source-project map. Keep that map as a
summary only; the authoritative attribution matrix stays in `CREDITS.md`.

## Package Metadata

The default package maintainer is `unifi-stubd maintainers <info@spinas.org>`.
An authorized runner invocation may override it with `PKG_MAINTAINER`.

## Host Configurations

Keep real host configurations outside this repository in private deployment
storage. The package does not change VLANs, firewall rules, routes, or
controller network definitions.
