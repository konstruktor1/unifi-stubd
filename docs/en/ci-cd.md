# Build, Checks, and Delivery

The `Makefile`, scripts, and workflows are the executable sources of truth.
This page only maps their stable entry points; it does not mirror individual
jobs, tool versions, runner labels, or artifact lists.

## Consumer entry points

All agent-run Go work uses the configured Build Runner as described in
[`AGENTS.md`](../../AGENTS.md). Do not invoke host Go or a Go-based Make target
as a fallback.

| Outcome | Entry point |
| --- | --- |
| Complete pre-merge gate | `make code-checker` |
| Runner-local product subgate | `make check` |
| Current Stable/RC controller-facing gate | `make integration-nonunifi-metrics` |
| Linux packages | `make package` |
| FreeBSD package set (compatibility-only; blocked as evidence) | `make package-freebsd-pkg-repos` |
| Static package repositories | `make package-repos` |

`make code-checker` is the sole complete repository gate. It runs the product
and security subgates in one prepared Build Runner instance and finishes with
Code Checker; its components are not separate verification candidates.

`make integration-nonunifi-metrics` is the sole controller-facing gate. It
resolves the latest published Stable and Release Candidate at run start and
runs the same candidate in both disposable lanes. Details and acceptance
conditions live in the [Docker controller lab](docker-lab.md). Full product
acceptance additionally requires the focused host-safety and no-fabrication
evidence defined by the [requirements](requirements.md).

## Automation ownership

- [CI](../../.github/workflows/ci.yml) owns repository checks, diagnostics, and
  the main-branch Linux package smoke path.
- [Controller Integration](../../.github/workflows/tier2-docker.yml) owns the
  manually dispatched controller gate. Its self-hosted runner owner must
  prepare the checked-out consumer's Build Runner binding after checkout.
- [Package Repositories](../../.github/workflows/package-pages.yml) owns Linux
  release artifacts and the package site.
- [FreeBSD Packaging](../../.github/workflows/freebsd-packaging.yml) owns the
  separately operated FreeBSD builder and optional publication path.

The checked-in CI and package workflows still contain host-Go steps and cannot
provide validation or release evidence until their owners supply an approved
CI-visible Build Runner route. The direct FreeBSD builder is a separate
external-machine compatibility path; it is not Build Runner evidence.

Release decisions and operator steps are documented in the
[development workflow](development.md) and [release checklist](release.md).
