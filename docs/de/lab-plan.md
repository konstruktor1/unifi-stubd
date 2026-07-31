# Stub-Lab-Plan

## Umfang

Dieses Repository enthaelt nur Lab-Fixtures, die den `unifi-stubd`-Build
unmittelbar validieren. Firmware-Research, Gateway-Simulationen, UTM-/QEMU-VMs
und maschinengebundene Runtime-Werkzeuge liegen im getrennten Repository
`unifi-stubd-labs`.

Das Lab im Repository ist `lab/stub`. Es ist der Controller-Produktgate fuer
US8-Discovery, Adoption, verbundene Metrikabfragen und read-only
Bridge-Observation. Es haengt weder vom Schwesterprojekt noch von einem
externen Lab-Datenvolume ab.

## Sicherheitsgrenze

- Nur ein isoliertes Lab- oder Management-Netz verwenden.
- In versionierten Beispielen Dokumentationsadressen wie `192.0.2.0/24`
  verwenden.
- Controller-Tokens, Adoption-Keys, Captures, Zertifikate, echte MAC-Tabellen
  und private Controller-Adressen nie committen.
- Controller-Befehle duerfen keine beliebigen Shell-Kommandos, Upgrades,
  Restarts oder Host-Netzwerkaenderungen ausloesen.

## Validierungseinstiege

Repository-Checks und Packaging verwenden den konfigurierten Build-Runner-
Vertrag; dieses Dokument definiert keinen direkten Host-Aufruf von Go oder
Compose.

Die aktuelle Produktqualifikation beginnt mit
`make integration-nonunifi-metrics`. Sie recherchiert und loest zu
Matrixbeginn das neueste veroeffentlichte Stable und den neuesten
veroeffentlichten Release Candidate auf und prueft denselben Kandidaten in
beiden Spuren. Kann einer der beiden veroeffentlichten Kanaele nicht aufgeloest
werden, ist die Qualifikation blockiert; es gibt keinen Fallback auf eine feste
Version. QEMU-Maschinen im Schwesterprojekt sind getrennte
Kompatibilitaetspfade, kein alternativer Releaseeinstieg.

## Docker-Controller-Regressionsfixture

Das wiederholbare Regressionsfixture besteht aus:

```text
lab/stub/compose.yaml
lab/stub/compose.tests.yaml
lab/stub/mongo-init-unifi.sh
lab/stub/configs/
lab/stub/scripts/
lab/stub/tools/
```

Das gesamte Compose-Projekt wird atomar durch den Runner-Payload erzeugt und
entfernt. Container haben keine festen host-globalen Namen und keine
Restart-Policy; alle benannten Volumes und Netzwerke werden am Ende jedes
Laufs entfernt. Dieses Fixture liefert fokussierte Controller-
Regressionsevidence fuer das US8-Ergebnis aus Adoption, Verbindung und
Metriken.

## Exaktkandidaten-Releasequalifikation

Die Matrix bindet exakt dieselben Kandidatenbytes, denselben Source-Snapshot,
dasselbe Profil und dieselben Assertions an beide Controllerspuren. Pro Spur
erfasst sie:

- offiziellen Kanal, Release, aufgeloeste Version, Releasequelle und offizielle
  Artefakt-SHA;
- aufgeloesten unveraenderlichen Runtime-Digest und die tatsaechlich per API
  gemeldete Controllerversion;
- Kandidatenidentitaet und -SHA, controller-sichtbare Produkttelemetrie und
  vollstaendige Cleanup-Evidence.

Jede Spur muss das Produktergebnis belegen:

- Der Nicht-UniFi-Host oder die Bridge erscheint als konfiguriertes
  evidenzgebundenes Profil, wird adoptiert und bleibt verbunden.
- Controller-Readback zeigt erwartete Identitaet, Link-State und die von der
  Quelle beobachtete Speed des ausgewaehlten Uplinks.
- Beide kumulativen Zaehler steigen waehrend begrenztem bidirektionalem Traffic
  zwischen Controllerabfragen an; aktuelle RX-/TX-Raten sind positiv.
- Controller, Datenbank, Stub-State und laufspezifische Ressourcen
  werden vollstaendig bereinigt.

Eine kandidatgebundene Safety-Pruefung muss getrennt belegen, dass
Controllerinput den Host nicht provisionieren oder veraendern kann; das wird
nicht je Spur dupliziert. Diese aufbewahrte Kandidatenbindung ist derzeit noch
offen.

Die Releasematrix besteht nur, wenn sowohl die recherchierte Stable- als auch
die recherchierte Release-Candidate-Spur mit denselben Kandidatenbytes
bestehen. Sie belegt kein physisches Layer-2-Verhalten, keine realen
Proxmox-Bridges, kein FreeBSD-/OPNsense-Runtime-Verhalten, keine Firmware und
keine physischen UniFi-Geraete. Diese Themen bleiben getrennte
maschinengebundene Aufgaben in `unifi-stubd-labs`.
