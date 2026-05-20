# Testing Guide

This page describes which test level to use for each kind of change.

## Standard Local Gate

Run before committing:

```sh
make check
git diff --check
```

`make check` verifies lint configuration, runs lint, enforces repository policy,
validates packaged config/profile YAML, and runs `go test ./...`.

## Docker Controller Gate

Run when adoption, inform, payload shape, profiles, or controller compatibility
changes:

```sh
make integration-docker
```

The Docker lab is the reference controller path. It validates the pinned UniFi
Network Application container, pending adoption, controller-triggered adoption,
persisted local state, and selected dry-run payloads.

## Package Gate

Run when packaged configs, service files, filesystem paths, or release metadata
change:

```sh
make package
make package-freebsd-tgz
```

Inspect package contents only unless the test explicitly calls for temporary
installation. No permanent service enablement belongs in a package smoke test.

## Real Linux Bridge Gate

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

## Real FreeBSD/OPNsense Gate

Use for FreeBSD parsing and runtime smoke tests. Do not install permanently.

Read-only preflight:

```sh
ifconfig
ifconfig <bridge> addr
tail /var/log/messages
```

Use temporary extraction and temporary state paths. Missing tools should be
documented as skipped, not installed during this gate.

## Controller Lab Gate

Use only with disposable MACs and explicit cleanup.

Checklist:

- dry-run first;
- no real MAC/IP collisions;
- one disposable device per profile test;
- controller forget/remove after test;
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

