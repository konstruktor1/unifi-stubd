# Lab Switch Identities

These files describe local UniFi controller lab devices. They are not installed
by packages automatically.

- `service-us8-minimal.yaml`: Small eight-port switch identity.
- `service-us16p150.yaml`: Sixteen-port PoE switch identity.
- `service-us16xg-10g.yaml`: Sixteen-port 10G switch identity.
- `service-usaggpro.yaml`: Controller-known Pro Aggregation identity with 10G and 25G port groups.
- `service-usw-pro-xg-48.yaml`: Pro XG 48 identity with 2.5G, 10G, and 25G port groups.
- `controller-gateway-stubs.compose.yaml`: Docker controller lab with
  profile-selectable `us8` switch stub plus `ugw3`, `uxg-lite`, `uxgpro`, and
  `ucg-fiber` gateway stubs.
- `mongo-init-unifi.sh`: MongoDB user bootstrap used by the Docker controller lab.
- `mitm-inform-dump.py`: mitmproxy addon that records inform request/response metadata and raw local lab bodies.
- `observe-bridge.sh`: Create or remove a lab-only Linux bridge with veth members for observe-mode tests.
- `openrc/unifi-stubd-observe-bridge`: Optional OpenRC service for the observe bridge fixture.
- `local.d/unifi-stubd-observe-bridge.start`: Optional Alpine local.d boot hook for the observe bridge fixture.
- `us16p150-dry-output.sh`: Print the US-16-150W discovery and inform data.
- `us16xg-single-inform.sh`: Send one US-16-XG inform cycle.
- `minimal-switch-payload.json`: Minimal payload fixture for protocol work.

## Docker Gateway Stub Lab

`controller-gateway-stubs.compose.yaml` starts a private Docker lab with:

- a UniFi Network Application container,
- a MongoDB container,
- an inform MITM container, and
- one selected `unifi-stubd` stub container.

Start the generic US-8 switch stub:

```sh
mkdir -p lab/captures
docker compose -f lab/controller-gateway-stubs.compose.yaml \
  --profile stub \
  up -d --build
```

Start one gateway profile:

```sh
mkdir -p lab/captures
docker compose -f lab/controller-gateway-stubs.compose.yaml \
  --profile uxg-lite \
  up -d --build
```

Other profiles:

```sh
docker compose -f lab/controller-gateway-stubs.compose.yaml --profile us8 up -d --build
docker compose -f lab/controller-gateway-stubs.compose.yaml --profile ugw3 up -d --build
docker compose -f lab/controller-gateway-stubs.compose.yaml --profile uxgpro up -d --build
docker compose -f lab/controller-gateway-stubs.compose.yaml --profile ucg-fiber up -d --build
```

Open the UniFi UI at:

```text
https://localhost:8443
```

During setup, keep device communication on TCP `8080` and set the Inform Host
override to:

```text
unifi
```

The gateway stub sends informs to `http://unifi:8080/inform`. Inside the stub
container, `unifi` is mapped to the MITM container, which forwards to the real
controller service. Captures are written to `lab/captures/`, which is ignored
by Git because adopted inform traffic can contain controller state and keys.

The `stub-us8` service builds the generic root `Dockerfile` and passes
`-profile us8` at runtime. Gateway services build the profile-specific
Dockerfiles under `lab/gateway-profiles/`.

Use one gateway profile per clean controller site when testing gateway
adoption. The Compose profile `gateways` can start all gateway profiles for
packet-shape comparison, but a normal UniFi site should not be expected to
adopt multiple gateway devices at once.

Stop and remove the disposable controller state:

```sh
docker compose -f lab/controller-gateway-stubs.compose.yaml down -v
```

The real firmware simulation catalog is tracked separately in
`research/firmware/profiles.yaml`. A real firmware profile needs a local
vendor firmware image and extracted rootfs; the stub profiles above do not
execute vendor firmware.

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
