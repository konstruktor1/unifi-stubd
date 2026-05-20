# unifi-stubd Wiki

Dieses Wiki ist die praktische Navigationsschicht fuer `unifi-stubd`. Nutze es,
wenn du verstehen willst, was das Projekt ist, wo eine Aenderung hingehoert, wie
der Stub sicher betrieben wird und welche Tests eine Aenderung belegen.

## Projekt in einem Absatz

`unifi-stubd` ist ein experimentelles Lab-Tool, das einen Linux-Host, eine
Proxmox-Bridge, einen FreeBSD-/OPNsense-Host oder eine VM als minimales UniFi-
Geraet im UniFi Network Controller erscheinen laesst. Es kann Switch-artige
Geraete und experimentelle Gateway-Identitaeten emulieren, darf aber kein vom
Controller gemanagter Host-Agent werden. Der Controller darf lokalen Adoption-
State aktualisieren; er darf den Host nicht provisionieren.

## Startpunkte

| Ziel | Seite |
| --- | --- |
| Stub sicher installieren oder laufen lassen | [Betriebsleitfaden](operator-guide.md) |
| Code-Ownership und Datenfluss verstehen | [Architekturkarte](architecture-map.md) |
| Tests planen oder ausfuehren | [Testleitfaden](testing-guide.md) |
| Aktuelle Architekturentscheide verstehen | [Entscheidungen](decisions.md) |
| Vollstaendige Dokumentation lesen | [Deutsche Dokumentation](../../de/README.md) |
| Architektur-Referenz lesen | [Architektur](../../de/architecture.md) |
| Betriebsmodi im Detail lesen | [Betriebsmodi](../../de/operation-modes.md) |
| Projektstand lesen | [Projektstand](../../de/project-status.md) |
| Roadmap lesen | [Roadmap](../../de/roadmap.md) |

## Safety-Grenze

Das Projekt hat eine nicht verhandelbare Grenze:

- Controller-Antworten duerfen lokalen Stub-State aktualisieren;
- read-only Observation darf Payload und Status anreichern;
- Controller-Provisioning darf kein Host-Netzwerk veraendern und keine
  beliebigen Shell-Kommandos ausfuehren.

Das gilt fuer `stub`, `bridge-observe`, `port-map`, Management-LAN-Metadaten,
passives LLDP, Logs, D-Bus-Capability-Pruefungen und kuenftige Plattform-
Adapter.

## Aktuelle Produktform

Unterstuetzt und aktiv:

- Switch-artige Discovery- und Inform-Payloads;
- Advanced-Adoption-SSH-Shim mit begrenzter Kommandoverarbeitung;
- Adoption-State-Persistenz und Restore-Default-/Forget-Reset-Verhalten;
- datengetriebene Built-in- und externe Profile;
- `bridge-observe` fuer Proxmox-/Linux-Bridge-Darstellung;
- `port-map` fuer explizites Port-zu-Interface-Mapping;
- Linux-Plattform-Fassade fuer sysfs, procfs, journalctl, D-Bus-Probe und lldpd;
- konservative FreeBSD-/OPNsense-Unterstuetzung;
- Docker-Controller-Integrationstests.

Experimentell:

- Gateway-Identitaetsprofile;
- vollstaendige Topologie-Darstellung;
- Management-LAN-Modellierung jenseits von Metadaten und preexisting interface;
- FreeBSD-Bridge-Observe-Paritaet.

Nicht-Ziele:

- produktiver Gateway-Ersatz;
- vollstaendiger UniFi-OS-Ersatz;
- blindes Controller-Provisioning;
- automatische Host-VLAN-Erzeugung in der aktuellen Version.

## Repository-Karte

| Pfad | Zweck |
| --- | --- |
| `cmd/unifi-stubd/` | CLI, Config-Layering, Validierung, Daemon-Orchestrierung |
| `internal/config/` | YAML-Schema und Defaults |
| `internal/device/` | Profile, aufgeloeste Ports, Payload-Einstieg |
| `internal/device/payload/` | Switch-/Gateway-JSON-Payload-Renderer |
| `internal/observe/` | read-only Observation-Modell |
| `internal/platform/` | OS-Fassade fuer read-only Host-Integrationen |
| `internal/inform/` | Inform-Packet-Crypto, Padding, HTTP-Response-Handling |
| `internal/adoption/` | Adoption-Response-Parsing und lokaler State |
| `internal/adoptionssh/` | minimaler SSH-Kompatibilitaets-Shim |
| `tests/` | alle Go-Tests |
| `docs/en/`, `docs/de/` | detaillierte User- und Projektdoku |
| `docs/wiki/` | diese Navigations- und Runbook-Schicht |

## Pflege-Regel

Wenn eine Aenderung Verhalten betrifft, zuerst die Detaildokumentation
aktualisieren und danach das Wiki nur anpassen, wenn sich Navigation oder
Betriebssicht aendern. Keine ganzen Referenzabschnitte ins Wiki kopieren; kurz
zusammenfassen und verlinken.

