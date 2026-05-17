# Lab Switch Identities

These files describe local UniFi controller lab devices. They are not installed
by packages automatically.

- `service-us8-minimal.yaml`: Small eight-port switch identity.
- `service-us16p150.yaml`: Sixteen-port PoE switch identity.
- `service-us16xg-10g.yaml`: Sixteen-port 10G switch identity.
- `service-usaggpro.yaml`: Controller-known Pro Aggregation identity with 10G and 25G port groups.
- `service-usw-pro-xg-48.yaml`: Pro XG 48 identity with 2.5G, 10G, and 25G port groups.
- `stub/compose.yaml`: Docker controller lab with only the generic `stub`
  switch stub container.
- `gateway-profiles/`: real firmware simulation container wrappers for gateway
  firmware profiles.
- `mongo-init-unifi.sh`: MongoDB user bootstrap used by the Docker controller lab.
- `mitm-inform-dump.py`: mitmproxy addon that records inform request/response metadata and raw local lab bodies.
- `observe-bridge.sh`: Create or remove a lab-only Linux bridge with veth members for observe-mode tests.
- `openrc/unifi-stubd-observe-bridge`: Optional OpenRC service for the observe bridge fixture.
- `local.d/unifi-stubd-observe-bridge.start`: Optional Alpine local.d boot hook for the observe bridge fixture.
- `us16p150-dry-output.sh`: Print the US-16-150W discovery and inform data.
- `us16xg-single-inform.sh`: Send one US-16-XG inform cycle.
- `minimal-switch-payload.json`: Minimal payload fixture for protocol work.

## Docker Stub Lab

`stub/compose.yaml` starts a private Docker lab with:

- a UniFi Network Application container,
- a MongoDB container,
- an inform MITM container, and
- the generic `stub` switch stub container.

The path, service, default container name, hostname, and persistent volume are
declared as `stub`:

```text
lab/stub/compose.yaml
services.stub
container_name: stub
hostname: stub
volume: stub_state
```

Start the generic stub service and its controller/MITM dependencies:

```sh
mkdir -p lab/stub/captures
docker compose -f lab/stub/compose.yaml up -d --build stub
```

Open the UniFi UI at `https://localhost:8443`. During setup, keep device
communication on TCP `8080` and set the Inform Host override to `unifi`.
The stub sends informs to `http://unifi:8080/inform`; inside the container,
`unifi` is mapped to the MITM container, which forwards to the real controller
service. Captures are written to ignored `lab/stub/captures/`.

The `stub` service builds the generic root `Dockerfile` and passes
`${UNIFI_STUB_PROFILE:-us8}` at runtime. The default emulated UniFi profile is
`us8`; the Docker path and container identity remain `stub`.

Stop and remove the disposable stub lab state:

```sh
docker compose -f lab/stub/compose.yaml down -v
```

## Docker Gateway Firmware Labs

`gateway-profiles/` contains the Docker files for real gateway firmware
simulation. These directories are not `internal/device` stub profile copies.

Current profiles:

- `gateway-profiles/ugw3/`: QEMU-MIPS runner for an extracted UGW3 rootfs.
- `gateway-profiles/uxg-lite/`: ARM64 UbiOS userspace wrapper; partial
  simulation.
- `gateway-profiles/uxgpro/`: ARM64 UbiOS userspace wrapper plus controller
  lab helpers.
- `gateway-profiles/ucg-fiber/`: ARM64 UbiOS userspace wrapper; partial
  simulation.

Run a firmware simulation through the profile's own Compose file:

```sh
docker compose -f lab/gateway-profiles/ugw3/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/uxg-lite/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/uxgpro/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/ucg-fiber/compose.yaml up -d --build
```

UXG-Pro also has a controller/MITM lab:

```sh
mkdir -p lab/gateway-profiles/uxgpro/captures
docker compose -f lab/gateway-profiles/uxgpro/controller-lab.compose.yaml up -d --build
```

Each firmware profile keeps its own `Dockerfile`, `compose.yaml`,
`docker-howto.md`, startup wrapper, and safe firmware notes in the same
directory. Firmware images, extracted rootfs trees, raw captures, adoption
keys, controller tokens, certificates, and private controller data stay out of
Git.

## Observe Bridge Fixture

The observe bridge fixture is intentionally isolated. It creates `stubbr0` with
two veth bridge members, but it does not attach the management interface.

Manual install on an Alpine/OpenRC lab VM:

```sh
sudo install -m 0755 lab/observe-bridge.sh /usr/local/sbin/unifi-stubd-observe-bridge
sudo install -m 0755 lab/openrc/unifi-stubd-observe-bridge /etc/init.d/unifi-stubd-observe-bridge
sudo install -m 0644 lab/openrc/unifi-stubd-observe-bridge.confd /etc/conf.d/unifi-stubd-observe-bridge
sudo rc-update add unifi-stubd-observe-bridge default
sudo rc-service unifi-stubd-observe-bridge start
```

Alternative `local.d` install:

```sh
sudo install -m 0755 lab/observe-bridge.sh /usr/local/sbin/unifi-stubd-observe-bridge
sudo install -m 0755 lab/local.d/unifi-stubd-observe-bridge.start /etc/local.d/unifi-stubd-observe-bridge.start
sudo install -m 0755 lab/local.d/unifi-stubd-observe-bridge.stop /etc/local.d/unifi-stubd-observe-bridge.stop
sudo rc-update add local default
```
