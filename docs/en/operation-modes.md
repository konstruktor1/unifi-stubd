# Operation Modes

`unifi-stubd` currently targets Linux lab hosts first, including Proxmox,
Alpine, and UTM Linux VMs. FreeBSD/OPNsense is supported conservatively through
stub mode, explicit `port-map`, bridge FDB parsing, and read-only syslog-style
metadata.

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

`UXG` is available through the experimental `uxg-lite` gateway identity
profile. It reports a two-port Gateway Lite layout with `LAN` and `WAN`, but
remains stub-only and does not emulate router services yet.

`UXGPRO` is available as an experimental 10G gateway identity profile. It
keeps the original gateway-style assignment: `WAN` on the 1G RJ45 primary WAN,
`LAN` on the 1G RJ45 LAN, `WAN2` on the 10G SFP+ secondary WAN, and `LAN2` on
the 10G SFP+ LAN. Use `uplink_port` and `port_overrides` when a lab maps the
active internet side to the SFP+ port.

`UCGF` is available through the experimental `ucg-fiber` Cloud Gateway Fiber
identity profile. It reports `udm` device type with four 2.5G RJ45 LAN ports,
one 10G RJ45 `WAN2` port, one 10G SFP+ `WAN` port, and one 10G SFP+ LAN port.
It is stub-only and does not run UniFi OS or bundled controller applications.

## YAML Configuration Note

Runtime config is one operator-owned YAML document. Block-style YAML is the
recommended form. JSON-compatible YAML is accepted only when the whole config is
one complete JSON object; do not append block YAML such as `wan_health:` after a
closing `}`. Validate after every edit:

```sh
unifi-stubd -validate -config /etc/unifi-stubd/config.yaml
unifi-stubd -status-json
```

On FreeBSD/OPNsense, use `/usr/local/etc/unifi-stubd/config.yaml` for the same
commands. Controller `setparam`/`system_cfg` data is not a runtime config
source and must not be treated as authority for host networking.

## Modes

### `stub`

Default mode. The daemon sends discovery and inform payloads from profile data
only. It does not read host bridge state and does not change host networking.
This is the supported FreeBSD/OPNsense mode.

### `bridge-observe`

Read-only Linux observation mode. It keeps the same fake device identity, but
uses passive host data when available:

- `/sys/class/net/<interface>/statistics/*` for port counters
- `/sys/class/net/<interface>/speed` for uplink speed
- per-member interface state, speed, and counters for mapped bridge ports
- optional `/proc/net/dev` counters when `proc_source: procfs` is enabled
- `bridge fdb show br <bridge>` for learned MAC table entries
- optional passive LLDP neighbors when `lldp_source: lldpd` is enabled

FDB rows are grouped by bridge member and classified before port mapping.
`bridge_observe.uplink_interface` is mapped to the current UniFi upstream link. The
bridge device itself, for example `vmbr0` or `bridge0`, is treated as backplane
metadata and does not consume a UniFi port. Proxmox/VM members such as `tap*`,
`veth*`, `fwpr*`, `fwln*`, `fwbr*`, and FreeBSD-style `epair*`/`vnet*` members
are access/downstream ports with their learned MAC tables. If no uplink
interface is configured and exactly one physical-looking bridge member exists,
it is treated as the uplink candidate; otherwise unknown members are mapped as
normal ports. `bridge_observe.member_port_map` can pin a member to a specific
UniFi port when deterministic sorting is not enough.
`bridge_observe.ignored_members` excludes local bridge members entirely; use it
for TAP/epair sides that are already represented by another explicit physical
or uplink port.

`observe` remains accepted as a migration alias and is normalized internally to
`bridge-observe`. Existing `observe_interface` and `observe_bridge` configs are
still read as fallback values for `bridge_observe.uplink_interface` and
`bridge_observe.bridge`.

```yaml
operation_mode: bridge-observe
bridge_observe:
  bridge: vmbr0
  uplink_interface: eno1
  ignored_members:
    - tap10000i0
  member_port_map:
    - member: tap101i0
      port: 2
```

On a Proxmox host this lets `vmbr0` act as the represented switch. The uplink
member, for example `eno1`, carries the uplink counters and link speed. MAC
entries learned on that uplink are treated as remote devices behind the real
neighbor switch. The daemon tracks those remote MACs first and excludes them
from all local access-port MAC tables, even if the bridge reports a duplicate
FDB row elsewhere. VM or container participants such as `tap101i0` and
`veth200i0` are learned from the bridge FDB; their MAC addresses are placed
into `port_table[].mac_table`, while their own interface speed and counters are
used for the mapped access ports. Ports without a mapped bridge member are
reported as disconnected instead of inheriting synthetic profile link state.
The daemon re-reads sysfs and FDB data on each heartbeat; it does not subscribe
to netlink events in this wave.

The profile chooses the uplink port by default. Set `uplink_port` to a positive
port number to move the uplink marker to a specific physical port while keeping
that port's profile speed and media. For example, `uplink_port: 1` puts the
`usaggpro` uplink on a 10G SFP+ port instead of the default 25G SFP28 uplink
group.

For switch profiles with dedicated SFP/SFP+ uplink groups, those ports remain
the profile-defined uplink-capable cages. Separately, `bridge-observe` defaults
an explicitly configured physical `uplink_interface` to the last normal GE port
when `uplink_port` is unset. This makes the real host link appear as the active
copper upstream connection without redefining the SFP/SFP+ cages. Simple switch
profiles without a dedicated uplink group keep their profile default.

If the represented host is cabled through an SFP/SFP+ link, set `uplink_port`
explicitly to the matching profile port instead of relying on the GE fallback.
For example, a 48-port switch profile with SFP+ cages can report the bridge
uplink on `uplink_port: 49`, leaving port 48 disconnected and preserving the
SFP+ media label on the active uplink.

Topology direction depends on the controller's real view of the reported device
MAC. If the stub uses the physical bridge or NIC MAC, an upstream UniFi switch
may already report that MAC on one of its own ports. The controller can then
prefer the real switch's observation and render the link in the wrong direction,
even when `uplink_neighbor` is configured. A synthetic locally administered
stub MAC avoids that collision and is usually better for pure representation
tests. Use the physical MAC only when the goal is to intentionally test that
controller heuristic.

On Proxmox, the bridge interface itself can also have the host management IP.
That is normal for Proxmox but not identical to a hardware UniFi switch: the
stub's management identity then represents the host bridge IP, not an isolated
switch management interface. A dedicated management VLAN, macvlan/ipvlan, or a
separate test IP gives a cleaner model when the test needs to resemble a
physical switch more closely.

Profiles describe the real hardware layout: model, port count, speed/media
groups, default port names, and default gateway roles. Use `port_overrides`
for lab-specific assignments after profile and observation data have been
applied:

External profiles can extend built-in defaults with `profile_file` or
`profile_dir`. Treat them as lab data until validated against a concrete UniFi
Network version. Use `-profile-template`, `-profile-validate`, `-profile-export`,
and `-validate` to create and check YAML profiles without sending discovery or
inform traffic.

```yaml
uplink_neighbor:
  mac: 02:aa:bb:cc:dd:01
  vlan: 1
  type: usw

port_neighbors:
  - port: 2
    mac: 02:00:5e:00:53:03
    hostname: lab-host-2
    ip: 192.0.2.52
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
    wan_uptime_percent: 100
    wan_latency_ms: 7
    wan_connected: true
    speed: 2500
  - port: 4
    speed: 100
  - port: 5
    up: false
```

`port_neighbors` populates `port_table[].mac_table` on specific ports. It is
useful when the controller needs to see a downstream switch or host MAC on a
non-uplink port. `hostname` and `ip` are optional client metadata fields; `name`
is accepted as a YAML alias for `hostname`. When `type` is omitted, port
neighbors default to `client`; `uplink_neighbor` defaults to `usw`.

In Linux `bridge-observe` mode, learned bridge FDB MACs are also matched against
the local `/proc/net/arp` cache when available. This can add client IPv4
addresses to the MAC table without changing host networking. Hostnames are not
guessed from DNS; set `hostname` or `name` explicitly for deterministic labels.

Gateway models report observed WAN/LAN link facts through `if_table`,
`network_table`, a read-only physical `port_table`, `config_port_table`,
`ethernet_overrides`, `reported_networks`, `uplink`, `uplink_table`, and
`wan1`. WAN-like ports also report `uptime_stats` rows. Explicit assignment
IDs, network names, and VLAN metadata from
`port_overrides` are mirrored in the gateway port tables when configured.
Client neighbors are reported through `network_table[].host_table` with
`hostname` and `ip` when configured; upstream switch neighbors are not rendered
as gateway hosts.

Gateway interface names are profile data, not host interface names. The selected
profile's `gateway_interface_prefix` and one-based physical port index produce
the controller-facing names (`eth0`, `eth1`, ...). `port_overrides[].interface`
is only the local source used to read MAC, IP, link, speed, and counters; that
host name is rendered as `source_interface`. For example, a UXG-Pro lab where
OPNsense `ixl0` is cabled to physical port 3 should report:

```yaml
profile: uxgpro
uplink_port: 3
port_overrides:
  - port: 3
    role: wan
    network_group: WAN
    interface: ixl0
    speed: 10000
    media: SFP+
```

The resulting gateway rows use `ifname: eth2` because physical profile port 3
is `eth2`; `source_interface: ixl0` records where the local facts came from.
The role and network group describe the port's function. They do not rename the
physical profile interface.

Gateway management and gateway data-plane state are intentionally separate. The
stub can use a management or transport network to reach the controller through
top-level identity/runtime fields such as `ip`, `controller_url`,
`discovery_interface`, and `discovery_targets`. That does not make the
management address a LAN or WAN port address. Gateway WAN/LAN ownership comes
from profile ports plus `port_overrides`; a controller-management address
should stay out of `wan1`, `config_network_wan`,
`config_network_lan`, `network_table`, and `port_table` unless it is explicitly
configured as the address of that gateway port.

Unused physical profile ports are also kept separate from routed LAN state.
Ports marked `role: unassigned`, disabled ports, or disconnected ports without
an explicit gateway role are physical inventory only and do not inherit the LAN
IP. This avoids extra LAN/Gateway hints in controller web and mobile views.

For gateway lab displays, `port_overrides[].wan_uptime_percent`,
`wan_latency_ms`, `wan_downtime_seconds`, and `wan_connected` are deterministic
status hints only. They can make the controller see a configured WAN/WAN2 as
online, degraded, or down without creating routes, changing interfaces, or
accepting provisioning from the controller.

`wan_health` can optionally replace those static WAN hints with active,
read-only ping samples after profile ports, observation data, and
`port_overrides` have been merged. This works for gateway profiles in `stub`,
`bridge-observe`, and `port-map`; switch profiles do not render WAN health
payload fields. The default source is `off`, so no network probe runs unless the
operator opts in:

```yaml
wan_health:
  source: ping
  interval_seconds: 10
  timeout_ms: 1000
  targets:
    - port: 3
      host: 192.0.2.1
```

Ping results only update WAN telemetry fields such as connected state, latency,
downtime, and uptime percentage. The daemon still does not change host
interfaces, routes, VLANs, firewall rules, or controller provisioning state.
When ICMP is blocked or the local `ping` binary is unavailable, `-status` and
`-status-json` expose the last probe error instead of attempting a repair.
The probe uses the host's ordinary routing table. `targets[].port` selects the
gateway port whose telemetry is updated; it does not bind the ICMP packet to
`port_overrides[].interface`.

`wan_health.source` values:

- `off`: keep WAN health inactive. The payload uses link state and any explicit
  `wan_*` hints already present on the port.
- `static`: document that only static `port_overrides[].wan_*` values should be
  used. No command is executed.
- `ping`: run local read-only pings for `targets[]` and overlay only WAN health
  fields on the resolved port. Target ports must be effective `wan` or `wan2`
  ports after `port_overrides` have been applied.

Provider and ISP names are not inferred. The gateway payload can report
`speedtest-status.latency` and success/failure-like status values from WAN
health, but it does not run a UniFi speed-test service and does not populate
`speedtest-status.server.provider`, `isp_name`, or `isp_info`.

For `UXGPRO`, the gateway `port_table` is physical inventory plus optional
operator-provided assignment metadata. The daemon still does not create VLANs
or apply UniFi Network gateway settings to the host. Fields such as
`network_group`, `networkconf_id`, `native_networkconf_id`, `network_name`, and
`vlan` are controller/payload metadata only.

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

### `port-map`

Read-only explicit mapping mode. Each UniFi port can be assigned one source:

- `interface`: copy a physical OS interface into the payload.
- `disabled`: report the port administratively disabled, down, with speed `0`,
  and without learned MAC entries.
- `unmapped`: leave the profile port without a sensor source.

The validate path checks that explicitly mapped interfaces exist on the local
host. `disabled` and `unmapped` entries are valid without an OS interface.
Every profile port must have one `port_mappings[]` entry in this mode. Interface
data is read through the platform facade from `net.InterfaceByName`, interface
addresses, sysfs counters/speed/state on Linux, optional `/proc/net/dev`
counters, and best-effort `ifconfig`/`netstat` output.

```yaml
operation_mode: port-map
port_mappings:
  - port: 1
    interface: eno1
  - port: 2
    disabled: true
  - port: 3
    unmapped: true
  # Continue until every profile port has an explicit entry.
```

### `host-direct`

Direct host identity mode. It does not create a separate MAC or IP. The special
`mac: host` value is only allowed in this mode and requires
`observe_interface` so the daemon can read the host interface MAC explicitly.

### `macvlan`

Planning-only mode in this release. It is Linux-only and must be combined with
`-dry-run-plan`. The daemon prints the planned macvlan commands, but does not
execute them.

## Passive Sources

Passive sources are read-only and hang behind `internal/platform`. They enrich
payloads or status, but never mutate host networking.

```yaml
lldp_source: lldpd
traffic_rates_enabled: false
log_source: journalctl
proc_source: procfs
dbus_enabled: false
dbus_bus: system
```

`lldp_source: lldpd` runs `lldpcli -f json show neighbors` with a timeout and
maps known local interfaces to uplink, bridge-member, `port-map`, or
`port_overrides[].interface` ports. Missing `lldpcli` is reported as a warning
and does not stop the daemon.

LLDP is not required for adoption or for a manually configured topology hint.
When it is available, it reduces manual `uplink_neighbor` mistakes by learning
the upstream chassis and port from the host interface. When it is missing, keep
`uplink_neighbor` explicit and treat topology direction as controller-derived:
the controller may still prefer what real UniFi switches report about the same
MAC.

`log_source: journalctl` reads recent Linux unit logs through
`journalctl --output=json`. `log_source: syslog` reads a configured syslog file,
defaulting to `/var/log/messages` for FreeBSD-style systems. These sources are
exposed through status/capabilities and remain read-only.

`proc_source: procfs` is Linux-only and supplements interface counters from
`/proc/net/dev`; it does not replace `/sys/class/net` for link speed or media.

`traffic_rates_enabled: true` reports read-only RX/TX byte rates and the
available byte, packet, error, link-state, speed, media, and source-interface
metadata for mapped or observed interfaces in UniFi inform payloads. It is off
by default so existing labs keep their previous controller display. The rate
source is the same interface counter path used by
`port_overrides[].interface`, `port-map`, and bridge observation; it does not
enable packet capture, NetFlow/IPFIX, DPI, or packet/error rate fields.

`dbus_enabled: true` only checks optional system or session D-Bus connectivity.
D-Bus is not required for normal stub operation.

Traffic metadata is currently `traffic_source: off` only; packet capture and
DPI are intentionally out of scope for the first observation wave.

`management_lan` is the switch management VLAN configuration. Values `1..4094`
are reported in the inform payload and status output, while `0` leaves it
unset:

```yaml
management_lan:
  enabled: true
  vlan: 20
  network_name: Management
  mode: preexisting-interface
  interface: vmbr0.20
  ip: 192.0.2.50
  controller_reachable: off
  adoption_strategy: untagged-first
```

`mode: metadata-only` only reports the VLAN to the controller. `mode:
preexisting-interface` is the recommended first real mode: the VLAN interface
must already exist, and the daemon uses its IPv4 address for the reported
management IP, discovery source, and outbound inform source binding. `mode:
planned-host-vlan` is dry-run-plan only. The daemon still does not create VLAN
interfaces or apply controller provisioning to the host.

Gateway profiles do not use `management_lan` for WAN/LAN modeling. Keep
gateway data-plane assignments in `port_overrides` and keep controller
transport/management reachability in the top-level runtime fields.
