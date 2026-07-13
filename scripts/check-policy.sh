#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

printf '== policy: test files live below tests/ ==\n'
if find . -name '*_test.go' -not -path './.git/*' -not -path './tests/*' | grep -q .; then
  find . -name '*_test.go' -not -path './.git/*' -not -path './tests/*' >&2
  fail "Go test files must live below tests/"
fi

printf '== policy: lab secrets are not committed ==\n'
if git grep -nE 'JoemcROV|X-API-KEY|10\.0\.0\.194|/Users/corspi' -- \
  ':!scripts/check-policy.sh'; then
  fail "possible lab secret or API endpoint leaked"
fi

printf '== policy: privileged FreeBSD builder stays manual Tier 3 ==\n'
if grep -q 'freebsd-pkg-builder' \
  .github/workflows/ci.yml \
  .github/workflows/package-pages.yml; then
  fail "automatic CI and package workflows must not select the Tier 3 FreeBSD builder"
fi
tier3_workflow=.github/workflows/tier3-freebsd.yml
grep -q 'workflow_dispatch:' "$tier3_workflow" || fail "Tier 3 FreeBSD workflow must be manually dispatched"
if grep -Eq '^[[:space:]]+(push|pull_request|pull_request_target|schedule|workflow_call|workflow_run|repository_dispatch|merge_group):' "$tier3_workflow"; then
  fail "Tier 3 FreeBSD workflow must not have an automatic trigger"
fi
grep -Fq 'runs-on: [self-hosted, Linux, X64, unifi-stubd, tier3, freebsd-pkg-builder]' "$tier3_workflow" ||
  fail "Tier 3 FreeBSD workflow must use only the dedicated FreeBSD builder selector"

printf '== policy: release version is centralized and documented ==\n'
release_version=$(sed -n '1p' VERSION)
case "$release_version" in
  [0-9]*.[0-9]*.[0-9]*-alpha)
    ;;
  *)
    fail "VERSION must contain an alpha semantic version without a leading v"
    ;;
esac
grep -Fq "## [$release_version]" CHANGELOG.md || fail "CHANGELOG.md is missing VERSION $release_version"

printf '== policy: Go module and workspace use a flexible minor floor ==\n'
mod_go=$(sed -n 's/^go //p' go.mod | head -n 1)
work_go=$(sed -n 's/^go //p' go.work | head -n 1)
[ "$mod_go" = "$work_go" ] || fail "go.mod and go.work must use the same Go version"

major=${mod_go%%.*}
minor=${mod_go#*.}
if [ "$major" = "$mod_go" ] || [ -z "$major" ] || [ -z "$minor" ]; then
  fail "go.mod/go.work must use a major.minor version without a patch component"
fi
case "$major" in
  '' | *[!0-9]*)
    fail "go.mod/go.work major version must be numeric"
    ;;
esac
case "$minor" in
  '' | *.* | *[!0-9]*)
    fail "go.mod/go.work must not pin a Go patch version"
    ;;
esac
if grep -Eq '^toolchain ' go.mod go.work; then
  fail "toolchain must not be pinned in committed module or workspace files"
fi
grep -qx 'use \.' go.work || fail "go.work must include the root module"
grep -qx '	github.com/golangci/golangci-lint/v2/cmd/golangci-lint' go.mod || fail "golangci-lint must be tracked as a go tool"
grep -qx '	github.com/goreleaser/nfpm/v2/cmd/nfpm' go.mod || fail "nfpm must be tracked as a go tool"

printf 'policy ok\n'
