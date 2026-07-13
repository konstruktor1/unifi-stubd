# Docker Controller Lab

The Docker lab under `lab/stub/` is the project-owned integration environment
for the Go stub. It uses three services:

- UniFi Network Application on `https://127.0.0.1:8443/`
- MongoDB for the controller
- Go inform/discovery proxy on the internal lab network

The default controller image is pinned to
`lscr.io/linuxserver/unifi-network-application:10.4.57-ls135`. Override
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

Empty controller volumes are initialized automatically through the same local
setup API flow used by the web UI. Local runs may reuse configured volumes after
the helper verifies the lab login. The Tier 2 workflow always removes all lab
volumes before and after its run. The setup helper never prints controller
responses, generated device credentials, API tokens, or adoption keys.

## Controller Gate

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
- One inform request and one forwarded discovery packet per mode through the Go
  proxy.
- Controller API login against the Docker UniFi Network Application.
- Controller `/status` version check for the pinned Docker image.
- Pending adoption visibility for the bridge-observe and gateway-smoke devices.
- Controller-triggered adoption through the controller API for both switch and
  gateway-shaped payloads.
- A default-key `setparam` response with authkey, cfgversion, and
  `use_aes_gcm=true`, captured only as sanitized presence flags.
- Persisted local stub adoption state with `STATE=connected`, authkey,
  cfgversion, and AES-GCM enabled, without printing the authkey.
- At least two post-adoption AES-GCM inform heartbeats per adopted
  switch/gateway test device.
- Three consecutive controller samples with `adopted=true`, `state=1`, and
  completed adoption.
- No gateway command, command result, duplicate, or pending result before and
  after the 75-second stability window expected from the current 10.4.57 pin.

The default loopback-only lab credentials are `admin` / `admin`; they are also
used when bootstrapping empty volumes. Override them only for a local lab
controller:

```sh
UNIFI_STUB_LAB_ADMIN_USER=admin \
UNIFI_STUB_LAB_ADMIN_PASSWORD=... \
make integration-docker
```

## Tier 2 Docker Gate

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

The repository provides the standalone, manually dispatched `Tier 2 Docker
Integration` workflow. It is independent of the standard `CI` and package
workflows and runs only `make integration-docker` on the trusted runner selected
by the `docker-dmz` and `unifi-stubd` labels. Start it for the exact ref that is
being validated:

```sh
gh workflow run tier2-docker.yml --ref dev
```

The equivalent local run is:

```sh
git switch dev
git pull --ff-only origin dev
make integration-docker
git rev-parse HEAD
```

Pass criteria:

- `make integration-docker` exits `0` and prints `docker integration: ok`.
- The log shows the pinned UniFi Network Application version check.
- The log shows the generated bridge, port-map, and gateway throwaway
  identities.
- Switch and gateway gates both record discovery, sanitized `setparam`, two GCM
  heartbeats, complete persisted adoption state, and stable connected samples.
- The gateway gate matches the explicit controller-command expectation before
  and after a 75-second stability window. The current 10.4.57 pin expects no
  command; compatibility runs can set
  `UNIFI_STUB_DOCKER_GATEWAY_COMMAND_EXPECTATION=required` for controllers
  known to issue `wg-genkeypair`.
- The final gateway status reports the controller's positive interval, clamped
  to local bounds, as effective and next delay, with zero consecutive failures
  and no active backoff.

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

The workflow has only a `workflow_dispatch` trigger, read-only repository
permissions, one non-cancelling concurrency group, and a 30-minute timeout. Its
preflight verifies Linux/X64, Go, Docker daemon access, Docker Compose v2, and
the combined Compose configuration before any lab service starts. It uploads no
captures or controller state. The standard `CI`, package workflows, branch
rules, and required checks do not depend on this Tier 2 gate. FreeBSD packaging
is separately manual Tier 3. The old FreeBSD proxy runner must not carry the
`docker-dmz` label, so Tier 2 requires a different dedicated Docker runner.

## Cleanup Semantics

The script derives throwaway MAC/IP identities for every run, stops and removes
temporary stub containers and volumes, and asks the controller to delete any
adopted state for the test MACs. For local runs, controller volumes are not reset
and the controller delete request is best-effort; tests must treat fresh
throwaway MACs as the reliable isolation boundary. The Tier 2 workflow provides
the stronger CI boundary: it removes the complete Compose project, all lab
volumes, orphans, and ignored captures before and after every run, including
failed runs.

At the beginning of a standard run, the helper also removes adopted devices
with the reserved Docker-fixture MAC prefix `02:15:6d`. This makes the gate
self-recover from an earlier `UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` run without
touching devices outside the project-owned fixture prefix.

If UniFi returns `api.err.DeviceBusy` while deleting one of those reserved
fixtures, the gate restarts only the disposable controller service and retries
the normal controller API deletion. The controller and MongoDB volumes are not
reset.

UniFi Network can keep non-adopted Pending rows in process memory until its
discovery TTL expires. Those rows are not persisted in MongoDB in the observed
Docker lab. Fresh throwaway MACs avoid collisions between repeated runs. In the
lab helper, `wait-clean` means "not adopted"; use `wait-absent` only when a
test truly needs the row to disappear.

Set `UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` only when you intentionally want to
inspect the adopted test device or stub state volume after a failing run.

## Boundaries

This Tier 2 lab proves container-local Linux bridge/FDB observation, sysfs counters,
explicit port mapping, gateway table rendering, discovery forwarding, inform
framing, controller adoption, local adoption state persistence, and AES-GCM.
It does not prove Proxmox host bridge behavior, FreeBSD runtime behavior, LLDP
import, or event subscriptions.

It also does not prove every physical topology. Container tests use throwaway
identities, while real Proxmox or bridge deployments should validate directed
`uplink_neighbor`, optional `remote_port`, `uplink_port`, and real interface
MAC reporting against the target controller before the result is treated as
representative.

Firmware research stacks under `lab/gateway-profiles`, real Proxmox/Linux
bridges, FreeBSD/OPNsense hosts, deployed controllers, and physical UniFi
devices remain Tier 3 tests. FreeBSD package creation on the dedicated
SSH-proxy builder is also Tier 3, not part of this Docker lab.
