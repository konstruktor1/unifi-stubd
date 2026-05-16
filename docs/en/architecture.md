# Architecture

## Components

```mermaid
flowchart LR
  A["Linux bridge / host"] --> B["adapters"]
  B --> C["device payload builder"]
  C --> D["inform client"]
  E["discovery announcer"] --> F["UniFi Network Controller"]
  D --> F
  F --> G["adoption/provisioning responses"]
  G --> H["adoption state store"]
  H --> D
  I["/etc/unifi-stubd/config.yaml"] --> J["cmd/unifi-stubd"]
  J --> C
  J --> E
```

## Packages

| Package | Responsibility |
| --- | --- |
| `cmd/unifi-stubd` | CLI/daemon entrypoint |
| `internal/discovery` | Build and send UDP discovery TLVs |
| `internal/inform` | Encode/decode `TNBU` packets |
| `internal/adoption` | Authkey, cfgversion, and lifecycle state |
| `internal/device` | Build UniFi status payloads |
| `internal/adapters/linuxbridge` | Translate Linux bridge FDB data into MAC tables |
| `internal/observe` | Read-only Linux sysfs/FDB snapshot and payload merge |
| `internal/config` | Load runtime configuration |

## Runtime Layout

| Path | Content |
| --- | --- |
| `/usr/local/bin/unifi-stubd` | Program binary |
| `/etc/unifi-stubd/config.yaml` | Runtime config for controller, profile, MAC/IP, SSH adoption, and port speed |
| `/etc/unifi-stubd/ssh_host_rsa_key` | SSH host key for fake adoption |
| `/var/lib/unifi-stubd/adoption.env` | Persistent controller state, authkey, cfgversion, and inform URL |

## State Machine

```mermaid
stateDiagram-v2
  [*] --> Factory
  Factory --> Discovered: UDP discovery / default inform
  Discovered --> Adopting: controller adopt
  Adopting --> Provisioning: setparam + authkey saved
  Provisioning --> Connected: heartbeat accepted
  Connected --> Connected: noop / status inform
  Connected --> Provisioning: cfgversion changes
  Adopting --> Failed: decrypt / invalid inform / timeout
  Provisioning --> Failed: controller rejected payload
```

## Design Decisions

### Fake Switch Before Fake Gateway

A switch profile mainly needs ports, interface state, MAC tables, and counters. A gateway profile needs WAN/LAN state, routing, DHCP, DPI, firewall, health, and more controller-specific fields. The switch MVP is therefore much more robust.

### No Real Provisioning

Controller commands are interpreted and persisted first, but not applied to the host. Anything that would mutate the host belongs in logs or debug output until it is explicitly implemented.

### Pin the Lab Version

UniFi Network changes implicit payload expectations. Development should pin one controller version in a VM first, then add more versions to a compatibility matrix.

## Data Sources for Proxmox

| Source | UniFi target |
| --- | --- |
| `bridge fdb show` | `port_table[].mac_table`, grouped by bridge member |
| `/sys/class/net/<if>/statistics/*` | rx/tx bytes, packets, errors |
| `ip -json addr` | `if_table` |
| `lldpcli -f json show neighbors` | later neighbor hints |
| Proxmox API | map VM names to MACs |
