# Docker-Controller-Lab

Das Docker-Lab unter `lab/stub/` beantwortet eine Produktfrage: Kann das
aktuelle `unifi-stubd`-Binary eine Nicht-UniFi-Linux-Bridge als `US8`
darstellen, adoptiert werden, verbunden bleiben und Live-Metriken in UniFi
Network anzeigen?

Der Gate verwendet genau ein Produkt-Fixture: `stub-bridge-observe`. Die
uebrigen Services sind nur Bereitstellung: UniFi Network, MongoDB, ein
Controller-API-Helfer und ein Browser-Helfer. Payload-Dumps, lokale
Statusschemata, Gateway-Kommandos, synthetische Port-Maps und
Management-LAN-Varianten sind keine eigenen Produktgates.

## Abnahmekommando

Die Produktabnahme laeuft immer gegen das neueste veroeffentlichte Stable und
den neuesten veroeffentlichten Release Candidate:

```sh
BUILD_RUNNER_DIR=/path/to/build-runner make integration-nonunifi-metrics
```

Die Matrix fragt beide Versionen zu Beginn jedes Laufs beim
Ubiquiti-Release-Dienst ab; keiner der beiden Kanaele wird im Repository
festgeschrieben. Das Produkt wird einmal gebaut und in beiden wegwerfbaren
Controller-Spuren bytegleich geprueft.

Die Abnahme ist nur erfolgreich, wenn beide Spuren Folgendes belegen:

- der tatsaechliche Controller meldet die aufgeloeste Version;
- der US8 ist vor der Adoption sichtbar und akzeptiert die Controller-Adoption;
- beide Metrikabfragen melden den US8 adoptiert und verbunden;
- begrenzter bidirektionaler Bridge-Traffic laesst beide im Controller
  sichtbaren Root-Zaehler zwischen Abfragen ansteigen;
- der Controller-Readback meldet `uplink0` aktiv, den beobachteten Speed sowie
  positive aktuelle Uplink- und Port-Raten;
- die Geraeteansicht zeigt den US8 mit dem beobachteten Speed sowie positiven
  `Down`- und `Up`-Trafficwerten.

Das sind Beobachtungen an der Produktgrenze. Der Gate bildet die eigenen
JSON-Strukturen des Produkts bewusst nicht in Assertion-Helfern nach.

## Evidence und Cleanup

Die Stable-/RC-Matrix behaelt nicht committete Laufartefakte unter
`dist/controller-channel-matrix/`:

- aufgeloeste Stable-/RC-Versionen und Releasemetadaten;
- Controller-Readback-Logs fuer beide Spuren; und
- erst nach erfolgreichen Spuren und Cleanup das akzeptierte Binary sowie
  `accepted.env` mit Produktversion, OS, Architektur und Binary-SHA-256.

Ein fehlgeschlagener Lauf besitzt keinen Acceptance-Marker. Ein Claim gilt nur
fuer das akzeptierte Binary in genau diesem Laufverzeichnis. Eine spaetere
Produkt- oder Gate-Aenderung verlangt einen neuen Lauf.

Jede Spur startet mit leeren Controller- und Datenbank-Volumes. Bei Erfolg und
Fehler entfernt der Gate seine Container, Netzwerke, benannten Volumes,
Orphans und eigenen Image-Aliase. Es gibt keinen Keep-Running-Bypass und keine
festen Containernamen.

## Grenze

Das Lab belegt container-lokale Linux-Bridge-/FDB-Beobachtung und das daraus
entstehende controller-sichtbare Produktverhalten. Es belegt keine physische
Topologie, Proxmox-Host-Bridge, FreeBSD-/OPNsense-Runtime oder einen deployten
Controller. Diese bleiben separat besessene maschinengebundene
Kompatibilitaetstests.
