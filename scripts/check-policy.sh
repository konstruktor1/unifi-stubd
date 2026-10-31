#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

consumer_files() {
  git ls-files --cached --others --exclude-standard | while IFS= read -r consumer_file; do
    [ -f "$consumer_file" ] && printf '%s\n' "$consumer_file"
  done
}

printf '== policy: test files live below tests/ ==\n'
invalid_tests="$(consumer_files | awk '/_test[.]go$/ && $0 !~ /^tests\//')"
if [ -n "$invalid_tests" ]; then
  printf '%s\n' "$invalid_tests" >&2
  fail "Go test files must live below tests/"
fi

printf '== policy: runtime identity and passive observation use the platform facade ==\n'
if git grep -nE 'InterfaceByName\(|net\.Interfaces\(|DialContext\(|exec\.Command(Context)?\(' -- \
  ':(glob)cmd/unifi-stubd/*.go' ':(glob)internal/device/*.go' ':(glob)internal/observe/*.go'; then
  fail "runtime identity and portable bridge observation must use the platform facade for host I/O"
fi

printf '== policy: machine-local user paths are not committed ==\n'
machine_path_pattern='/(Users|home)/[^/[:space:]]+'
if git grep --untracked -nE "$machine_path_pattern" -- .; then
  fail "machine-local user path leaked"
fi

printf '== policy: remote FreeBSD packaging stays manually dispatched ==\n'
if grep -q 'freebsd-pkg-builder' \
  .github/workflows/ci.yml \
  .github/workflows/package-pages.yml \
  .github/workflows/tier2-docker.yml; then
  fail "CI, package, and Docker-lab workflows must not select the FreeBSD package builder"
fi
freebsd_workflow=.github/workflows/freebsd-packaging.yml
grep -q 'workflow_dispatch:' "$freebsd_workflow" || fail "FreeBSD Packaging workflow must be manually dispatched"
if grep -Eq '^[[:space:]]+(push|pull_request|pull_request_target|schedule|workflow_call|workflow_run|repository_dispatch|merge_group):' "$freebsd_workflow"; then
  fail "FreeBSD Packaging workflow must not have an automatic trigger"
fi
grep -Fq 'runs-on: [self-hosted, Linux, X64, docker-dmz, unifi-stubd, freebsd-pkg-builder]' "$freebsd_workflow" ||
  fail "FreeBSD Packaging must select the containerized docker.dmz.lan runner with FreeBSD build capability"
if grep -Eiq 'tier[ -]?3|tier3' "$freebsd_workflow"; then
  fail "FreeBSD Packaging must not couple the GitHub Actions runner to a test tier"
fi
controller_workflow=.github/workflows/tier2-docker.yml
for hosted_workflow in .github/workflows/ci.yml .github/workflows/package-pages.yml "$controller_workflow"; do
  if grep -Eq 'runs-on:.*self-hosted' "$hosted_workflow"; then
    fail "Only FreeBSD Packaging may select the self-hosted runner"
  fi
  grep -Eq 'runs-on: ubuntu-(latest|[0-9]+\.[0-9]+)' "$hosted_workflow" ||
    fail "General CI and container orchestration require a GitHub-hosted Ubuntu runner"
done
grep -Fq 'run: make integration-nonunifi-metrics' "$controller_workflow" ||
  fail "Controller Integration must enter through the runner-aware Make target"
if grep -Eq 'actions/setup-go|(^|[[:space:]])go version|docker (compose|info|version)' "$controller_workflow"; then
  fail "Controller Integration must not create a workflow-local Go or Docker toolchain"
fi

printf '== policy: controller lab lifecycle stays behind Build Runner ==\n'
lab_compose=lab/stub/compose.yaml
lab_runner=lab/stub/scripts/run-docker-integration.sh
if grep -Eq '^[[:space:]]*(container_name:|restart:[[:space:]]+unless-stopped)' "$lab_compose"; then
  fail "Controller lab containers must not use host-global names or persistent restart policies"
fi
if grep -Eq 'docker volume (ls|rm)|UNIFI_STUB_DOCKER_KEEP_RESOURCES' "$lab_runner"; then
  fail "Controller lab lifecycle must not expose partial volume cleanup or a keep-running bypass"
fi
grep -Fq 'compose down --volumes --remove-orphans' "$lab_runner" ||
  fail "Controller lab runner must remove the complete Compose project"

printf '== policy: release version is centralized and documented ==\n'
release_version=$(sed -n '1p' VERSION)
case "$release_version" in
[0-9]*.[0-9]*.[0-9]*-alpha) ;;
*)
  fail "VERSION must contain an alpha semantic version without a leading v"
  ;;
esac

printf '== policy: Go module and workspace use a flexible minor floor ==\n'
mod_go=$(sed -n 's/^go //p' go.mod | head -n 1)
work_go=$(sed -n 's/^go //p' go.work | head -n 1)
[ "$mod_go" = "$work_go" ] || fail "go.mod and go.work must use the same Go version"

major=${mod_go%%.*}
minor=${mod_go#*.}
minor=${minor%.0}
if [ "$major" = "$mod_go" ] || [ -z "$major" ] || [ -z "$minor" ]; then
  fail "go.mod/go.work must use a major.minor floor with an optional .0 patch"
fi
case "$major" in
'' | *[!0-9]*)
  fail "go.mod/go.work major version must be numeric"
  ;;
esac
case "$minor" in
'' | *.* | *[!0-9]*)
  fail "go.mod/go.work must use a major.minor floor with an optional .0 patch"
  ;;
esac
if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 27 ]; }; then
  fail "go.mod/go.work must require Go 1.27 or newer"
fi
if grep -Eq '^toolchain ' go.mod go.work; then
  fail "toolchain must not be pinned in committed module or workspace files"
fi
grep -qx 'use \.' go.work || fail "go.work must include the root module"
grep -qx 'tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint' golangci-lint.mod || fail "golangci-lint must be tracked as a go tool in golangci-lint.mod"
grep -qx '	github.com/goreleaser/nfpm/v2/cmd/nfpm' go.mod || fail "nfpm must be tracked as a go tool"

printf 'policy ok\n'
