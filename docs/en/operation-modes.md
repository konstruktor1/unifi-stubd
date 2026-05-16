# Operation Modes

`unifi-stubd` currently targets Linux lab hosts first, including Proxmox,
Alpine, and UTM Linux VMs.

## Current Validated State

The validated live lab device is:

- Host: `10.0.0.151`
- Profile: `usaggpro`
- Controller model: `USAGGPRO` / `USW Pro Aggregation`
- MAC: `32:c1:80:4f:7e:bc`
- Controller state: online and adopted
- Ports: 28 10G SFP+ ports and four 25G SFP28 ports, with port 29 as uplink

`USAGGPRO` is the currently validated large 10G profile. `USWProXG48` remains
experimental because the current lab controller did not accept it as a known
pending adoption model.

## Modes

### `stub`

Default mode. The daemon sends discovery and inform payloads from profile data
only. It does not read host bridge state and does not change host networking.

### `observe`

Read-only Linux observation mode. It keeps the same fake device identity, but
uses passive host data when available:

- `/sys/class/net/<interface>/statistics/*` for port counters
- `/sys/class/net/<interface>/speed` for uplink speed
- `bridge fdb show br <bridge>` for learned MAC table entries

FDB rows are grouped by Linux bridge member. The configured
`observe_interface` is mapped to the UniFi uplink port, while `tap*`, `veth*`,
and other learned bridge members are mapped deterministically to free switch
ports with their learned MAC tables.

If any source is missing or unreadable, the daemon logs a warning and falls back
to profile defaults. This mode must not create interfaces, assign addresses, or
change routes.

### `host-direct`

Direct host identity mode. It does not create a separate MAC or IP. The special
`mac: host` value is only allowed in this mode and requires
`observe_interface` so the daemon can read the host interface MAC explicitly.

### `macvlan`

Planning-only mode in this release. It is Linux-only and must be combined with
`-dry-run-plan`. The daemon prints the planned macvlan commands, but does not
execute them.

## Passive Sources

LLDP is currently accepted as `lldp_source: off` or `lldp_source: lldpd`, but
only `off` has runtime behavior today. Traffic metadata is currently
`traffic_source: off` only; packet capture and DPI are intentionally out of
scope for the first observation wave.
