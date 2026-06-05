# Docker Controller Lab

The Docker lab under `lab/stub/` is the project-owned integration environment
for the Go stub. It reuses three long-lived services:

- UniFi Network Application on `https://127.0.0.1:8443/`
- MongoDB for the controller
- inform MITM on the internal lab network

The default controller image is pinned to
`lscr.io/linuxserver/unifi-network-application:10.3.58-ls129`. Override
`UNIFI_NETWORK_IMAGE` only when intentionally validating another controller
version; set `UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION` with that controller's
`/status` `server_version` when the integration test should enforce it.

The integration overlay `lab/stub/compose.tests.yaml` adds temporary
`stub-bridge-observe`, `stub-port-map`, and `stub-gateway-smoke` services. They
are built from the current repository checkout and are removed again by the
test harness.

Project-owned lab defaults live in
`lab/stub/configs/hosts/<hostname>/config.yaml`, with one directory per
reported stub hostname, and are mounted read-only into the stub containers. The
test harness still passes throwaway MAC/IP/profile/hostname values as CLI
overrides, so these files stay stable and do not contain controller state or
secrets.

## Smoke Test

Run from the repository root:

```sh
make integration-docker
```

The target verifies:

- Compose configuration for the base lab plus the test overlay.
- Runtime image build, including `iproute2` for bridge/FDB observation.
- `bridge-observe` dry-run payload from a container-local Linux bridge.
- `management_lan.mode: preexisting-interface` dry-run payload against the
  container `eth0` address, proving the new switch management LAN config path.
- `port-map` dry-run payload from container-local veth interfaces.
- Gateway dry-run payload from the `uxg-lite` profile, including
  `if_table`, `network_table`, read-only physical `port_table`,
  `reported_networks`, `uplink_table`, and `wan1` from the shared port view.
- One inform request per mode through the MITM.
- Controller API login against the Docker UniFi Network Application.
- Controller `/status` version check for the pinned Docker image.
- Pending adoption visibility for the bridge-observe and gateway-smoke devices.
- Controller-triggered adoption through the controller API for both switch and
  gateway-shaped payloads.
- Persisted local stub adoption state with `STATE=connected` and an authkey
  present, without printing the authkey.
- At least one post-adoption inform heartbeat per adopted switch/gateway test
  device through the MITM.

The default lab credentials are `admin` / `admin`. Override them only for a
local lab controller:

```sh
UNIFI_STUB_LAB_ADMIN_USER=admin \
UNIFI_STUB_LAB_ADMIN_PASSWORD=... \
make integration-docker
```

## Dev To Main Docker Gate

Use the Docker lab as the standard controller-compatibility gate before
promoting `dev` to `main` whenever the accumulated `dev` changes touch any of
these areas:

- inform framing, encryption, compression, or adoption responses;
- controller-visible payload shape, profile data, port rendering, gateway
  tables, or WAN health;
- observation modes that affect rendered controller state;
- Docker lab fixtures, controller image pinning, or adoption test helpers.

The gate is tied to a commit, not to a branch name. The commit promoted from
`dev` to `main` must be the same commit that passed the Docker gate, or a
descendant that only contains unrelated documentation or release metadata.

Standard manual run:

```sh
git switch dev
git pull --ff-only origin dev
make check
make integration-docker
git rev-parse HEAD
```

Pass criteria:

- `make check` exits `0`.
- `make integration-docker` exits `0` and prints `docker integration: ok`.
- The log shows the pinned UniFi Network Application version check.
- The log shows the generated bridge, port-map, and gateway throwaway
  identities.
- Switch and gateway adoption smoke tests both reach connected local state and
  at least one post-adoption inform heartbeat.

Promotion evidence should record:

- commit SHA;
- command used;
- controller image and expected controller version;
- start and end time;
- final exit code;
- link to a GitHub Actions run or preserved local terminal log.

Do not promote `dev` to `main` while the Docker gate is failing, skipped for a
controller-facing change, or run against a different commit. If the failure
needs inspection, rerun once with `UNIFI_STUB_DOCKER_KEEP_RESOURCES=1`, inspect
the temporary device and state volume, then remove the resources before the next
standard run.

Automation target:

1. Keep the manual gate as the baseline until the Docker lab runner is stable.
2. Add a separate `Docker Integration` GitHub Actions workflow with
   `workflow_dispatch` and `pull_request` for `dev` to `main`.
3. Run only `make integration-docker` in that workflow after the normal `CI /
   check` is green.
4. Use a single concurrency group for the Docker controller lab so two adoption
   tests cannot share controller state at the same time.
5. Once stable, require the `Docker Integration` status check on `dev` to
   `main` pull requests that include controller-facing changes.

The standard `CI` workflow remains the fast gate. The Docker gate is the
controller-compatibility gate, and the `main` package job remains the package
install smoke gate.

## Cleanup Semantics

The script derives throwaway MAC/IP identities for every run, stops and removes
temporary stub containers and volumes, and asks the controller to delete any
adopted state for the test MACs. Controller volumes are not reset. The
controller delete request is best-effort; tests must treat fresh throwaway MACs
as the reliable isolation boundary.

UniFi Network can keep non-adopted Pending rows in process memory until its
discovery TTL expires. Those rows are not persisted in MongoDB in the observed
Docker lab. Fresh throwaway MACs avoid collisions between repeated runs. In the
lab helper, `wait-clean` means "not adopted"; use `wait-absent` only when a
test truly needs the row to disappear.

Set `UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` only when you intentionally want to
inspect the adopted test device or stub state volume after a failing run.

## Boundaries

This lab proves container-local Linux bridge/FDB observation, sysfs counters,
explicit port mapping, gateway table rendering, inform framing, controller
adoption, and local adoption state persistence. It does not prove Proxmox host
bridge behavior, FreeBSD runtime behavior, LLDP import, or event subscriptions.

It also does not prove physical-topology direction. Container tests use
throwaway synthetic identities, so they do not cover the case where a real
upstream UniFi switch already reports the same physical host MAC. Real Proxmox
or bridge deployments should validate `uplink_neighbor`, `uplink_port`, and
synthetic-versus-physical MAC selection against the target controller before the
result is treated as representative.
