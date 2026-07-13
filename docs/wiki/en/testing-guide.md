# Testing Guide

This page describes which test level to use for each kind of change.

- Tier 1 covers deterministic lint, policy, unit, fixture, configuration, and
  package-build checks.
- Tier 2 is the pinned, container-only UniFi controller adoption gate under
  `lab/stub`.
- Tier 3 covers firmware research containers and tests against real bridges,
  hosts, controllers, or physical UniFi devices.

## Tier 1: Standard Local Gate

Run before committing:

```sh
make check
git diff --check
```

`make check` verifies lint configuration, runs lint, enforces repository policy,
validates packaged config/profile YAML, and runs `go test ./...`.

## Tier 2: Docker Controller Gate

Run when adoption, inform, payload shape, profiles, or controller compatibility
changes:

```sh
make integration-docker
```

The Docker lab is the reference controller path. It validates the pinned UniFi
Network Application container, pending adoption, controller-triggered adoption,
persisted local state, and selected dry-run payloads.

Run the standalone Tier 2 workflow manually for an exact trusted ref:

```sh
gh workflow run tier2-docker.yml --ref dev
```

This workflow is independent of the standard `CI` and package workflows. It
starts with fresh controller volumes, always removes its lab state, and uploads
no captures or controller state.

For `dev` to `main` promotion, this gate is standardized in
[`docs/en/docker-lab.md`](../../en/docker-lab.md). The promoted `dev` commit
must be the commit that passed `make integration-docker`, or a descendant that
only contains unrelated documentation or release metadata. Record the commit
SHA, controller image/version, command, final exit code, and log or Actions run
URL before merging to `main`.

## Package Gate

Run when packaged configs, service files, filesystem paths, or release metadata
change:

```sh
make package
make package-freebsd-tgz
```

Inspect package contents only unless the test explicitly calls for temporary
installation. No permanent service enablement belongs in a package smoke test.

## Tier 3: Real Linux Bridge Gate

Use for `bridge-observe` behavior that Docker cannot prove, especially
topology direction, physical uplinks, SFP/SFP+ placement, and upstream UniFi
switch interactions.

Read-only preflight:

```sh
ip link show
bridge fdb show br <bridge>
cat /sys/class/net/<iface>/speed
cat /proc/net/dev
```

Then run temporary dry-run or one controlled controller test with disposable
MAC/state.

## Tier 3: Real FreeBSD/OPNsense Gate

Use for FreeBSD parsing and runtime smoke tests. Do not install permanently.

Read-only preflight:

```sh
ifconfig
ifconfig <bridge> addr
tail /var/log/messages
```

Use temporary extraction and temporary state paths. Missing tools should be
documented as skipped, not installed during this gate.

## Tier 3: Real Controller Lab Gate

Use only with disposable MACs and explicit cleanup.

Checklist:

- dry-run first;
- no real MAC/IP collisions;
- one disposable device per profile test;
- controller forget/remove after test, treated as best-effort;
- local state cleanup after stop;
- no controller tokens, private URLs, or real MACs committed.

## What Each Test Proves

| Test | Proves | Does not prove |
| --- | --- | --- |
| `go test ./...` | unit and fixture behavior | live controller compatibility |
| `make check` | lint, policy, validation, tests | package install behavior |
| `make integration-docker` | pinned controller adoption path | physical topology direction |
| package builds | artifacts are buildable | target host runtime correctness |
| real Linux bridge | physical bridge observation | FreeBSD behavior |
| real FreeBSD host | FreeBSD runtime basics | Linux bridge behavior |
| real controller | adoption against deployed controller | safety under MAC/IP collisions |

Firmware research stacks under `lab/gateway-profiles` are also Tier 3 because
they require ignored firmware/rootfs inputs, architecture-specific execution,
or mock hardware state.
