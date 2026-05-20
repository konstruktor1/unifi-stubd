# Architekturkarte

Diese Seite erklaert, wo Code hingehoert. Die vollstaendige Referenz ist
[Architektur](../../de/architecture.md).

## Schichtenregel

```text
config/CLI -> Profil-Registry -> Plattform-Observation -> aufgeloeste Ports -> Payload -> Protokoll
```

Jede Schicht gibt typisierte Daten an die naechste Schicht weiter. Untere
Schichten sollen nicht auf CLI-Flags oder YAML-Dateien zurueckgreifen.

## Ownership nach Anliegen

### Konfiguration

Besitzt:

- `internal/config`
- `cmd/unifi-stubd/settings.go`
- `cmd/unifi-stubd/config.go`

Verantwortung:

- Defaults;
- striktes YAML-Laden;
- CLI-Override-Prioritaet;
- Validierungsinputs.

Hier keine Payload-Entscheide einbauen.

### Profile

Besitzt:

- `internal/device/profilemodel`
- `internal/device/profiledata`
- `internal/device/profiles/*`

Verantwortung:

- Modellidentitaet;
- Portlayout;
- Payload-Art;
- sichere Feature-Defaults;
- YAML-`extends`-Handling.

Nicht anhand von Profilnamen im Payload-Runtime-Code verzweigen, wenn
Profilfelder das Verhalten abbilden koennen.

### Observation

Besitzt:

- `internal/observe`
- `internal/platform`
- `internal/adapters/*`

Verantwortung:

- read-only Host-Fakten;
- Bridge-Member-Rollenklassifikation;
- Interface-Counter und Speed;
- LLDP-/Log-/procfs-/D-Bus-Capability-Daten;
- normalisierte `PortObservation` und `BridgeObservation`.

Hier kein Controller-JSON rendern.

### Payload

Besitzt:

- `internal/device`
- `internal/device/payload`

Verantwortung:

- Profilports, Overrides, Observations und Management-Metadaten in controller-
  seitige Strukturen mergen;
- Switch- und Gateway-Payload-Tabellen rendern;
- Port-Medium, Speed, MAC, Rolle und Network-Group synchron halten.

Hier keine OS-Kommandos ausfuehren und keine Live-Interfaces inspizieren.

### Protokoll

Besitzt:

- `internal/discovery`
- `internal/inform`
- `internal/adoption`
- `internal/adoptionssh`

Verantwortung:

- Discovery-TLVs;
- Inform-Packet-Framing und Crypto;
- HTTP-Response-Limits;
- Adoption-State-Parsing und Persistenz;
- minimale SSH-Kompatibilitaet.

Keine beliebigen Controller-Kommandos ausfuehren.

## Wichtige Datentypen

| Typ | Schicht | Bedeutung |
| --- | --- | --- |
| `config.Config` | config | rohe Runtime-Konfiguration nach YAML-Load |
| `runtimeFlags` | cmd | effektiver CLI-/YAML-Runtime-State |
| `profilemodel.Profile` | profile | kanonische Profildaten |
| `device.Profile` | device | aufgeloeste Profil-API fuer Runtime |
| `observe.BridgeObservation` | observe | read-only Bridge-Fakten |
| `observe.PortObservation` | observe | Interface- oder explizite Portquellen-Fakten |
| `device.Port` | device | aufgeloester controller-seitiger Portinput |
| `payload.PortView` | payload | normalisierte Renderer-Sicht |
| `adoption.Store` | adoption | lokaler persistenter Controller-State |

## Safety-Gates

Die wichtigsten Safety-Gates sind:

- strikte Config-/Profilvalidierung vor Runtime;
- Operation-Mode-Validierung in `cmd/unifi-stubd/operation.go`;
- read-only Plattform-Fassade;
- Adoption-Parser statt Shell-Ausfuehrung;
- lokaler Reset fuer Controller Forget/Restore-Default;
- Status-Sanitisierung vor JSON-/Human-Output.

## Wenn ein Feature dazukommt

Platzierungsregel:

| Feature-Art | Wahrscheinlicher Ort |
| --- | --- |
| Neues YAML-Feld | `internal/config`, Schemas, Paket-Configs, Doku |
| Neue Profildaten | `internal/device/profilemodel`, Profil-YAML, Validierung |
| Neue OS-Lesequelle | `internal/platform` oder `internal/adapters` |
| Neue Bridge-Klassifikationsregel | `internal/observe/classify.go` plus Tests |
| Neues Payload-Feld | `internal/device/payload` plus Fixture-Tests |
| Neue Controller-Antwort | `internal/adoption` oder `internal/adoptionssh` |
| Neue CLI-Validierung | `cmd/unifi-stubd/operation.go` |

