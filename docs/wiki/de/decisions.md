# Entscheidungen

Diese Seite haelt Architekturentscheide sichtbar, waehrend sich das Projekt
weiterentwickelt. Sie ist ein kompakter Index; die Details stehen in der
normalen Doku und in Code-Kommentaren.

## Controller-Provisioning wird nicht angewandt

Entscheid: Controller-Provisioning-Daten werden nur geparst, zusammengefasst
oder persistiert, wenn sie fuer kuenftigen Inform-Traffic sicher sind. Sie
werden nicht auf den Host angewandt.

Grund: Das Projekt ist ein Lab-Stub, kein gemanagter Host-Agent. Unbekannte
Controller-Kommandos auf einem Linux- oder FreeBSD-Host auszufuehren wuerde die
Safety-Grenze verletzen.

## Profile sind Daten

Entscheid: Geraeteform gehoert in Profildaten, nicht in Modellnamen-Branches.

Grund: Externe Profile und kuenftige Kompatibilitaetsarbeit brauchen einen
vorhersehbaren datengetriebenen Pfad. Payload-Rendering soll `payload.kind`,
Ports, Medien, Rollen, Network-Groups und Profil-Defaults nutzen.

## YAML-Extends mergt vor typisiertem Decode

Entscheid: Externe Profilvererbung wird auf YAML-Mapping-Ebene aufgeloest und
danach genau einmal in das kanonische Modell decodiert.

Grund: Das erhaelt explizite Zero-Value-Overrides und vermeidet handgeschriebene
Feldkopier-Kaskaden.

## Bridge-Observe ist rollenbasiert

Entscheid: Bridge-FDB-Zeilen werden vor dem Payload-Mapping als bridge, uplink,
access, unknown oder ignored klassifiziert.

Grund: Eine Proxmox-Bridge enthaelt lokale Teilnehmer und Remote-Infrastruktur.
Der Upstream-Switch und dessen Clients duerfen nicht als direkte lokale Access-
Port-Clients gerendert werden.

## Synthetische MACs fuer Darstellungs-Tests bevorzugen

Entscheid: Darstellungs-Tests sollen eine lokal administrierte synthetische
Stub-MAC nutzen, ausser die Physical-MAC-Heuristik ist das Testziel.

Grund: UniFi Network kann bevorzugen, was ein echter Upstream-UniFi-Switch ueber
eine physische Host-MAC meldet, und dadurch die dargestellte Topologiekante
umdrehen.

## Plattform-Integrationen sind optional

Entscheid: LLDP, journalctl, syslog, procfs und D-Bus sind optionale read-only
Quellen hinter `internal/platform`.

Grund: Linux- und FreeBSD-Umgebungen variieren. Fehlende optionale Tools sollen
im Status sichtbar sein, aber nicht installiert oder standardmaessig fatal sein.

## Management LAN erzeugt keine VLANs

Entscheid: Management-LAN-Unterstuetzung ist in dieser Version metadata-only
oder an ein preexisting Interface gebunden.

Grund: Aktiver VLAN-Lifecycle ist Host-Mutation und braucht ein separates
Review-Design.

## Tests liegen unter `tests/`

Entscheid: Go-Testdateien liegen unter `tests/`, nicht neben internen Packages.

Grund: Das ist eine durch `make check` erzwungene Projekt-Policy und haelt
Produktionspackage-Verzeichnisse auf Runtime-Code fokussiert.

