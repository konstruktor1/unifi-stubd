# Release Checklist

Use this checklist before publishing a tag or package set.

## Before Tagging

1. Run `make check`.
2. Run `make package`.
3. Inspect `dist/packages/` for Debian, RPM, Arch Linux, and `.tar.gz` output.
4. Confirm `go.mod`, `go.work`, and `go.sum` are tidy after dependency or tool updates.
5. Search for private lab data:

   ```sh
   sh scripts/check-policy.sh
   ```

6. Update `CHANGELOG.md`.
7. Update `CREDITS.md` and `NOTICE.md` if new sources, copied code, packages,
   or tools were added.
8. Confirm `packaging/linux/etc/unifi-stubd/config.yaml` and `lab/` contain
   documentation addresses or neutral defaults only.

## Tagging

Use semantic version tags:

```sh
git tag -a v0.1.0 -m "unifi-stubd v0.1.0"
git push origin v0.1.0
```

The GitHub Actions CI workflow builds packages for `amd64` and `arm64`.

## Package Metadata

The default package maintainer is `unifi-stubd maintainers <info@spinas.org>`.
Override maintainer metadata for public package builds:

```sh
PKG_MAINTAINER='Name <email@example.com>' PKG_GOARCH=amd64 make package
```
