# Docker-Controller-Lab

Das Docker-Lab unter `lab/stub/` beantwortet eine Produktfrage: Kann das
aktuelle `unifi-stubd`-Binary eine Nicht-UniFi-Linux-Bridge als `US8`
darstellen, adoptiert werden, verbunden bleiben und Live-Metriken in UniFi
Network anzeigen?

Der Gate verwendet genau ein Produkt-Fixture: `stub-bridge-observe`. Die
uebrigen Services sind nur Bereitstellung: UniFi Network, MongoDB, ein
Controller-API-Helfer und ein Browser-Helfer. Payload-Dumps, lokale
Statusschemata, Gateway-Kommandos, synthetische Port-Maps und
Management-LAN-Varianten sind keine eigenen Controller-Gates.

## Controller-Gate

Der controller-sichtbare Gate laeuft immer gegen das neueste veroeffentlichte Stable und
den neuesten veroeffentlichten Release Candidate:

```sh
make integration-nonunifi-metrics
```

Die Matrix fragt beide Versionen zu Beginn jedes Laufs beim
Ubiquiti-Release-Dienst ab; keiner der beiden Kanaele wird im Repository
festgeschrieben. Das Produkt wird einmal gebaut und in beiden wegwerfbaren
Controller-Spuren bytegleich geprueft.

Der Gate ist nur erfolgreich, wenn beide Spuren Folgendes belegen:

- der tatsaechliche Controller meldet die aufgeloeste Version;
- der US8 ist vor der Adoption sichtbar und akzeptiert die Controller-Adoption;
- waehrend begrenzter bidirektionaler Bridge-Traffic laeuft, zeigt die
  Geraeteansicht den beobachteten Speed sowie positive `Down`- und
  `Up`-Trafficwerte.

Das sind Beobachtungen an der Produktgrenze. Der Gate bildet die eigenen
JSON-Strukturen des Produkts bewusst nicht in Assertion-Helfern nach.
Die vollstaendige Produktabnahme verlangt zusaetzlich die getrennten
Host-Safety- und No-Fabrication-Belege aus den [Requirements](requirements.md).

## Evidence und Cleanup

Die Stable-/RC-Matrix behaelt nicht committete Laufartefakte unter
`dist/controller-channel-matrix/`:

- aufgeloeste Stable-/RC-Versionen und Releasemetadaten;
- Product-Gate-Logs fuer beide Spuren.

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
