# Architecture

`unifi-stubd` is a controller-facing device emulator. It renders a Linux,
FreeBSD, or bridge-like host as a minimal UniFi switch or experimental gateway
identity, but it is not a host provisioning agent. The architecture therefore
separates four concerns that must stay independent:

- configuration and profile resolution;
- read-only host observation;
- controller-facing payload rendering;
- adoption state persistence.

The main invariant is simple: controller data may change local stub state, but
it must not mutate host networking, users, packages, services, routes, firewall
rules, or files outside the explicit stub state paths.

## Component Graph

```mermaid
flowchart LR
  Cfg["YAML config + CLI overrides"] --> Cmd["cmd/unifi-stubd"]
  Profiles["Built-in and external profiles"] --> Registry["profile registry"]
  Registry --> PortView["resolved port view"]
  Cmd --> Registry
  Cmd --> Platform["internal/platform"]
  Platform --> Observe["internal/observe model"]
  Linux["Linux: bridge, sysfs, procfs, journalctl, D-Bus, lldpd"] --> Platform
  BSD["FreeBSD: ifconfig, netstat, syslog, lldpd"] --> Platform
  PortMap["explicit port-map"] --> Observe
  Observe --> PortView
  PortView --> Payload["internal/device/payload"]
  Payload --> Inform["internal/inform"]
  Discovery["internal/discovery"] --> Controller["UniFi Network Controller"]
  Inform --> Controller
  Controller --> Adoption["internal/adoption parser"]
  Adoption --> State["adoption.env"]
  State --> Inform
  Cmd --> Status["status.json / --status"]
  Platform --> Status
  Adoption --> Status
```

## Package Ownership

| Package | Responsibility | Must not own |
| --- | --- | --- |
| `cmd/unifi-stubd` | CLI, config layering, validation, daemon orchestration, operation-mode wiring | packet encoding details, OS-specific parsing, controller protocol internals |
| `internal/config` | user-facing YAML schema, defaults, strict config loading | runtime network reads, payload construction |
| `internal/device` | canonical profile structs, built-in/external profile registry, YAML-level `extends`, validation, export/templates, port definitions, and port resolution | controller HTTP, OS commands, JSON payload table encoding |
| `internal/device/payload` | switch/gateway JSON payload rendering from typed device profile, identity, and resolved ports | profile loading, port generation, model-specific runtime guesses, host networking |
| `internal/observe` | normalized read-only observations for bridge and port-map modes | direct Linux/FreeBSD command execution outside adapters |
| `internal/platform` | read-only OS facade for interfaces, bridges, LLDP, logs, procfs, D-Bus, capabilities | payload rendering, controller adoption decisions |
| `internal/adapters/linuxbridge` | Linux bridge FDB command parsing | payload policy |
| `internal/adapters/freebsdifconfig` | FreeBSD bridge/interface parser helpers | payload policy |
| `internal/discovery` | UDP discovery TLV construction | adoption state |
| `internal/inform` | `TNBU` framing, padding, encryption, HTTP response limits | profile resolution |
| `internal/adoption` | controller response parsing and local adoption store | executing controller shell/provisioning commands |
| `internal/adoptionssh` | minimal SSH compatibility shim for advanced adoption | arbitrary shell execution |

## Runtime Flow

1. `cmd/unifi-stubd` builds defaults from `internal/config`.
2. YAML config is loaded, if configured. Unknown YAML fields are errors.
3. Explicit CLI flags override YAML values.
4. Built-in profiles are registered first; optional `profile_file` and
   `profile_dir` entries are loaded afterwards.
5. External profile `extends` is resolved at YAML mapping level, then decoded
   once into `device.Profile`. This preserves intentional zero-value
   overrides such as `recommended: false`, `payload.has_dpi: false`, or
   `port_names: []`.
6. Each runtime instance must select one concrete profile through YAML or
   `-profile`; an empty or unknown selection is rejected.
7. Operation-mode validation checks the resolved profile, port count,
   management-LAN intent, bridge/port-map settings, and dry-run ability.
8. The platform facade reads optional host facts only when the selected mode or
   status path asks for them.
9. Profile ports, configured overrides, passive observations, management-LAN
   metadata, and neighbor hints are merged into `[]device.Port`.
10. `internal/device/payload` renders the controller-facing JSON payload from
   the profile, identity, and resolved ports.
11. `internal/discovery` and `internal/inform` send discovery/inform traffic
    unless validation or dry-run mode stops before network I/O.
12. Controller responses update only the local adoption store and status.

`-validate` runs the same config/profile/operation checks without sending
packets, starting listeners, or writing normalized config files.

## Configuration And Profiles

Configuration is intentionally explicit at the service boundary. The packaged
YAML shows every supported field with conservative defaults. Runtime defaults
live in `internal/config.Default`, while CLI registration and YAML application
share the setting registry in `cmd/unifi-stubd/settings.go`.

Profiles describe device shape. They are data, not code hooks. A profile owns:

- model identity and display metadata;
- `device_type` and payload kind;
- port count, names, roles, network groups, media, and speed layout;
- payload defaults such as management interface name, gateway interface prefix,
  required controller version, and safe feature flags.

Separately, the device package owns an evidence-bound `ProtocolProfile`
registry for exact analysed builds. Its immutable key contains family,
platform/component role, version/build marker, firmware SHA-256,
component/app SHA-256, and digest kind. An exact build selects only a
version-independent `ProtocolBehaviorID`; a separate
`ProtocolBehaviorProfile` catalog composes the six typed metadata axes: wire,
discovery, field groups, commands, positive capabilities, and presence.
Lookup never falls back by version prefix, family similarity, or a neighboring
digest. An unknown build or behavior ID selects the composition whose six
axes are all `unsupported`.

Exact evidence identity and implemented behavior are deliberately different
layers. An evidence row does not create a complete firmware-specific behavior
type. Instead, it selects a composition of small, reusable contracts such as a
TNBU codec, discovery dialect, common identity fields, switch or gateway field
groups, command namespace, platform capabilities, and presence rules. Builds
with the same evidenced semantics share those components. A new firmware
version creates a new behavior component only when a reviewed differential
proves a relevant semantic change.

Registry metadata keeps the evidence identity and recovered principle for
review. Executable proof is required only when a behavior is accepted into a
named product requirement; research-only selectors do not create their own
product tests.

Protocol platforms come only from the architecture column of
`device-types.tsv`: `aarch64-little-endian`, `arm32-little-endian-eabi5`,
`mips32r2-little-endian`, `mips32r2-big-endian`, `mips32-big-endian`,
`arm-cortex-m-thumb-little-endian`, `xtensa-esp32-little-endian`, and explicit
artifact-only `unknown`. Exact subplatform text remains in the evidence build
marker, family, or component role; it is not promoted into a new platform
class. Any value outside this evidence taxonomy fails closed.

Behavior components are named after their contract or protocol generation,
not after every product and release. Product/build labels remain in the
evidence-binding layer. Platform differences are expressed as narrow
positive capability flags or presence overlays; for example, mainline BCM and
MSCC share their mainline switch base while retaining distinct platform and
table gates. Missing capability flags do not assert that physical hardware is
absent. They only prevent unproven vocabulary or behavior from being selected.
Likewise, the common zlib wire component is the smallest safe controller-side
subset and does not enable BCM's observed Snappy branch. Mainline and
Switch-WAN retain separate discovery components because their alternate
builders differ. This prevents both unsafe nearest-profile fallback and a
one-to-one copy of the firmware matrix into product code.

This registry is the selector contract for later wire, discovery, command, and
payload implementations. GOV-003 does not bind those research entries to the
existing user-facing shape profiles and does not change current runtime
traffic by itself.

The Core is a compatibility stub, not a firmware emulator and not a goal of
one-to-one device simulation. It implements only the smallest safe,
controller-facing contracts needed for the selected product use cases.
Optional synthetic scenarios may vary adoption state, link state, counters,
or command results deterministically for tests, but they do not reproduce
vendor processes, internal IPC, hardware side effects, incidental timing, or
every observed payload. Reverse-engineering evidence constrains these
contracts; it is not copied as an implementation blueprint.

### Evidence And Product Traceability

Reverse-engineering evidence is research input, not an automatic product
contract. A behavior accepted into a named requirement keeps the smallest
reviewable chain from requirement through recovered principle to product
component and executable product proof. Deferred, rejected, or superseded lab
findings may remain in the companion research record without product code or a
regression test.

Cleanup may delete isolated patch functions, fixtures, and tests that protect
no accepted product behavior. It must preserve the public product outcome and
safety boundary. The product architecture, not the size of the research
inventory, decides what is implemented.

The device package is the single source of truth for profile and port data.
Renderer-specific views are derived from `device.Profile` and `device.Port`;
they are not separate profile or port models.

Live daemon starts use a host-global advisory instance guard before SSH,
discovery, or inform traffic starts. `instance_guard: fail` is the default and
aborts a second live instance on the same host. Intentional multi-stub labs must
opt in explicitly with `instance_guard: warn` or `instance_guard: off` and a
reviewed lock path.

The payload renderer must not branch on model names such as `UXG`, `UXGPRO`, or
`US48P500` to decide behavior. Renderer behavior is driven by profile fields:
`payload.kind`, port roles, port media, management interface, and explicit
overrides. Protocol-specific rendering is assembled from shared field groups,
capabilities, and presence rules rather than one complete schema per firmware
build.

## Operation Modes

### `stub`

`stub` is synthetic. It uses profile data, configured neighbors, optional
per-port overrides, and generated counters. It does not inspect a bridge unless
explicit legacy observation flags are configured for compatibility.

### `bridge-observe`

`bridge-observe` represents a host bridge as a virtual UniFi switch. The bridge
is the observation boundary. It is read-only: no VLAN, bridge, route, or
interface mutation is performed.

The bridge member classifier assigns roles before MACs reach the payload:

- `bridge`: the bridge device itself, stored as metadata, not a UniFi port;
- `uplink`: the configured physical upstream interface, or the only safe
  physical-looking candidate when no explicit uplink is configured;
- `access`: VM/container style members such as `tap*`, `veth*`, `fwpr*`,
  `fwln*`, `fwbr*`, `epair*`, and `vnet*`;
- `unknown`: kept eligible for deterministic mapping when the platform cannot
  classify it safely;
- `ignored`: excluded from payload mapping.

MACs learned on the uplink are remote behind the real upstream switch and are
filtered out of local access-port MAC tables. Unused profile ports are reported
disconnected, not synthetic-up. This prevents the controller from seeing the
upstream switch or its downstream clients as directly attached to the virtual
stub switch.

`uplink_port` controls which profile port carries the represented physical
uplink. This matters for mixed-speed profiles: a 48-port switch with SFP/SFP+
cages should put a 10G upstream on the intended SFP+ port, not on the last GE
port by accident.

### `port-map`

`port-map` requires one explicit source per profile port:

- `interface`: read one host interface and copy MAC, IP, link, speed/media, and
  counters when available;
- `disabled`: render link down and speed `0`;
- `unmapped`: keep profile defaults without a physical sensor source.

This mode is for appliances or VMs where each represented UniFi port maps to a
known host NIC. It is still read-only.

## Platform Facade

`internal/platform` is the only runtime boundary allowed to talk to optional
host facilities. It exposes small interfaces:

- `InterfaceReader` for link state, addresses, speed/media, and counters;
- `BridgeReader` through `observe.ObservationSource`;
- `LLDPReader` for passive `lldpcli -f json show neighbors`;
- `LogReader` for `journalctl` or syslog files;
- `ProcReader` for Linux `/proc/net/dev` counters;
- `ServiceBus` for optional D-Bus capability checks.

The platform layer reports capability state as `disabled`, `available`,
`missing`, or `unsupported`. Missing optional tools are status warnings, not a
reason to mutate the host or install dependencies.

## Payload Boundary

Payload rendering starts after profile and observation data are resolved. The
renderer receives:

- the resolved `device.Profile`;
- controller-facing identity;
- a resolved ordered port list.

Switch payloads derive `if_table`, `ethernet_table`, and `port_table` from that
data. Gateway payloads derive `if_table`, `network_table`, `port_table`,
`config_port_table`, `ethernet_overrides`, `uplink_table`, `reported_networks`,
and WAN status rows from the same data. Explicit operator-provided assignment
metadata such as `portconf_id`, `networkconf_id`, `native_networkconf_id`,
`network_name`, and `vlan` may be mirrored in the gateway port tables. This
keeps speed, media, MAC, IP, source interface, WAN/LAN role, network group,
counters, MAC table entries, and management-LAN metadata aligned.
For switches, the selected physical uplink's counters and explicitly measured
rates are also projected onto the management `if_table` row named by the
top-level `uplink` selector; local source-interface identity remains on the
physical port view.

Management LAN is payload metadata or binding to a preexisting local interface.
`planned-host-vlan` remains dry-run-plan only. The daemon does not create host
VLANs.

## Controller Boundary

The controller-facing protocol has three separate paths:

- UDP discovery advertises the fake device identity;
- HTTP inform sends encrypted `TNBU` payloads and receives adoption/provisioning
  responses;
- the optional SSH shim accepts only the minimal commands needed for advanced
  adoption compatibility.

The payload boundary owns the exact typed Inform states `0 Unknown`,
`1 Default`, `2 Managed`, `3 Selfrun`, and `4 Upgrade`. The active adoption
lifecycle is the observed UXG Pro path `1 -> 0 -> 2`: the accepted key/cipher
change makes the immediately following Inform use state `0`, and the captured
next `noop` leads to state `2`. Static firmware analysis
shows that successful `noop` and `setparam` responses enter Managed before an
authkey-bearing `setparam` or changed Inform URL returns the reporter to Unknown. It does
not establish the former synthetic restart chains. States `3` and `4` are
status vocabulary only in the Stub and do not authorize host actions.

Accepted controller adoption configuration may update the allowlisted
`AUTHKEY`, `CFGVERSION`, `USE_AES_GCM`, and `INFORM_URL`; the product derives
`STATE` from those accepted transitions. Persisted `VERSION` is a legacy field
whose provenance may be controller-learned in older releases. It is read for
compatibility; when it is absent, the profile/flag-selected operator version
applies. A controller `_type: upgrade` response is retained
as ignored status only; it does not rewrite the advertised operator version,
lifecycle state, firmware, services, or host.

The Stub may clear its own local adoption state, but it never maps a controller
reset request to a vendor factory reset, host restart, network change, service
change, or clock change. This is a local safety boundary, not a reconstructed
firmware lifecycle.

Block selection and controller time are process-local reporting inputs. Every
accepted `noop` replaces the previous block selection in full; each payload
producer consults only the exact include/exclude identifiers it owns. A string
`server_time_in_utc` on `noop` establishes a controller-minus-device offset for
later reported payload time only. Neither block descriptors nor controller time
can execute commands or mutate the host.

Provisioning commands that would affect the host are summarized as ignored
metadata. Arbitrary controller shell commands are not executed. The sole
command-result exception is an argument-free `wg-genkeypair` with a request ID:
the daemon generates an X25519 key pair in process and attaches the correlated
result to the next inform. The keys and request ID remain memory-only; logs and
status expose only the fixed command name and sanitized lifecycle counters.
Sanitized adoption and last-inform status expose both the numeric wire state
and its fixed name; they never expose the individual adoption key.

The presence loop uses a resettable timer rather than a fixed ticker. The
operator's `interval_seconds` is the initial interval. Positive controller
`noop.interval` hints are enabled only when the built-in source, profile, model,
version, and board identity select one exact LIFE-010 firmware-minimum binding.
The current candidate enables six exact tuples and leaves the other six
built-ins, external profiles, adjacent identities, and model/version overrides
fail-closed for interval hints. The exact list belongs in the protocol notes.

For an enabled tuple, the lower bound is
`max(firmware minimum, operator minimum)` and the upper bound is
`min(operator maximum, 300 seconds)`. Without a binding, a positive hint is
ignored and the operator base remains active. Configuration fails validation
when the effective upper bound is below the effective lower bound. Dry-run plan
and sanitized status show whether hints are supported, the evidence ID,
firmware and effective minima, or the fixed rejection reason.
`noop.immediate=true` requests one
bounded, nonzero near-immediate follow-up without changing that retained
effective interval. Transport, response-decode, JSON, dispatcher, and non-200
failures double the effective delay up to the configured backoff maximum and
add bounded symmetric jitter. Only an empty HTTP 200 or an accepted
controller-object 200 resets the failure count; the most recent effective
interval, next delay, failure count, and backoff state are recorded as
sanitized status.

The exact interval bindings remain candidate behavior until the selected
product candidate is accepted against its designated controller environment.

## Persistence And Status

Persistent state is intentionally narrow:

| File | Purpose |
| --- | --- |
| `/var/lib/unifi-stubd/adoption.env` | Linux adoption store: state, authkey, cfgversion, inform URL, cipher preference |
| `/var/lib/unifi-stubd/status.json` | Linux sanitized runtime status |
| `/var/db/unifi-stubd/adoption.env` | FreeBSD adoption store |
| `/var/db/unifi-stubd/status.json` | FreeBSD sanitized runtime status |

Adoption and runtime status updates are written to same-directory temporary
files, synced, and atomically renamed. An interrupted update therefore leaves
the previous complete state file in place rather than exposing partial data.
`SIGINT` and `SIGTERM` both stop the daemon through its normal cleanup path.

Status output must be useful for operators but safe to share. It reports
identity, selected profile, operation mode, observation summary, platform
capabilities, and last inform result. It must not print adoption authkeys,
controller tokens, SSH passwords, private captures, or raw lab secrets.

## Topology Limits

UniFi Network calculates topology links from several controller-side signals.
The stub can provide hints, but it cannot fully control the controller's graph.

Topology resolution separates direct infrastructure neighbors from client/FDB
MAC rows. The active represented uplink determines upstream direction; direct
infrastructure on other ports is downstream. Configured neighbors override
LLDP, while FDB visibility never determines direction. Switch renderers emit
the selected upstream through `uplink` and `last_uplink` and keep downstream
infrastructure on its physical port. Real device and port MACs remain valid
identities and are not replaced for topology rendering.

LLDP is helpful but optional. Without LLDP, `uplink_neighbor` is manual. With
LLDP, neighbor discovery should reduce manual mistakes, but the controller still
owns final topology rendering.

## Safety Rules

These rules are architectural, not just implementation preferences:

- no controller-provided shell command is executed as a shell;
- no controller provisioning data is applied to host networking;
- no package manager, service manager, firewall, routing, or user database is
  changed by adoption;
- `bridge-observe` and `port-map` are read-only observations;
- `management_lan` never creates VLAN devices in the current release;
- private controller URLs, tokens, real MAC tables, and captures stay out of
  Git;
- profile changes under the same MAC are not migrated; use a new fake MAC,
  controller forget, or local reset.

## OS Support Matrix

| Source | Linux | FreeBSD/OPNsense |
| --- | --- | --- |
| `stub` profile payload | supported | supported |
| `bridge-observe` FDB | `bridge fdb show br <bridge>` plus member role classification | `ifconfig <bridge> addr` parser plus the same member role model; runtime parity remains conservative |
| `bridge-observe` counters/speed | `/sys/class/net`, optional `/proc/net/dev` counters | planned for richer bridge counters; `port-map` can read interface metadata |
| `port-map` interface metadata | `net.Interface`, sysfs, optional command fallbacks | `net.Interface`, optional `ifconfig`/`netstat` |
| passive LLDP | `lldpcli -f json show neighbors` | `lldpcli -f json show neighbors` when lldpd is installed |
| logs | optional `journalctl --output=json` | optional syslog file, default `/var/log/messages` |
| D-Bus | optional system/session bus availability check | optional best-effort availability check |
| event subscription | planned | planned |
| native helpers/C++ | not used | not used |

## Known Extension Points

- `uplink_neighbor.remote_port` as explicit topology metadata.
- LLDP-derived uplink neighbor selection with manual override.
- richer FreeBSD media/counter readers, possibly through `x/sys/unix`.
- event subscriptions for bridge/FDB changes instead of polling.
- active macvlan/ipvlan or VLAN lifecycle only after a separate reviewed design.
