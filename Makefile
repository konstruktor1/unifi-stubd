GO ?= go
GOLANGCI_LINT ?= $(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint
NFPM ?= $(GO) tool github.com/goreleaser/nfpm/v2/cmd/nfpm
PKG_VERSION ?= 0.1.0
PKG_RELEASE ?= 1
PKG_GOOS ?= linux
PKG_GOARCH ?= $(shell $(GO) env GOARCH)
PKG_FREEBSD_GOARCH ?= amd64
PKG_FORMATS ?= deb rpm archlinux tgz
PKG_LICENSE ?= AGPL-3.0-or-later
PKG_MAINTAINER ?= unifi-stubd maintainers <info@spinas.org>
BUILD_LDFLAGS := -s -w -X main.version=$(PKG_VERSION)

PKG_ENV_NONFPM := PKG_VERSION='$(PKG_VERSION)' \
  PKG_RELEASE='$(PKG_RELEASE)' \
  PKG_GOOS='$(PKG_GOOS)' \
  PKG_GOARCH='$(PKG_GOARCH)' \
  PKG_LICENSE='$(PKG_LICENSE)' \
  PKG_MAINTAINER='$(PKG_MAINTAINER)' \
  BUILD_LDFLAGS='$(BUILD_LDFLAGS)'

PKG_ENV := NFPM='$(NFPM)' \
  $(PKG_ENV_NONFPM)

.PHONY: build build-freebsd check clean-dist coverage fmt help lint package package-arch package-deb package-freebsd-tgz package-rpm package-tgz policy switch-emulation switch-payload test validate-config

help:
	@printf '%s\n' \
		'Targets:' \
		'  make check        Run lint, policy checks, and tests' \
		'  make build-freebsd  Cross-build the FreeBSD/OPNsense binary' \
		'  make coverage     Generate HTML coverage report in dist/coverage.html' \
		'  make lint         Run golangci-lint and repository policy checks' \
		'  make test         Run all Go tests' \
		'  make validate-config  Validate packaged configs and example profiles' \
		'  make switch-payload  Print discovery and inform payloads' \
		'  make switch-emulation  Start the lab switch emulator' \
		'  make package      Build deb, rpm, archlinux, and tgz packages' \
		'  make package-freebsd-tgz  Build FreeBSD/OPNsense stub-only tgz' \
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

validate-config:
	$(GO) run ./cmd/unifi-stubd -validate -config packaging/linux/etc/unifi-stubd/config.yaml
	$(GO) run ./cmd/unifi-stubd -validate -config packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml
	$(GO) run ./cmd/unifi-stubd -profile-validate tests/fixtures/profiles

coverage:
	$(GO) test -coverprofile=dist/cover.out ./...
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

clean-dist:
	rm -rf dist

switch-emulation:
	$(GO) run ./cmd/unifi-stubd

switch-payload:
	$(GO) run ./cmd/unifi-stubd -dry-run

fmt:
	$(GOLANGCI_LINT) fmt ./...
