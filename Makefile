GO ?= go
GOLANGCI_LINT ?= $(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOLANGCI_LINT_CACHE ?= $(GOCACHE)/golangci-lint
RUNNER_WORKDIR ?= $(GOCACHE)/unifi-stubd-work
NFPM ?= $(GO) tool github.com/goreleaser/nfpm/v2/cmd/nfpm
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
PKG_ACCEPTED_RUN_ID ?=
PKG_LICENSE ?= AGPL-3.0-or-later
PKG_MAINTAINER ?= unifi-stubd maintainers <info@spinas.org>
BUILD_LDFLAGS := -s -w -X main.version=$(PKG_VERSION)
BUILD_RUNNER_PROJECT_NAME := unifi-stubd
BUILD_RUNNER_INTEGRATION_MANIFEST := build-runner/instances/controller-gate.yaml
CODE_CHECKER_CONFIG := $(CURDIR)/code-checker.toml
CODE_CHECKER_MANIFEST := build-runner/instances/code-checker.yaml

ifneq ($(strip $(GOCACHE)),)
GOTMPDIR ?= $(RUNNER_WORKDIR)
TMPDIR ?= $(RUNNER_WORKDIR)
export GOTMPDIR TMPDIR
endif

PKG_ENV_NONFPM := PKG_VERSION='$(PKG_VERSION)' \
  PKG_RELEASE='$(PKG_RELEASE)' \
  PKG_GOOS='$(PKG_GOOS)' \
  PKG_GOARCH='$(PKG_GOARCH)' \
  PKG_ACCEPTED_RUN_ID='$(PKG_ACCEPTED_RUN_ID)' \
  PKG_LICENSE='$(PKG_LICENSE)' \
  PKG_MAINTAINER='$(PKG_MAINTAINER)' \
  BUILD_LDFLAGS='$(BUILD_LDFLAGS)'

PKG_ENV := NFPM='$(NFPM)' \
  $(PKG_ENV_NONFPM)

.PHONY: anti-bloat-changed anti-bloat-full build build-freebsd check clean-dist code-checker code-checker-ready coverage fmt help integration-nonunifi-metrics lint package package-arch package-deb package-freebsd-pkg-repos package-freebsd-tgz package-repos package-rpm package-tgz policy race runner-workdir switch-emulation switch-payload test validate-config vulncheck

build build-freebsd coverage fmt lint package package-arch package-deb package-freebsd-tgz package-rpm package-tgz race switch-emulation switch-payload test validate-config vulncheck: | runner-workdir

runner-workdir:
	@test -n '$(GOCACHE)' || { printf '%s\n' 'GOCACHE must be supplied by Build Runner' >&2; exit 1; }
	@mkdir -p '$(RUNNER_WORKDIR)'

help:
	@printf '%s\n' \
		'Targets:' \
		'  make check        Run lint, policy checks, and tests' \
		'  make code-checker  Run the independent Code Checker gate through Build Runner' \
		'  make anti-bloat-changed  Print the subscribed changed-scope instruction' \
		'  make anti-bloat-full  Print the subscribed full-scope instruction' \
		'  make build-freebsd  Cross-build the FreeBSD/OPNsense binary' \
		'  make coverage     Write a diagnostic coverage report' \
		'  make lint         Run golangci-lint and repository policy checks' \
		'  make race         Run all Go tests with the race detector' \
		'  make test         Run all Go tests' \
		'  make validate-config  Validate packaged configs and example profiles' \
		'  make vulncheck    Run Go vulnerability scanning' \
		'  make integration-nonunifi-metrics  Prove one non-UniFi bridge on current Stable and RC' \
		'  make switch-payload  Print discovery and inform payloads' \
		'  make switch-emulation  Start the lab switch emulator' \
		'  make package      Build deb, rpm, archlinux, and tgz packages' \
		'  make package-repos  Build unsigned alpha package repositories in dist/package-site' \
		'  make package-freebsd-tgz  Build FreeBSD/OPNsense stub-only tgz' \
		'  make package-freebsd-pkg-repos  Build FreeBSD tarballs and pkg repositories through the configured FreeBSD builder' \
		'  make clean-dist   Remove package build output'

build:
	$(GO) build -ldflags='$(BUILD_LDFLAGS)' ./...

build-freebsd:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=freebsd GOARCH='$(PKG_FREEBSD_GOARCH)' $(GO) build -trimpath -ldflags='$(BUILD_LDFLAGS)' -o dist/unifi-stubd_freebsd_$(PKG_FREEBSD_GOARCH) ./cmd/unifi-stubd

check: lint validate-config test

code-checker-ready:
	@test -n '$(CODE_CHECKER_DIR)' || { printf '%s\n' 'CODE_CHECKER_DIR must name the Code Checker checkout' >&2; exit 2; }
	@test -x '$(CODE_CHECKER_DIR)/bin/code-checker' || { printf '%s\n' 'Code Checker launcher is not executable' >&2; exit 2; }
	@test -f '$(CODE_CHECKER_CONFIG)' || { printf '%s\n' 'code-checker.toml is missing' >&2; exit 2; }

code-checker: code-checker-ready
	@: "$${BUILD_RUNNER_DIR:?set BUILD_RUNNER_DIR to the provider checkout}"
	BUILD_RUNNER_HOST_PATH_CODE_CHECKER_SOURCE='$(CODE_CHECKER_DIR)' \
		"$$BUILD_RUNNER_DIR/build-runner" --project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
		instance run --manifest '$(CODE_CHECKER_MANIFEST)' -- \
		/artifacts/code-checker-source/bin/code-checker \
		--config /workspace/code-checker.toml check /workspace --require-tools

anti-bloat-changed: code-checker-ready
	'$(CODE_CHECKER_DIR)/bin/code-checker' --config '$(CODE_CHECKER_CONFIG)' \
		anti-bloat plan '$(CURDIR)' --scope changed --json

anti-bloat-full: code-checker-ready
	'$(CODE_CHECKER_DIR)/bin/code-checker' --config '$(CODE_CHECKER_CONFIG)' \
		anti-bloat plan '$(CURDIR)' --scope full --json

lint:
	GOLANGCI_LINT_CACHE='$(GOLANGCI_LINT_CACHE)' $(GOLANGCI_LINT) config verify
	GOLANGCI_LINT_CACHE='$(GOLANGCI_LINT_CACHE)' $(GOLANGCI_LINT) run ./...
	sh scripts/check-policy.sh

policy:
	sh scripts/check-policy.sh

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

vulncheck:
	$(GOVULNCHECK) ./...

validate-config:
	@work="$$(mktemp -d '$(RUNNER_WORKDIR)/validate.XXXXXX')"; \
		trap 'rm -rf "$$work"' EXIT; \
		$(GO) build -o "$$work/unifi-stubd" ./cmd/unifi-stubd; \
		"$$work/unifi-stubd" -validate -config packaging/linux/etc/unifi-stubd/config.yaml; \
		"$$work/unifi-stubd" -validate -config packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml; \
		"$$work/unifi-stubd" -dry-run-plan -config lab/stub/configs/hosts/stub-bridge-observe/config.yaml >/dev/null; \
		"$$work/unifi-stubd" -profile-validate tests/fixtures/profiles

define run_product_gate
	@if [ "$$(pwd -P)" = /workspace ]; then \
		$(1); \
	else \
		: "$${BUILD_RUNNER_DIR:?set BUILD_RUNNER_DIR to the provider checkout}"; \
		host_lab_dir="$$(pwd -P)/lab/stub"; \
		"$$BUILD_RUNNER_DIR/build-runner" \
			--project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
			instance run \
			--manifest '$(BUILD_RUNNER_INTEGRATION_MANIFEST)' \
			-- env \
				UNIFI_STUB_DOCKER_HOST_LAB_DIR="$$host_lab_dir" \
				$(1); \
	fi
endef

integration-nonunifi-metrics:
	$(call run_product_gate,bash lab/stub/scripts/run-controller-channel-matrix.sh)

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
	$(PKG_ENV_NONFPM) sh scripts/package.sh tgz

package-freebsd-tgz:
	$(PKG_ENV_NONFPM) PKG_GOOS='freebsd' PKG_GOARCH='$(PKG_FREEBSD_GOARCH)' PKG_FORMATS='tgz' sh scripts/package.sh tgz

package-freebsd-pkg-repos:
	$(PKG_ENV_NONFPM) sh scripts/package-freebsd-pkg-repos.sh

package-repos:
	sh scripts/build-package-repos.sh

clean-dist:
	rm -rf dist

switch-emulation:
	$(GO) run ./cmd/unifi-stubd -profile us8

switch-payload:
	$(GO) run ./cmd/unifi-stubd -profile us8 -dry-run

fmt:
	$(GOLANGCI_LINT) fmt ./...
