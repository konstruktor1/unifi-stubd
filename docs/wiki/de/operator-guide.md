# Betriebsleitfaden

Diese Seite ist die Runbook-Sicht fuer den sicheren Betrieb von `unifi-stubd`.
Sie ersetzt nicht die Referenzdokumentation.

## Vor dem Start

Pruefe diese Punkte, bevor Discovery- oder Inform-Traffic gesendet wird:

- isoliertes Lab- oder Management-Netz nutzen;
- disposable Fake-MAC nutzen, ausser physisches MAC-Verhalten wird getestet;
- in committeten Dateien nur Beispieladressen verwenden;
- Controller-Tokens, Adoption-Keys, SSH-Passwoerter, echte MAC-Tabellen und
  Captures aus Git heraushalten;
- vor Daemon-Betrieb `-validate` ausfuehren;
- vor Controller-Traffic `-dry-run` ausfuehren;
- entscheiden, ob das Geraet synthetisch, bridge-observed oder port-mapped sein
  soll.

## Empfohlener erster Befehl

```sh
go run ./cmd/unifi-stubd -validate -config packaging/linux/etc/unifi-stubd/config.yaml
```

Danach Payload ohne Netzwerk-Nebenwirkungen inspizieren:

```sh
go run ./cmd/unifi-stubd -dry-run -no-discovery
```

## Operation-Mode-Auswahl

| Modus | Sinnvoll wenn | Hauptrisiko |
| --- | --- | --- |
| `stub` | synthetisches UniFi-Geraet aus Profildaten gebraucht wird | Payload entspricht nicht zwingend einem echten Host |
| `bridge-observe` | eine Proxmox-/Linux-Bridge wie ein Switch aussehen soll | Topologie-Richtung bleibt controller-abgeleitet |
| `port-map` | jeder UniFi-Port auf ein bekanntes Host-Interface mappt | jeder Port braucht eine explizite Quelle |
| `host-direct` | die Host-Identitaet selbst dargestellt werden soll | nur nutzen, wenn diese Identitaet gewollt ist |
| `macvlan` | ein kuenftiger aktiver Host-Netzwerk-Modus geplant wird | in dieser Version nur dry-run-plan |

Details stehen in [Betriebsmodi](../../de/operation-modes.md).

## Bridge-Observe-Runbook

`bridge-observe` nutzen, wenn eine Host-Bridge einen Switch darstellen soll.
Die Bridge selbst ist kein UniFi-Port, sondern die Beobachtungsgrenze.

Minimalform:

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

Regeln:

- VM-/Container-Member als Access-Ports mappen;
- physisches Upstream-Interface als Uplink mappen;
- `uplink_port` explizit setzen, wenn der echte Link SFP/SFP+ ist;
- fuer Darstellungs-Tests synthetische lokal administrierte Stub-MAC bevorzugen;
- physische Host-MAC nur nutzen, wenn das Controller-Verhalten mit dieser
  echten MAC getestet wird.

## Port-Map-Runbook

`port-map` nutzen, wenn jeder dargestellte Port eine bewusste Quelle hat:

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

Regeln:

- jeder Profilport braucht genau ein Mapping;
- `interface`-Quellen muessen bei Validierung/Runtime existieren;
- `disabled` rendert Link down und Speed `0`;
- `unmapped` behaelt Profil-Defaults ohne Host-Sensor;
- der Daemon konfiguriert oder veraendert kein Host-Interface.

## Management LAN

Aktuelle Switch-Management-LAN-Unterstuetzung ist bewusst konservativ:

- `metadata-only`: VLAN nur in Payload/Status melden;
- `preexisting-interface`: Management-Identitaet an ein bereits existierendes
  Interface binden, zum Beispiel `vmbr0.20`;
- `planned-host-vlan`: nur dry-run-plan.

Der Daemon erzeugt in der aktuellen Version keine VLAN-Interfaces.

## Adoption und Cleanup

Fuer Controller-Tests disposable MACs nutzen. Ein sauberer Adoption-Zyklus ist:

1. `-dry-run` mit finalen MAC-/IP-/Profilwerten ausfuehren.
2. Einen Inform gegen den Controller senden.
3. Nur das disposable Device adoptieren.
4. Lokalen `STATE=connected` ueber Status pruefen.
5. Controller Forget/Remove fuer das disposable Device ausfuehren.
6. Stub stoppen.
7. Lokales temporaeres State-Verzeichnis loeschen, wenn der Test fertig ist.

Dieselbe MAC/IP nicht mit unterschiedlichen Profilen wiederverwenden, ausser
das vorherige Device wurde im Controller forgotten und lokaler Adoption-State
wurde zurueckgesetzt.

## Status-Pruefungen

```sh
unifi-stubd -status
unifi-stubd -status-json
```

Status sollte erklaeren:

- Profil und Operation-Mode;
- effektive MAC/IP/Hostname;
- Adoption-State ohne Authkey;
- Plattform-Capability-State;
- Observation-Warnungen;
- letztes Inform-Ergebnis.

## Haeufige Fehlerbilder

Device bleibt pending/adopting:
stale Controller-Device oder lokaler Adoption-State. Controller-Device
forgetten und lokalen State resetten.

Topologiekante zeigt falsch herum:
physische Host-MAC ist auch am Upstream-UniFi-Switch sichtbar. Synthetische
lokal administrierte Stub-MAC nutzen.

Uplink erscheint am falschen Port:
Mixed-Speed-Profil ohne `uplink_port`. Profilport explizit setzen.

Zu viele Clients erscheinen direkt angeschlossen:
Uplink-MACs sind nicht gefiltert oder der Uplink ist falsch klassifiziert.
`bridge_observe.uplink_interface` explizit setzen.

`port-map`-Validierung scheitert:
Mapping oder Interface fehlt. Einen gueltigen Eintrag pro Profilport setzen.
