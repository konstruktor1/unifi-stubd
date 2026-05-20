# Architecture Map

This page explains where code belongs. The full reference is
[Architecture](../../en/architecture.md).

## Layer Rule

```text
config/CLI -> profile registry -> platform observation -> resolved ports -> payload -> protocol
```

Each layer should pass typed data to the next layer. Lower layers should not
reach back into CLI flags or YAML files.

## Ownership By Concern

### Configuration

Owned by:

- `internal/config`
- `cmd/unifi-stubd/settings.go`
- `cmd/unifi-stubd/config.go`

Responsibilities:

- defaults;
- strict YAML loading;
- CLI override priority;
- validation inputs.

Do not add payload decisions here.

### Profiles

Owned by:

- `internal/device/profilemodel`
- `internal/device/profiledata`
- `internal/device/profiles/*`

Responsibilities:

- model identity;
- port layout;
- payload kind;
- safe feature defaults;
- YAML `extends` handling.

Do not branch on profile names in payload runtime code when profile fields can
represent the behavior.

### Observation

Owned by:

- `internal/observe`
- `internal/platform`
- `internal/adapters/*`

Responsibilities:

- read-only host facts;
- bridge member role classification;
- interface counters and speed;
- LLDP/log/procfs/D-Bus capability data;
- normalized `PortObservation` and `BridgeObservation`.

Do not render controller JSON here.

### Payload

Owned by:

- `internal/device`
- `internal/device/payload`

Responsibilities:

- merge profile ports, overrides, observations, and management metadata into
  controller-facing structures;
- render switch and gateway payload tables;
- keep port media, speed, MAC, role, and network group synchronized.

Do not call OS commands or inspect live interfaces here.

### Protocol

Owned by:

- `internal/discovery`
- `internal/inform`
- `internal/adoption`
- `internal/adoptionssh`

Responsibilities:

- discovery TLVs;
- inform packet framing and crypto;
- HTTP response limits;
- adoption-state parsing and persistence;
- minimal SSH compatibility.

Do not execute arbitrary controller commands.

## Main Data Types

| Type | Layer | Meaning |
| --- | --- | --- |
| `config.Config` | config | raw runtime configuration after YAML load |
| `runtimeFlags` | cmd | effective CLI/YAML runtime state |
| `profilemodel.Profile` | profile | canonical profile data |
| `device.Profile` | device | resolved profile API used by runtime |
| `observe.BridgeObservation` | observe | bridge-level read-only facts |
| `observe.PortObservation` | observe | interface or explicit port source facts |
| `device.Port` | device | resolved controller-facing port input |
| `payload.PortView` | payload | normalized renderer view |
| `adoption.Store` | adoption | local persistent controller state |

## Safety Gates

The most important safety gates are:

- strict config/profile validation before runtime;
- operation-mode validation in `cmd/unifi-stubd/operation.go`;
- read-only platform facade;
- adoption parser instead of shell execution;
- local reset for controller forget/restore-default;
- status sanitization before JSON/human output.

## When Adding A Feature

Use this placement guide:

| Feature type | Likely place |
| --- | --- |
| New YAML field | `internal/config`, schemas, packaged configs, docs |
| New profile data | `internal/device/profilemodel`, profile YAML, validation |
| New OS read source | `internal/platform` or `internal/adapters` |
| New bridge classification rule | `internal/observe/classify.go` plus tests |
| New payload field | `internal/device/payload` plus fixture tests |
| New controller response | `internal/adoption` or `internal/adoptionssh` |
| New CLI validation | `cmd/unifi-stubd/operation.go` |

