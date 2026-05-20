# Operator Guide

This page is the runbook-level view for operating `unifi-stubd` safely. It is
not a replacement for the reference documentation.

## Before Running

Check these items before sending discovery or inform traffic:

- use an isolated lab or management network;
- use a disposable fake MAC unless testing physical-MAC controller behavior;
- use documentation/example addresses in committed files;
- keep controller tokens, adoption keys, SSH passwords, real MAC tables, and
  captures out of Git;
- run `-validate` before daemon mode;
- run `-dry-run` before talking to a controller;
- decide whether the device should be synthetic, bridge-observed, or port-mapped.

## Recommended First Command

```sh
go run ./cmd/unifi-stubd -validate -config packaging/linux/etc/unifi-stubd/config.yaml
```

Then inspect the payload without network side effects:

```sh
go run ./cmd/unifi-stubd -dry-run -no-discovery
```

## Operation-Mode Selection

| Mode | Use when | Main risk |
| --- | --- | --- |
| `stub` | You need a synthetic UniFi device from profile data | payload may not match a real host |
| `bridge-observe` | A Proxmox/Linux bridge should look like a switch | topology direction can be controller-derived |
| `port-map` | Each UniFi port maps to a known host interface | every port needs an explicit source |
| `host-direct` | The host identity itself is the represented device | only use when that identity model is intentional |
| `macvlan` | You are planning a future active host-network mode | dry-run-plan only in this release |

Detailed behavior is documented in [Operation Modes](../../en/operation-modes.md).

## Bridge-Observe Runbook

Use `bridge-observe` when one host bridge should represent a switch. The bridge
itself is not a UniFi port; it is the observation boundary.

Minimum shape:

```yaml
operation_mode: bridge-observe
profile: us48p500
mac: auto
bridge_observe:
  bridge: vmbr0
  uplink_interface: eno1
uplink_port: 49
uplink_neighbor:
  mac: 02:00:5e:00:53:01
  vlan: 1
  type: usw
```

Rules:

- map VM/container members as access ports;
- map the physical upstream interface as the uplink;
- set `uplink_port` explicitly when the real link is SFP/SFP+;
- prefer a synthetic locally administered stub MAC for representation tests;
- use a physical host MAC only when testing how the controller handles that
  real MAC on an upstream switch.

## Port-Map Runbook

Use `port-map` when each represented port has a deliberate source:

```yaml
operation_mode: port-map
port_mappings:
  - port: 1
    interface: eno1
  - port: 2
    disabled: true
  - port: 3
    unmapped: true
```

Rules:

- every profile port must have exactly one mapping;
- `interface` sources must exist at validation/runtime;
- `disabled` renders link down and speed `0`;
- `unmapped` keeps profile defaults without a host sensor;
- no host interface is configured or changed by the daemon.

## Management LAN

Current switch management LAN support is intentionally conservative:

- `metadata-only`: report the VLAN in payload/status only;
- `preexisting-interface`: bind management identity to an interface that already
  exists, for example `vmbr0.20`;
- `planned-host-vlan`: dry-run-plan only.

The daemon does not create VLAN interfaces in the current release.

## Adoption And Cleanup

Use disposable MACs for controller tests. A clean adoption cycle is:

1. Run `-dry-run` with final MAC/IP/profile values.
2. Run one inform against the controller.
3. Adopt only the disposable device.
4. Verify local `STATE=connected` through status.
5. Use controller forget/remove for the disposable device.
6. Stop the stub.
7. Delete the local temporary state directory if the test is complete.

Do not reuse the same MAC/IP with different profiles unless the previous device
has been forgotten and local adoption state has been reset.

## Status Checks

```sh
unifi-stubd -status
unifi-stubd -status-json
```

Status should explain:

- profile and operation mode;
- effective MAC/IP/hostname;
- adoption state without printing authkey;
- platform capability state;
- observation warnings;
- last inform result.

## Common Failure Modes

Device stays pending/adopting:
stale controller device or local adoption state. Forget the controller device
and reset local state.

Topology edge points the wrong way:
physical host MAC is also visible on an upstream UniFi switch. Use a synthetic
locally administered stub MAC.

Uplink appears on wrong port:
mixed-speed profile without `uplink_port`. Set the profile port explicitly.

Too many clients appear directly attached:
uplink MACs are not filtered or the uplink is misclassified. Set
`bridge_observe.uplink_interface` explicitly.

`port-map` validation fails:
mapping or interface is missing. Provide one valid entry per profile port.
