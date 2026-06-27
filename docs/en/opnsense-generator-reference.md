# OPNsense API Generator Reference

This page documents the `unifi-stubd-opnsense` source file and generator
behavior. For a command-by-command setup on an OPNsense host, use the
[OPNsense API Generator How-to](opnsense-generator.md).

## Scope

`unifi-stubd-opnsense` is intentionally separate from the daemon:

- It reads an existing `unifi-stubd` YAML config.
- It reads one OPNsense source YAML file.
- It performs HTTP `GET` requests against OPNsense.
- It writes generated `unifi-stubd` YAML to stdout or to `-out`.
- It does not run as a live sync service.
- It does not mutate OPNsense interfaces, routes, firewall rules, VLANs, or the
  running `unifi-stubd` daemon.

## Command Reference

```sh
unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml
```

Flags:

| Flag | Required | Meaning |
| --- | --- | --- |
| `-config` | no | Base `unifi-stubd` YAML config. Defaults to the packaged config path. |
| `-source` | yes | OPNsense source YAML file. |
| `-out` | no | Write generated YAML to this path. Without it, YAML is printed to stdout. |
| `-validate` | no | Validate base config, source YAML, and credential loading without API calls or output. |

## OPNsense API Calls

The client builds URLs as `<base_url>/api/<path>` and sends Basic Auth with the
configured API key as username and API secret as password.

Implemented reads:

| Source setting | Endpoint |
| --- | --- |
| always | `GET /api/interfaces/overview/interfaces_info` |
| fallback for a missing mapped interface | `GET /api/interfaces/overview/get_interface/<interface>` |
| `gateway_status: true` | `GET /api/routes/gateway/status` |

Responses are limited to 8 MiB. Error messages include endpoint paths and HTTP
status codes, but not API key or secret values.

## Source YAML Fields

Top-level source file:

| Field | Required | Default | Description |
| --- | --- | --- | --- |
| `base_url` | yes | none | OPNsense WebGUI/API base URL, for example `https://127.0.0.1`. Must use `http` or `https` and include a host. |
| `api_key_file` | conditional | empty | File containing the raw API key. Used when `api_key_env` is unset or empty. |
| `api_secret_file` | conditional | empty | File containing the raw API secret. Used when `api_secret_env` is unset or empty. |
| `api_key_env` | conditional | empty | Environment variable containing the raw API key. Takes precedence over `api_key_file` when non-empty. |
| `api_secret_env` | conditional | empty | Environment variable containing the raw API secret. Takes precedence over `api_secret_file` when non-empty. |
| `ca_file` | no | empty | PEM CA bundle used to verify OPNsense TLS. |
| `insecure_skip_verify` | no | `false` | Allows self-signed lab endpoints without certificate validation. Use only intentionally. |
| `timeout_ms` | no | `2000` | Per-request timeout in milliseconds. Must be positive. |
| `uplink_port` | no | `0` | Generated `uplink_port` hint. Applied only when the base config has no `uplink_port`. |
| `gateway_status` | no | `false` | Enables gateway status API reads and WAN health hint mapping. |
| `interfaces` | yes | none | Port-to-interface mappings. Must contain at least one entry. |
| `wan_health` | no | empty | Optional generated `wan_health` block. Applied only when `wan_health.source` is non-empty. |

Credential rules:

- At least one key source and one secret source must be configured.
- Environment variables win over files when they are configured and non-empty.
- Empty credential values are rejected.
- Credential values are never rendered into the generated `unifi-stubd` config.

## Interface Mapping Fields

Each `interfaces[]` entry maps one represented UniFi port to one OPNsense
interface:

| Field | Required | Description |
| --- | --- | --- |
| `port` | yes | One-based represented UniFi port index. Must be positive and unique in the source file. |
| `interface` | yes | OPNsense/FreeBSD interface name such as `ixl0`, `igb0`, or `vtnet0`. Slashes are rejected. |
| `name` | no | Generated port label override. |
| `role` | no | Effective role such as `wan`, `wan2`, `lan`, `lan2`, or `unassigned`. Normalized to lowercase. |
| `network_group` | no | UniFi network group label, for example `WAN`, `WAN2`, or `LAN`. |
| `portconf_id` | no | Controller port-profile assignment ID to mirror. |
| `networkconf_id` | no | Controller network assignment ID to mirror. |
| `native_networkconf_id` | no | Controller native-network assignment ID to mirror. |
| `network_name` | no | Controller network display name to mirror. |
| `vlan` | no | Controller display VLAN ID to mirror. |
| `speed` | no | Represented link speed override in Mbps. |
| `media` | no | Represented media label such as `GE`, `SFP+`, or `SFP28`. |

The OPNsense interface name is source metadata. The UniFi `ifname` remains
profile-derived. For example, UXG-Pro port 3 remains `eth2` in controller-facing
payloads even when the source interface is `ixl0`.

## Merge Behavior

Generation starts from the loaded base config.

Port overrides:

- OPNsense-derived overrides are generated from `interfaces[]`.
- Existing base `port_overrides` are merged by `port`.
- Base config values win field by field.
- Generated-only ports are kept.
- Output order is sorted by port.

Top-level fields:

- `uplink_port` from the source is applied only when the base config has
  `uplink_port: 0` or no value.
- `wan_health` from the source is applied only when `wan_health.source` is
  non-empty.

Gateway status:

- Gateway status values are applied only to overrides whose generated or base
  role is `wan` or `wan2`.
- `wan_connected`, `wan_latency_ms`, and `wan_uptime_percent` can be generated.
- LAN and unassigned ports do not receive WAN health hints.

## Example Source File

See
[`lab/stub/configs/hosts/opnsense-api-source.example.yaml`](../../lab/stub/configs/hosts/opnsense-api-source.example.yaml)
for a complete sanitized source example.

## Troubleshooting

`opnsense base_url is required`:

- Set `base_url` in the source YAML.

`api_key requires an env var or file` or `api_secret requires an env var or file`:

- Configure `api_key_file` and `api_secret_file`, or set `api_key_env` and
  `api_secret_env`.

`returned HTTP 401`:

- Check that the API key and secret are correct and belong to an OPNsense user
  that can access the queried API pages.

`certificate signed by unknown authority`:

- Prefer setting `ca_file` to a PEM CA bundle.
- For isolated labs only, set `insecure_skip_verify: true`.

Generated config uses `eth2` instead of `ixl0` as `ifname`:

- This is expected. `eth2` is the UniFi profile interface. `ixl0` is rendered as
  source metadata in the generated port override.

Manual base config values are not overwritten:

- This is expected. Existing base `port_overrides` win over generated values
  field by field.
