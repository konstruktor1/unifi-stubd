# Stub Lab Configurations

This directory contains the project-owned runtime configurations for the Docker
stub lab. They are safe, synthetic defaults for the private Compose network and
do not contain controller secrets, adoption keys, captures, or real site data.

- `hosts/stub/config.yaml`: long-lived generic `stub` service.
- `hosts/stub-bridge-observe/config.yaml`: temporary bridge-observe integration
  host.
- `hosts/stub-port-map/config.yaml`: temporary switch port-map integration
  host.
- `hosts/stub-gateway-smoke/config.yaml`: temporary `uxg-lite` gateway-shaped
  smoke host.
- `hosts/server-lan1-sfp-lab/config.example.yaml`: sanitized SFP+ bridge
  observer with OPNsense on port 49 and aggregation on port 50.
- `hosts/opnsense-uxg-sfp-lab/config.example.yaml`: sanitized UXG Pro gateway
  with WAN on SFP+ port 3 and LAN on SFP+ port 4.

The Compose services mount this directory read-only at
`/usr/local/share/unifi-stubd-lab/configs`. Start scripts still pass MAC, IP,
profile, hostname, and interval flags so tests can derive fresh throwaway
identities without editing these files. Host-specific directories are named
after the reported stub hostname so local lab hosts can keep stable labels and
additional per-host files next to `config.yaml`.

For additional local lab hosts, add `hosts/<reported-hostname>/config.yaml` and
keep `hostname:` inside the YAML aligned with the directory name. Use
`UNIFI_STUB_CONFIG`, `UNIFI_STUB_BRIDGE_CONFIG`, `UNIFI_STUB_PORTMAP_CONFIG`,
or `UNIFI_STUB_GATEWAY_CONFIG` when a service should load a non-default host
configuration.

The SFP+ examples are intentionally shareable references, not direct copies of a
private controller. They keep the topology shape from the live lab while using
documentation IP ranges and locally administered example MAC addresses.

Real-network host snapshots belong under
`hosts/<real-hostname>/real/config.yaml` or
`hosts/<real-hostname>/temp/config.yaml`. Those paths are intentionally ignored
by Git because they can contain private controller URLs, real addresses, MACs,
interface names, or adoption state paths.
