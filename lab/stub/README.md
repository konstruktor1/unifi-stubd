# Generic Stub Compose Lab

This is the main Docker lab for the Go `unifi-stubd` daemon. It starts a UniFi
Network Application, MongoDB, a Go inform proxy, and one `stub` container built
from the repository root. Use it when validating discovery, inform, adoption
state, or profile payload behavior in the Go service.

The default UniFi Network Application image is pinned to
`lscr.io/linuxserver/unifi-network-application:10.3.58-ls129`. Override
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

## Docker Integration Tests

The Docker integration smoke keeps external controller behavior out of the
default gate. It uses the Go inform proxy only as a local capture point and
does not require the UniFi Network Application API to be ready. The test
overlay in `compose.tests.yaml` adds three temporary services:

- `stub-bridge-observe` creates a container-local Linux bridge, two virtual
  member links, and dynamic FDB entries. The daemon reads the bridge and renders
  learned MACs onto virtual UniFi ports.
- `stub-port-map` creates two container-local veth sources and maps an 8-port
  switch profile to `interface`, `disabled`, and `unmapped` port states.
- `stub-gateway-smoke` maps the two-port `uxg-lite` gateway profile to the same
  veth sources and asserts the gateway-specific tables rendered from the shared
  port view.

Run the integration smoke test from the repository root:

```sh
make integration-docker
```

The script validates the Compose overlay, builds the lab image, asserts dry-run
payload JSON for all modes, including `management_lan.mode:
preexisting-interface` on the bridge-observe container, then sends one inform
request per mode through the Go inform proxy. The proxy target defaults to
`http://127.0.0.1:9` in this smoke test so the request is captured without
depending on a live controller response, controller version, adoption API, or
UniFi-internal timing. Override `UNIFI_STUB_INFORM_PROXY_TARGET` only when
intentionally debugging forwarding to a controller.

Temporary stub containers and stub volumes are removed again after the smoke
test. Controller volumes and controller state are not touched by the default
integration smoke.

By default, the script derives throwaway test MACs and IPs inside the Compose
lab subnet for each run. Pin them only when debugging a specific Docker smoke
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

The Docker tests prove the Linux bridge/FDB, sysfs, port-map, payload, and Go
inform proxy paths inside containers, including a gateway-shaped `uxg-lite`
smoke path. They do not prove Proxmox host bridge behavior or FreeBSD runtime
behavior; those remain separate host/VM tests.
