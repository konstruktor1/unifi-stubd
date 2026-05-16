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
if rg -n 'JoemcROV|X-API-KEY|10\.0\.0\.194|/Users/corspi' . \
  --glob '!.git/**' \
  --glob '!.utm-build/**' \
  --glob '!dist/**' \
  --glob '!scripts/check-policy.sh'; then
  fail "possible lab secret or API endpoint leaked"
fi

printf '== policy: Go module and workspace use a flexible minor floor ==\n'
mod_go=$(sed -n 's/^go //p' go.mod | head -n 1)
work_go=$(sed -n 's/^go //p' go.work | head -n 1)
[ "$mod_go" = "$work_go" ] || fail "go.mod and go.work must use the same Go version"

major=${mod_go%%.*}
minor=${mod_go#*.}
if [ "$major" = "$mod_go" ] || [ -z "$major" ] || [ -z "$minor" ]; then
  fail "go.mod/go.work must use an unpatched Go minor version, for example: go 1.25"
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
if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 25 ]; }; then
  fail "go.mod/go.work must use Go 1.25 or newer"
fi
if grep -Eq '^toolchain ' go.mod go.work; then
  fail "toolchain must not be pinned in committed module or workspace files"
fi
grep -qx 'use \.' go.work || fail "go.work must include the root module"
grep -qx '	github.com/golangci/golangci-lint/v2/cmd/golangci-lint' go.mod || fail "golangci-lint must be tracked as a go tool"
grep -qx '	github.com/goreleaser/nfpm/v2/cmd/nfpm' go.mod || fail "nfpm must be tracked as a go tool"

printf 'policy ok\n'
