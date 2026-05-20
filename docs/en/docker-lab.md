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
  `if_table`, `network_table`, `config_port_table`, `ethernet_overrides`, and
  `reported_networks` from the shared port view.
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

## Cleanup Semantics

The script derives throwaway MAC/IP identities for every run, stops and removes
temporary stub containers and volumes, and asks the controller to delete any
adopted state for the test MACs. Controller volumes are not reset.

UniFi Network can keep non-adopted Pending rows in process memory until its
discovery TTL expires. Those rows are not persisted in MongoDB in the observed
Docker lab. Fresh throwaway MACs avoid collisions between repeated runs.

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
