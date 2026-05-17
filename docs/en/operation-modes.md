# Operation Modes

`unifi-stubd` currently targets Linux lab hosts first, including Proxmox,
Alpine, and UTM Linux VMs. FreeBSD/OPNsense is supported as a stub-only target;
native observation is not implemented there yet.

## Current Validated State

The validated live lab device is:

- Host: `192.0.2.151`
- Profile: `usaggpro`
- Controller model: `USAGGPRO` / `USW Pro Aggregation`
- MAC: `02:00:5e:00:53:51`
- Controller state: online and adopted
- Ports: 28 10G SFP+ ports and four 25G SFP28 ports
- Uplink: port 1 by live `uplink_port` override; profile default is port 29

`USAGGPRO` is the currently validated large 10G profile. `USWProXG48` remains
experimental because the current lab controller did not accept it as a known
pending adoption model.

`UGW3` is available as an experimental gateway identity profile. It reports the
legacy UniFi Security Gateway model and three 1G ports, but remains stub-only
and does not emulate router services yet.

`UXGPRO` is available as an experimental 10G gateway identity profile. It
keeps the original gateway-style assignment: `WAN` on the 1G RJ45 primary WAN,
`LAN` on the 1G RJ45 LAN, `WAN2` on the 10G SFP+ secondary WAN, and `LAN2` on
the 10G SFP+ LAN. Use `uplink_port` and `port_overrides` when a lab maps the
active internet side to the SFP+ port.

## Modes

### `stub`

Default mode. The daemon sends discovery and inform payloads from profile data
only. It does not read host bridge state and does not change host networking.
This is the supported FreeBSD/OPNsense mode.

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

The profile chooses the uplink port by default. Set `uplink_port` to a positive
port number to move the uplink marker to a specific physical port while keeping
that port's profile speed and media. For example, `uplink_port: 1` puts the
`usaggpro` uplink on a 10G SFP+ port instead of the default 25G SFP28 uplink
group.

Profiles describe the real hardware layout: model, port count, speed/media
groups, default port names, and default gateway roles. Use `port_overrides`
for lab-specific assignments after profile and observation data have been
applied:

```yaml
uplink_neighbor:
  mac: 02:aa:bb:cc:dd:01
  vlan: 1
  type: usw

port_neighbors:
  - port: 2
    mac: 02:00:5e:00:53:03
    vlan: 1
    type: usw

port_overrides:
  - port: 2
    name: lab_lan
    role: lan
    network_group: LAN
    interface: eth1
    ip: 192.0.2.51
    netmask: 255.255.255.0
    speed: 1000
  - port: 3
    name: backup_wan
    role: wan2
    network_group: WAN2
    speed: 2500
  - port: 4
    speed: 100
  - port: 5
    up: false
```

`port_neighbors` populates `port_table[].mac_table` on specific ports. It is
useful when the controller needs to see a downstream switch or host MAC on a
non-uplink port.

Gateway models report WAN/LAN assignments through `config_port_table`,
`ethernet_overrides`, `network_table`, and `reported_networks`. Switch-style
MAC-table neighbors may be ignored by the controller for gateway identities, so
use `role` and `network_group` for gateway visualization instead of changing
the hardware profile.

For `UXGPRO`, the controller renders gateway ports from its gateway model and
reported WAN/LAN state. Do not expect it to expose the same switch `port_table`
view as a UniFi switch profile.

`port_overrides[].interface` is read-only. It lets the daemon copy an existing
host interface MAC, IPv4 address, link state, and available counter/speed data
into that port's inform payload. This is useful for FreeBSD/OPNsense stub-only
gateway tests where WAN/LAN should be visualized from existing interfaces
without changing host networking.

`uplink_neighbor` is useful for pure stubs and virtual lab ports where there is
no physical link partner. It adds a configured MAC-table entry to the current
uplink port.

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
