GO ?= go
# Keep installed dependencies such as node_modules outside consumer Go analysis.
GO_PACKAGES := ./cmd/... ./internal/... ./tests/... ./lab/stub/tools/controller-adoption
# Keep the linter dependency graph outside the product module.
GOLANGCI_LINT ?= env GOWORK=off $(GO) tool -modfile=golangci-lint.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOLANGCI_LINT_CACHE ?= $(GOCACHE)/golangci-lint
RUNNER_WORKDIR ?= $(GOCACHE)/unifi-stubd-work
GOVULNCHECK ?= $(GO) tool golang.org/x/vuln/cmd/govulncheck
COVERAGE_PACKAGES ?= ./internal/...
COVERAGE_TESTS ?= ./tests/...
PKG_VERSION ?= $(shell sed -n '1p' VERSION)
PKG_RELEASE ?= 1
PKG_GOOS ?= linux
# Keep Makefile parsing independent of Go. scripts/package.sh resolves the
# current architecture inside Build Runner when no explicit target is set.
PKG_GOARCH ?=
PKG_FREEBSD_GOARCH ?= amd64
PKG_FORMATS ?= deb rpm archlinux tgz
PKG_LICENSE ?= AGPL-3.0-or-later
PKG_MAINTAINER ?= unifi-stubd maintainers <info@spinas.org>
BUILD_LDFLAGS := -s -w -X main.version=$(PKG_VERSION)
BUILD_RUNNER_PROJECT_NAME := unifi-stubd
# Only host orchestration calls providers; payload targets execute in /workspace.
HOST_PROVIDER_GOALS := verify quality-deps quality-js quality-ct quality-ct-image provider-check provider-context doctor \
  integration-nonunifi-metrics
BUILD_RUNNER_CLI_GOALS := verify quality-deps quality-js \
  integration-nonunifi-metrics
ifneq ($(filter $(HOST_PROVIDER_GOALS),$(MAKECMDGOALS)),)
ifndef PROVIDER_CONSUMER_ROOT
$(error Inherit PROVIDER_CONSUMER_ROOT from the existing native project/session owner)
endif
ifndef PROVIDER_CONSUMER_CONTEXT
$(error Inherit PROVIDER_CONSUMER_CONTEXT from the existing native project/session owner)
endif
export PROVIDER_CONSUMER_ROOT PROVIDER_CONSUMER_CONTEXT
ifneq ($(PROVIDER_CONSUMER_ROOT),$(shell pwd -P))
$(error Inherit this project's canonical native PROVIDER_CONSUMER_ROOT unchanged)
endif
PROVIDER_CLIENT_PYTHON := $(PROVIDER_CONSUMER_ROOT)/.consumer-data/dependencies/unifi-stubd/mcp/bin/python
endif
ifneq ($(filter $(BUILD_RUNNER_CLI_GOALS),$(MAKECMDGOALS)),)
BUILD_RUNNER_ENTRY := $(shell jq -er '.command | select(endswith("/bin/provider")) | sub("/provider$$"; "/build-runner")' .consumer-data/connections/build-runner/management.json)
endif

ifneq ($(strip $(GOCACHE)),)
GOTMPDIR ?= $(RUNNER_WORKDIR)
TMPDIR ?= $(RUNNER_WORKDIR)
# Go's checksum database also needs writable project metadata outside GOMODCACHE.
GOPATH := $(RUNNER_WORKDIR)/gopath
export GOPATH GOTMPDIR TMPDIR
endif

PKG_ENV := PKG_VERSION='$(PKG_VERSION)' \
  PKG_RELEASE='$(PKG_RELEASE)' \
  PKG_GOOS='$(PKG_GOOS)' \
  PKG_GOARCH='$(PKG_GOARCH)' \
  PKG_LICENSE='$(PKG_LICENSE)' \
  PKG_MAINTAINER='$(PKG_MAINTAINER)' \
  BUILD_LDFLAGS='$(BUILD_LDFLAGS)'

.PHONY: build build-freebsd check clean-dist coverage doctor fmt fmt-check help integration-nonunifi-metrics lint package package-arch package-deb package-freebsd-pkg-repos package-freebsd-tgz package-repos package-rpm package-tgz policy provider-check provider-context quality-deps quality-js race runner-workdir switch-emulation switch-payload test validate-config verify vet vulncheck workflow-lint

build build-freebsd coverage fmt fmt-check lint package package-arch package-deb package-freebsd-tgz package-rpm package-tgz race switch-emulation switch-payload test validate-config vet vulncheck workflow-lint: | runner-workdir

runner-workdir:
	@test "$$(pwd -P)" = /workspace && test -n '$(GOCACHE)' || { printf '%s\n' 'Go targets must run in the Build Runner workspace' >&2; exit 1; }
	@mkdir -p '$(RUNNER_WORKDIR)'

help:
	@printf '%s\n' \
		'Targets:' \
		'  make verify       Run the complete quality gate through prepared providers' \
		'  make provider-check  Read and verify current native Build Runner context' \
		'  make quality-ct   Run CT in its confirmed check environment' \
		'  make doctor       Inspect local setup and report required coverage gaps' \
		'  make check        Run format checks, vet, lint, policy checks, and tests' \
		'  make build-freebsd  Cross-build the FreeBSD/OPNsense binary' \
		'  make coverage     Write a diagnostic coverage report' \
		'  make lint         Run golangci-lint and repository policy checks' \
		'  make race         Run all Go tests with the race detector' \
		'  make test         Run all Go tests' \
		'  make validate-config  Validate packaged configs and example profiles' \
		'  make vulncheck    Run Go vulnerability scanning' \
		'  make workflow-lint  Validate GitHub workflows and local action metadata' \
		'  make integration-nonunifi-metrics  Prove non-UniFi metrics with the Stable/RC matrix' \
		'  make switch-payload  Print discovery and inform payloads' \
		'  make switch-emulation  Start the lab switch emulator' \
		'  make package      Build deb, rpm, archlinux, and tgz packages' \
		'  make package-repos  Build unsigned alpha package repositories in dist/package-site' \
		'  make package-freebsd-tgz  Build FreeBSD/OPNsense stub-only tgz' \
		'  make package-freebsd-pkg-repos  Build FreeBSD tarballs and pkg repositories through the configured FreeBSD builder' \
		'  make clean-dist   Remove package build output'

provider-context:
	@printf '%s\n' '$(PROVIDER_CONSUMER_CONTEXT)'

provider-check:
	"$(PROVIDER_CLIENT_PYTHON)" -B scripts/check-ct.py --check-build-runner

doctor:
	"$(PROVIDER_CLIENT_PYTHON)" -B scripts/check-ct.py --doctor --consumer-role terminal

quality-deps: provider-check
	"$(BUILD_RUNNER_ENTRY)" --project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
		run --tier 1 --service node24 -- npm ci --ignore-scripts --no-audit --no-fund

quality-js: quality-deps
	"$(BUILD_RUNNER_ENTRY)" --project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
		run --tier 1 --service node24 -- npm run check

.PHONY: quality-ct quality-ct-image

# BR has no consumer Dockerfile build operation; execution stays with its runner.
quality-ct-image:
	@set -eu; case "$$(uname -m)" in \
		arm64|aarch64) platform=linux/arm64 ;; \
		x86_64|amd64) platform=linux/amd64 ;; \
		*) printf '%s\n' 'Unsupported native CT image architecture' >&2; exit 1 ;; \
		esac; \
		docker build --platform "$$platform" -t local/unifi-stubd-ct-quality:current \
			-f build-runner/project-setup/ct-quality/Dockerfile build-runner/project-setup/ct-quality; \
		test "$$(docker image inspect local/unifi-stubd-ct-quality:current --format '{{.Os}}/{{.Architecture}}')" = "$$platform"

quality-ct:
	"$(PROVIDER_CLIENT_PYTHON)" -B scripts/check-ct.py

verify: quality-js
	git diff --check HEAD
	"$(BUILD_RUNNER_ENTRY)" --project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
		run --tier 1 --service go -- make check vulncheck workflow-lint
	$(MAKE) quality-ct

workflow-lint:
	GOWORK=off $(GO) run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12

build:
	$(GO) build -ldflags='$(BUILD_LDFLAGS)' $(GO_PACKAGES)

build-freebsd:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=freebsd GOARCH='$(PKG_FREEBSD_GOARCH)' $(GO) build -trimpath -ldflags='$(BUILD_LDFLAGS)' -o dist/unifi-stubd_freebsd_$(PKG_FREEBSD_GOARCH) ./cmd/unifi-stubd

check: fmt-check vet lint validate-config test
lint:
	GOLANGCI_LINT_CACHE='$(GOLANGCI_LINT_CACHE)' $(GOLANGCI_LINT) config verify
	GOLANGCI_LINT_CACHE='$(GOLANGCI_LINT_CACHE)' $(GOLANGCI_LINT) run $(GO_PACKAGES)
	sh scripts/check-policy.sh

policy:
	sh scripts/check-policy.sh

test:
	$(GO) test $(GO_PACKAGES)

vet:
	$(GO) vet $(GO_PACKAGES)

race:
	$(GO) test -race $(GO_PACKAGES)

vulncheck:
	$(GOVULNCHECK) $(GO_PACKAGES)

validate-config:
	@set -eu; work="$$(mktemp -d '$(RUNNER_WORKDIR)/validate.XXXXXX')"; \
		trap 'rm -rf "$$work"' EXIT; \
		$(GO) build -o "$$work/unifi-stubd" ./cmd/unifi-stubd; \
		"$$work/unifi-stubd" -validate -config packaging/linux/etc/unifi-stubd/config.yaml; \
		"$$work/unifi-stubd" -validate -config packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml; \
		"$$work/unifi-stubd" -dry-run-plan -config lab/stub/configs/hosts/stub-bridge-observe/config.yaml >/dev/null; \
		"$$work/unifi-stubd" -profile-validate tests/fixtures/profiles

integration-nonunifi-metrics:
	"$(BUILD_RUNNER_ENTRY)" --project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
		instance run --manifest build-runner/instances/controller-gate.yaml -- \
		env UNIFI_STUB_DOCKER_HOST_LAB_DIR='$(PROVIDER_CONSUMER_ROOT)/lab/stub' \
		bash lab/stub/scripts/run-controller-channel-matrix.sh

coverage:
	mkdir -p dist
	$(GO) test -covermode=atomic -coverpkg='$(COVERAGE_PACKAGES)' -coverprofile=dist/cover.out $(COVERAGE_TESTS)
	@total="$$($(GO) tool cover -func=dist/cover.out | awk '$$1 == "total:" { gsub(/%/, "", $$3); print $$3 }')"; \
		test -n "$$total"; \
		printf 'total coverage: %s%% (diagnostic only)\n' "$$total"
	$(GO) tool cover -html=dist/cover.out -o dist/coverage.html
	@printf 'coverage report written to dist/coverage.html\n'

package:
	$(PKG_ENV) PKG_FORMATS='$(PKG_FORMATS)' sh scripts/package.sh

package-deb:
	$(PKG_ENV) sh scripts/package.sh deb

package-rpm:
	$(PKG_ENV) sh scripts/package.sh rpm

package-arch:
	$(PKG_ENV) sh scripts/package.sh archlinux

package-tgz:
	$(PKG_ENV) sh scripts/package.sh tgz

package-freebsd-tgz:
	$(PKG_ENV) PKG_GOOS='freebsd' PKG_GOARCH='$(PKG_FREEBSD_GOARCH)' PKG_FORMATS='tgz' sh scripts/package.sh tgz

package-freebsd-pkg-repos:
	$(PKG_ENV) sh scripts/package-freebsd-pkg-repos.sh

package-repos:
	sh scripts/build-package-repos.sh

clean-dist:
	rm -rf dist

switch-emulation:
	$(GO) run ./cmd/unifi-stubd -profile us8

switch-payload:
	$(GO) run ./cmd/unifi-stubd -profile us8 -dry-run

fmt:
	$(GOLANGCI_LINT) fmt $(GO_PACKAGES)

fmt-check:
	@output="$$(gofmt -l $(patsubst %/...,%,$(GO_PACKAGES)))" || exit $$?; \
		test -z "$$output" || { printf '%s\n' "$$output"; exit 1; }
