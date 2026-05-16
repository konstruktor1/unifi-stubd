# Lab Switch Identities

These files describe local UniFi controller lab devices. They are not installed
by packages automatically.

- `service-us8-minimal.yaml`: Small eight-port switch identity.
- `service-us16p150.yaml`: Sixteen-port PoE switch identity.
- `service-us16xg-10g.yaml`: Sixteen-port 10G switch identity.
- `service-usaggpro.yaml`: Controller-known Pro Aggregation identity with 10G and 25G port groups.
- `service-usw-pro-xg-48.yaml`: Pro XG 48 identity with 2.5G, 10G, and 25G port groups.
- `observe-bridge.sh`: Create or remove a lab-only Linux bridge with veth members for observe-mode tests.
- `openrc/unifi-stubd-observe-bridge`: Optional OpenRC service for the observe bridge fixture.
- `local.d/unifi-stubd-observe-bridge.start`: Optional Alpine local.d boot hook for the observe bridge fixture.
- `us16p150-dry-output.sh`: Print the US-16-150W discovery and inform data.
- `us16xg-single-inform.sh`: Send one US-16-XG inform cycle.
- `minimal-switch-payload.json`: Minimal payload fixture for protocol work.

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
