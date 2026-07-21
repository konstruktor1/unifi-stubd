GO ?= go
GOLANGCI_LINT ?= $(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint
NFPM ?= $(GO) tool github.com/goreleaser/nfpm/v2/cmd/nfpm
GOVULNCHECK ?= $(GO) tool golang.org/x/vuln/cmd/govulncheck
COVERAGE_MIN ?= 60.0
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
BUILD_RUNNER_INTEGRATION_MANIFEST := build-runner/instances/controller-gate.yaml

PKG_ENV_NONFPM := PKG_VERSION='$(PKG_VERSION)' \
  PKG_RELEASE='$(PKG_RELEASE)' \
  PKG_GOOS='$(PKG_GOOS)' \
  PKG_GOARCH='$(PKG_GOARCH)' \
  PKG_LICENSE='$(PKG_LICENSE)' \
  PKG_MAINTAINER='$(PKG_MAINTAINER)' \
  BUILD_LDFLAGS='$(BUILD_LDFLAGS)'

PKG_ENV := NFPM='$(NFPM)' \
  $(PKG_ENV_NONFPM)

.PHONY: build build-freebsd check clean-dist coverage fmt help integration-docker lint package package-arch package-deb package-freebsd-pkg-repos package-freebsd-tgz package-repos package-rpm package-tgz policy switch-emulation switch-payload test validate-config vulncheck

help:
	@printf '%s\n' \
		'Targets:' \
		'  make check        Run lint, policy checks, and tests' \
		'  make build-freebsd  Cross-build the FreeBSD/OPNsense binary' \
		'  make coverage     Enforce coverage and write dist/coverage.html' \
		'  make lint         Run golangci-lint and repository policy checks' \
		'  make test         Run all Go tests' \
		'  make validate-config  Validate packaged configs and example profiles' \
		'  make vulncheck    Run Go vulnerability scanning' \
		'  make integration-docker  Run the pinned Docker controller adoption gate' \
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

lint:
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) run ./...
	sh scripts/check-policy.sh

policy:
	sh scripts/check-policy.sh

test:
	$(GO) test ./...

vulncheck:
	$(GOVULNCHECK) ./...

validate-config:
	$(GO) run ./cmd/unifi-stubd -validate -config packaging/linux/etc/unifi-stubd/config.yaml
	$(GO) run ./cmd/unifi-stubd -validate -config packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml
	$(GO) run ./cmd/unifi-stubd -dry-run-plan -config lab/stub/configs/hosts/stub/config.yaml >/dev/null
	$(GO) run ./cmd/unifi-stubd -dry-run-plan -config lab/stub/configs/hosts/stub-bridge-observe/config.yaml >/dev/null
	$(GO) run ./cmd/unifi-stubd -dry-run-plan -config lab/stub/configs/hosts/stub-port-map/config.yaml >/dev/null
	$(GO) run ./cmd/unifi-stubd -dry-run-plan -config lab/stub/configs/hosts/stub-gateway-smoke/config.yaml >/dev/null
	$(GO) run ./cmd/unifi-stubd -dry-run-plan -config lab/stub/configs/hosts/server-lan1-sfp-lab/config.example.yaml >/dev/null
	$(GO) run ./cmd/unifi-stubd -dry-run-plan -config lab/stub/configs/hosts/opnsense-uxg-sfp-lab/config.example.yaml >/dev/null
	tmp="$$(mktemp)"; trap 'rm -f "$$tmp"' EXIT; $(GO) run ./cmd/unifi-stubd -dry-run -config lab/stub/configs/hosts/opnsense-uxg-sfp-lab/config.example.yaml > "$$tmp"; $(GO) run ./lab/stub/tools/assert-payload opnsense-uxg "$$tmp"
	$(GO) run ./cmd/unifi-stubd -profile-validate tests/fixtures/profiles

define run_docker_integration
	@if [ "$$(pwd -P)" = /workspace ]; then \
		$(1); \
	else \
		: "$${BUILD_RUNNER_DIR:?set BUILD_RUNNER_DIR to the provider checkout}"; \
		host_lab_dir="$$(pwd -P)/lab/stub"; \
		"$$BUILD_RUNNER_DIR/build-runner" \
			--project-name '$(BUILD_RUNNER_PROJECT_NAME)' \
			instance run \
			--manifest '$(BUILD_RUNNER_INTEGRATION_MANIFEST)' \
			-- env UNIFI_STUB_DOCKER_HOST_LAB_DIR="$$host_lab_dir" $(1); \
	fi
endef

integration-docker:
	$(call run_docker_integration,sh lab/stub/scripts/run-docker-integration.sh)

coverage:
	mkdir -p dist
	$(GO) test -covermode=atomic -coverpkg='$(COVERAGE_PACKAGES)' -coverprofile=dist/cover.out $(COVERAGE_TESTS)
	@total="$$($(GO) tool cover -func=dist/cover.out | awk '$$1 == "total:" { gsub(/%/, "", $$3); print $$3 }')"; \
		test -n "$$total"; \
		printf 'total coverage: %s%% (minimum $(COVERAGE_MIN)%%)\n' "$$total"; \
		if ! awk -v total="$$total" -v minimum='$(COVERAGE_MIN)' 'BEGIN { exit !(total + 0 >= minimum + 0) }'; then \
			printf 'coverage %.1f%% is below minimum %.1f%%\n' "$$total" '$(COVERAGE_MIN)' >&2; \
			exit 1; \
		fi
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
	$(GO) run ./cmd/unifi-stubd

switch-payload:
	$(GO) run ./cmd/unifi-stubd -dry-run

fmt:
	$(GOLANGCI_LINT) fmt ./...
