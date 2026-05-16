GO ?= go
GOLANGCI_LINT ?= $(GO) tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint
NFPM ?= $(GO) tool github.com/goreleaser/nfpm/v2/cmd/nfpm
PKG_VERSION ?= 0.1.0
PKG_RELEASE ?= 1
PKG_GOOS ?= linux
PKG_GOARCH ?= $(shell $(GO) env GOARCH)
PKG_FORMATS ?= deb rpm archlinux tgz
PKG_LICENSE ?= AGPL-3.0-or-later
PKG_MAINTAINER ?= unifi-stubd maintainers <info@spinas.org>

.PHONY: build check clean-dist fmt help lint package package-arch package-deb package-rpm package-tgz policy switch-emulation switch-payload test

help:
	@printf '%s\n' \
		'Targets:' \
		'  make check        Run lint, policy checks, and tests' \
		'  make lint         Run golangci-lint and repository policy checks' \
		'  make test         Run tests under tests/' \
		'  make switch-payload  Print discovery and inform payloads' \
		'  make switch-emulation  Start the lab switch emulator' \
		'  make package      Build deb, rpm, archlinux, and tgz packages' \
		'  make clean-dist   Remove package build output'

build:
	$(GO) build ./...

check: lint test

lint:
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) run ./...
	sh scripts/check-policy.sh

policy:
	sh scripts/check-policy.sh

test:
	$(GO) test ./tests/...

package:
	NFPM='$(NFPM)' PKG_VERSION='$(PKG_VERSION)' PKG_RELEASE='$(PKG_RELEASE)' PKG_GOOS='$(PKG_GOOS)' PKG_GOARCH='$(PKG_GOARCH)' PKG_FORMATS='$(PKG_FORMATS)' PKG_LICENSE='$(PKG_LICENSE)' PKG_MAINTAINER='$(PKG_MAINTAINER)' sh scripts/package.sh

package-deb:
	NFPM='$(NFPM)' PKG_VERSION='$(PKG_VERSION)' PKG_RELEASE='$(PKG_RELEASE)' PKG_GOOS='$(PKG_GOOS)' PKG_GOARCH='$(PKG_GOARCH)' PKG_LICENSE='$(PKG_LICENSE)' PKG_MAINTAINER='$(PKG_MAINTAINER)' sh scripts/package.sh deb

package-rpm:
	NFPM='$(NFPM)' PKG_VERSION='$(PKG_VERSION)' PKG_RELEASE='$(PKG_RELEASE)' PKG_GOOS='$(PKG_GOOS)' PKG_GOARCH='$(PKG_GOARCH)' PKG_LICENSE='$(PKG_LICENSE)' PKG_MAINTAINER='$(PKG_MAINTAINER)' sh scripts/package.sh rpm

package-arch:
	NFPM='$(NFPM)' PKG_VERSION='$(PKG_VERSION)' PKG_RELEASE='$(PKG_RELEASE)' PKG_GOOS='$(PKG_GOOS)' PKG_GOARCH='$(PKG_GOARCH)' PKG_LICENSE='$(PKG_LICENSE)' PKG_MAINTAINER='$(PKG_MAINTAINER)' sh scripts/package.sh archlinux

package-tgz:
	PKG_VERSION='$(PKG_VERSION)' PKG_RELEASE='$(PKG_RELEASE)' PKG_GOOS='$(PKG_GOOS)' PKG_GOARCH='$(PKG_GOARCH)' PKG_LICENSE='$(PKG_LICENSE)' PKG_MAINTAINER='$(PKG_MAINTAINER)' sh scripts/package.sh tgz

clean-dist:
	rm -rf dist

switch-emulation:
	$(GO) run ./cmd/unifi-stubd

switch-payload:
	$(GO) run ./cmd/unifi-stubd -dry-run

fmt:
	$(GOLANGCI_LINT) fmt ./...
