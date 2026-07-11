# Generic Stub Compose Lab

This is the main Docker lab for the Go `unifi-stubd` daemon. It starts a UniFi
Network Application, MongoDB, a Go inform/discovery proxy, and one `stub`
container built from the repository root. Use it when validating discovery,
inform, adoption state, or profile payload behavior in the Go service.

The default UniFi Network Application image is pinned to
`lscr.io/linuxserver/unifi-network-application:10.4.57-ls135`. Override
`UNIFI_NETWORK_IMAGE` only for explicit compatibility work.

The lab is deliberately named `stub` everywhere: Compose service, container
name, hostname, and persistent volume. That keeps the generic daemon lab
separate from firmware research directories, where containers are wrappers
around extracted vendor root filesystems.

Runtime configuration for the long-lived `stub` service and the temporary test
services lives in `configs/hosts/<hostname>/config.yaml`, with one directory
per reported stub hostname. Compose mounts `configs/` read-only inside the
containers at
`/usr/local/share/unifi-stubd-lab/configs`. Start scripts still pass throwaway
MAC/IP/profile/hostname values as CLI overrides when a test run needs fresh
identities.

Real-network host configs and temporary snapshots can use the same host
directory shape under `configs/hosts/<hostname>/real/config.yaml` or
`configs/hosts/<hostname>/temp/config.yaml`. Those paths are local-only and
ignored by Git; keep committed examples sanitized.

Captured inform traffic is local output and belongs in the ignored
`captures/` directory. Do not commit raw controller captures, adoption keys,
tokens, private URLs, or device-specific data from this lab.

## Docker Controller Gate

The Docker integration gate runs against the pinned UniFi Network Application.
Its Go controller helper completes the same local setup flow as the web UI when
the controller volumes are empty, then verifies API login and the exact
`server_version`. The Go proxy captures HTTP inform and UDP discovery traffic
and forwards both to the controller. The test overlay in `compose.tests.yaml`
adds three temporary services:

- `stub-bridge-observe` creates a container-local Linux bridge, two virtual
  member links, and dynamic FDB entries. The daemon reads the bridge and renders
  learned MACs onto virtual UniFi ports.
- `stub-port-map` creates two container-local veth sources and maps an 8-port
  switch profile to `interface`, `disabled`, and `unmapped` port states.
- `stub-gateway-smoke` maps the two-port `uxg-lite` gateway profile to the same
  veth sources and asserts the gateway-specific tables rendered from the shared
  port view.

Run the controller gate from the repository root:

```sh
make integration-docker
```

The script keeps the payload checks for all modes, including
`management_lan.mode: preexisting-interface`, and then verifies the complete
controller lifecycle for the switch and gateway fixtures:

- captured and controller-forwarded discovery;
- Pending Adoption and API-triggered adoption;
- a default-key `setparam` response containing authkey, cfgversion, and
  `use_aes_gcm=true`, recorded only as sanitized booleans;
- persisted `STATE=connected`, authkey, cfgversion, and AES-GCM preference;
- at least two post-adoption AES-GCM inform requests; and
- three consecutive connected controller samples;
- the pinned controller's explicit gateway-command expectation before and
  after a 75-second stability window (`absent` for the current 10.4.57 pin);
- the gateway's positive controller interval, clamped to local bounds, as
  effective and next delay, with zero consecutive failures and no active
  backoff.

Temporary stub containers, stub volumes, and adopted throwaway devices are
removed after the gate. Controller and MongoDB volumes remain reusable. A fresh
clone needs no manual controller setup; the default loopback-only lab account
is `admin` / `admin` and can be overridden with
`UNIFI_STUB_LAB_ADMIN_USER`/`UNIFI_STUB_LAB_ADMIN_PASSWORD`.
Before each standard run, adopted fixtures using the reserved `02:15:6d` lab
prefix are removed so a previous `KEEP_RESOURCES=1` debugging run cannot occupy
the controller's single gateway slot. If UniFi reports `api.err.DeviceBusy`
for one of those reserved fixtures, the gate restarts only the disposable
controller service and retries the normal API deletion; controller and MongoDB
volumes remain intact.

The 10.3.58 compatibility path can still require the complete safe command
response lifecycle explicitly:

```sh
UNIFI_NETWORK_IMAGE=lscr.io/linuxserver/unifi-network-application:10.3.58-ls129 \
UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION=10.3.58 \
UNIFI_STUB_DOCKER_GATEWAY_COMMAND_EXPECTATION=required \
make integration-docker
```

By default, the script derives throwaway test MACs and IPs inside the Compose
lab subnet for each run. Pin them only when debugging a specific controller
run:

```sh
UNIFI_STUB_BRIDGE_MAC=02:15:6d:00:08:21 \
UNIFI_STUB_BRIDGE_IP=172.31.242.25 \
make integration-docker
```

Set `UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` when you intentionally want to inspect
the temporary stub container or its state volume after a failing or exploratory
run. Without that override, the script stops and removes its temporary stub
resources before exiting.

Useful direct commands:

```sh
docker compose -f lab/stub/compose.yaml -f lab/stub/compose.tests.yaml config
docker compose -f lab/stub/compose.yaml -f lab/stub/compose.tests.yaml run --rm --no-deps stub-bridge-observe -dry-run
docker compose -f lab/stub/compose.yaml -f lab/stub/compose.tests.yaml run --rm --no-deps stub-port-map -dry-run
docker compose -f lab/stub/compose.yaml -f lab/stub/compose.tests.yaml run --rm --no-deps stub-gateway-smoke -dry-run
```

The Docker gate proves the Linux bridge/FDB, sysfs, port-map, payload,
discovery, inform, controller adoption, persisted state, AES-GCM, and the
configured safe gateway-command expectation, including a gateway-shaped
`uxg-lite` target. It does not prove Proxmox host
bridge behavior or FreeBSD runtime behavior; those remain separate host/VM
tests.
