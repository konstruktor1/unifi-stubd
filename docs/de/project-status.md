# Projektstand

Zuletzt aktualisiert: 2026-05-20.

Diese Seite beschreibt die Produktlinie `unifi-stubd`: den Go-Daemon, die
oeffentliche Konfigurationsoberflaeche, das Payload-Modell, Sicherheitsgrenzen,
Packaging und den Validierungsstand. Firmware-Wrapper und UDM-Pro-SE-VM-
Experimente sind separat im [Lab-Projektstand](lab-project-status.md)
dokumentiert.

## Produktdefinition

`unifi-stubd` ist ein lab-fokussierter UniFi-Device-Stub. Er laesst einen
Linux- oder FreeBSD-Host, eine VM, einen Container oder eine Bridge fuer
kontrollierte Experimente wie ein minimales UniFi-Geraet gegenueber einem
UniFi-Network-Controller erscheinen.

Das Produkt ist kein Gateway-Ersatz und wendet Controller-Provisioning nicht
auf den Host an. Es besitzt nur Stub-Identitaet, Payload-Erzeugung,
Adoption-State-Persistenz, read-only Host-Observation, Packaging und lokale
Test-Harnesses.

## Aktuelle Faehigkeiten

Der Daemon bietet aktuell:

- Discovery- und Inform-Framing fuer synthetische UniFi-Device-Identitaeten.
- Eingebaute datengetriebene Switch-Profile und experimentelle Gateway-
  Identitaetsprofile.
- Laden, Validieren, Exportieren und Vorlagen fuer externe YAML-Profile.
- YAML-Service-Konfiguration plus CLI-Overrides und einen `-validate`-Pfad.
- Die Betriebsmodi `stub`, `bridge-observe` und `port-map`.
- Read-only Linux-Observation ueber sysfs/procfs, Bridge-FDB-Daten,
  journalctl-Checks, optionale D-Bus-Verfuegbarkeitspruefung und passives LLDP
  ueber `lldpd`.
- Read-only FreeBSD-Observation ueber ifconfig/netstat/syslog-orientierte
  Adapter.
- Adoption-State-Persistenz fuer Controller-Inform-Antworten.
- Einen begrenzten Adoption-SSH-Shim fuer Advanced-Adoption-Kompatibilitaet.
- Docker-Controller-Integrationstests gegen das projekt-eigene UniFi-Network-
  Application-Lab.
- Linux- und FreeBSD-Paket-/Build-Ziele.

## Betriebsmodi

`stub` ist voll synthetisch. Der Modus nutzt das ausgewaehlte Profil und lokale
Konfiguration, um ein deterministisches controller-seitiges Geraet zu rendern.

`bridge-observe` repraesentiert einen Bridge-artigen Host-Aufbau, zum Beispiel
eine Proxmox-Bridge. Die Bridge ist die Beobachtungsgrenze; gelernte
Teilnehmer-MACs werden in die virtuelle Port-Tabelle projiziert, ohne Host-
Netzwerk zu mutieren.
Bridge-Member werden vor der Projektion klassifiziert: Das Bridge-Device gilt
als Backplane-Metadatum, VM-/Container-Member werden Access-Ports und der
konfigurierte physische Uplink bleibt von lokalen Teilnehmern getrennt.
MACs, die auf dem physischen Uplink gelernt werden, gelten als Remote hinter
dem echten Upstream-Switch und werden aus lokalen Access-Port-MAC-Tabellen
gefiltert.

`port-map` mapped UniFi-Profilports auf explizite OS-Interfaces oder deklariert
sie als `disabled` oder `unmapped`. Gemappte Ports uebernehmen read-only
physische Eigenschaften wie MAC-Adresse, Link-State, Speed/Media soweit
verfuegbar, Adressen, Counter und LLDP-Nachbarn.

Bei Proxmox-artiger Bridge-Darstellung ist die Topologie controller-abhaengig.
Ein konfigurierter `uplink_neighbor` kann UniFi Network dazu bringen, den
Upstream und die letzte Verbindung auf dem Stub-Uplink-Port zu melden. Wenn der
Stub die physische Host- oder Bridge-MAC meldet, kann ein echter
Upstream-UniFi-Switch diese MAC aber bereits auf einem eigenen Port sehen.
UniFi Network kann dann diese reale Beobachtung bevorzugen und die Link-Richtung
falsch herum darstellen. Reine Darstellungs-Tests sollten deshalb eine
synthetische lokal administrierte Stub-MAC bevorzugen, ausser genau diese
Physical-MAC-Heuristik soll getestet werden.

## Konfigurationsoberflaeche

Die Konfiguration ist absichtlich explizit:

- Device-Identitaet kommt aus Profildaten plus CLI-/Config-Overrides.
- Hardware-Form kommt aus eingebauten oder externen YAML-Profilen.
- Lab-spezifische Port-Zuweisungen kommen aus `port_overrides`, Observation-
  Modus-Konfiguration und `port_mappings`.
- Switch-Management-LAN-Absicht wird ueber `management_lan` modelliert.
- Die alte oeffentliche `management_vlan`-Konfiguration wurde entfernt.
  Controller-facing Payload-Felder behalten UniFi-kompatible Namen, wo das
  erforderlich ist.

`-validate` prueft die vollstaendige Runtime-Konfiguration, ohne den Daemon zu
starten. `-profile-validate` prueft Profil-Dateien oder Verzeichnisse isoliert.

## Sicherheitsgrenze

Die Sicherheitsgrenze bleibt die wichtigste Produktregel:

- Controller-getriggerte Restarts, Upgrades, Shell-, Firewall-, Routen-, User-
  und Host-Netzwerkaenderungen werden nicht blind ausgefuehrt.
- Controller-Provisioning-Daten duerfen geparst und zusammengefasst werden,
  bleiben aber Metadaten, solange kein spaeterer gepruefter lokaler Adapter
  eine enge Aktion explizit implementiert.
- Der SSH-Shim erkennt nur die kleine Kommandoform, die fuer Adoption und
  lokalen Stub-Reset noetig ist.
- Discovery- und Inform-Traffic bleiben opt-in und gehoeren nur in isolierte
  Lab- oder Management-Netze.

## Validierungsstand

Aktuelle automatisierte Validierung umfasst:

- `go test ./...`
- `make check`
- Paket-Build-Ziele ueber `make package`
- Konfigurations-/Profilvalidierung fuer paketierte Linux- und FreeBSD-Configs
- Docker-Integrationstests fuer Dry-Run-Payloads, Inform-MITM-Capture,
  Controller-Pending-State, controller-getriggerte Adoption und persistierten
  lokalen Adoption-State

Der Docker-Integrationspfad deckt auch `bridge-observe`, `port-map`, Gateway-
Payload-Rendering und den aktuellen Switch-Pfad `management_lan.mode:
preexisting-interface` ab.

Manuelle Realhost-Validierung am 2026-05-20 hat zusaetzlich bestaetigt:

- eine Linux-Bridge kann als 48-Port-Switch-Profil dargestellt werden, mit
  VM-/Container-Teilnehmern auf normalen Access-Ports und physischem Uplink auf
  einem dedizierten SFP+-Profil-Port via `uplink_port`;
- ungenutzte Bridge-Ports sollten getrennt gemeldet werden, nicht synthetisch
  als up;
- explizite `uplink_neighbor`-Metadaten reichen fuer eine UniFi-Network-
  Topologiekante, solange sie nicht mit der realen Controller-Sicht echter
  Switches kollidieren;
- die physische Bridge-/NIC-MAC des Hosts kann die dargestellte Topologie-
  Richtung umkehren, wenn der echte Upstream-Switch diese MAC bereits meldet.

## Packaging-Stand

Das Repository enthaelt Paketdefinitionen fuer:

- Linux-Service-Packaging mit systemd-/OpenRC-orientierter Konfiguration.
- FreeBSD-/OPNsense-orientierte Konfiguration und Tarball-Ausgabe.
- Non-root Linux-Service-Ausfuehrung mit dokumentierter Capability-Behandlung
  fuer den Adoption-SSH-Kompatibilitaetsport.

Paketierte Defaults sind lab-orientiert und muessen vor Nutzung in einem
geteilten Management-Netz weiterhin geprueft werden.

## Bekannte Produktgrenzen

- Gateway-Profile sind Identitaets- und Payload-Stubs, kein vollstaendiges
  Gateway-Verhalten.
- Externe Profile sind datengetrieben, muessen aber weiterhin gegen eine
  konkrete UniFi-Network-Version validiert werden.
- LLDP-Support ist passiv und haengt aktuell von `lldpd`-Ausgabe ab.
- LLDP ist fuer Adoption oder manuelle Topologie-Hinweise nicht erforderlich,
  aber ohne LLDP bleibt `uplink_neighbor` manuell und die Topologie-Richtung
  unterliegt weiterhin UniFi Networks eigenen Device-/MAC-Heuristiken.
- Linux `/proc` ergaenzt sysfs- und Bridge-Daten; es ersetzt keine
  vollstaendige OS-spezifische Interface-API.
- FreeBSD-Support ist bewusst konservativ und hat noch weniger Media-Details
  als spaetere native ioctl-/netlink-artige Adapter.
- Multi-Device-Simulation in einem Prozess ist noch nicht das Default-Design.

## Naechste Produktarbeit

Naheliegende Produktarbeit:

- Weitere gepinnte UniFi-Network-Versionen in der Docker-
  Kompatibilitaetsmatrix ergaenzen.
- Mehr Profil-Fixtures fuer Custom-Switch- und Custom-Gateway-Payloads.
- Strukturierterer Status fuer passives LLDP, Log-Reader und Plattform-
  Capabilities.
- First-class Topologie-Metadaten fuer Uplink-Nachbar-Remote-Port-Meldung,
  inklusive explizit konfiguriertem Remote-Port und LLDP-Fallback.
- Bessere Anleitung und Tools fuer die Wahl zwischen synthetischer und
  physischer Stub-MAC in Bridge-Observe-Deployments.
- Bessere FreeBSD-Interface-Media-/Counter-Details.
- CI-Coverage fuer Paketartefakte, SBOM und Dependency-Scanning.
- Release-Signaturen, sobald Package-CI stabil ist.

Firmware-Images, Captures, Adoption-Keys, Controller-Tokens, private
Controller-URLs, SSH-Host-Keys, MAC-Tabellen und Client-Daten bleiben aus Git.
