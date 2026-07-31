# Controller-Kompatibilitaet

Der aktuelle Produktkandidat ist fuer den Controller-Datenpfad des definierten
`US8`-Linux-Bridge-Schnitts qualifiziert. Lauf `matrix-20260729102907-1`
bestand dieses Gate am 29.07.2026. Direkte betreibersichtbare UI-Evidence ist
noch offen; dies ist daher noch keine volle Produktabnahme.

Jeder Kompatibilitaetsclaim loest zu Laufbeginn das neueste veroeffentlichte
Stable und den neuesten veroeffentlichten Release Candidate auf und nutzt in
beiden Spuren dieselben Produktbytes. Erfasste Versionen sind Laufevidence,
keine Produkt-Pins.

## Aktuelle Controller-Datenpfadevidence

| Erfasster Lauf | Controllerkanaele | Qualifizierte Produktidentitaet | Beobachtetes Produktergebnis |
| --- | --- | --- | --- |
| `matrix-20260729102907-1`, 29.07.2026 | Stable `10.5.67`, RC `10.5.66` | `0.2.5-alpha`, Linux/arm64; Image `sha256:e32bbf6f114f2d02d91c4cecc4d592f6708f7683cd5cf97776a312101792a9e5`; Binary `sha256:e6e6868577061e8fdd6fe52ccd5a5181d084a5de8f363acd66b321249713afa1` | Dasselbe `US8`-Produktimage war vor der Adoption sichtbar, wurde adoptiert, blieb verbunden und lieferte ueber Controller-Datenabfragen `uplink0` mit lokal beobachteten 10.000 Mb/s, positive RX-/TX-Raten fuer Uplink und Port sowie steigende Root-Zaehler in beiden Spuren. Cleanup bestand. UI-Ergebnis: nicht getestet. |

Der Lauf bindet das Produktsource-Manifest
`sha256:8f086331798ebe872c48def723d52b8d5ec325727ba5a6a5216d3c59969e4477`
und das Gate-Manifest
`sha256:7a644c182680eb2efe510c734f0e2eae6b02860c530e9d08236329902b314c94`.
Letzteres enthaelt 17 explizite ausgefuehrte Gate-Eingaben; Captures,
generierte Evidence und unbeteiligte lokale Dateien sind ausgeschlossen. Die
aufgeloeste Controller-Runtime-Basis ist
`lscr.io/linuxserver/unifi-network-application@sha256:cebacc22bd529909a85efd91003c20c1d77bab7d79850689e2c0da720677bf4c`;
die Stable- und RC-Controller-Image-IDs sind
`sha256:c926c226dd434cafb7ee932946afc0fa79069c6d9fa54f2ebbe289aac47e5084`
und `sha256:926afa0f0c5d06fca4f6fad77f18e766415462f757beab23504878e4dcfce45b`.
Er lief als `make integration-nonunifi-metrics` ueber Build Runner `6.9.1` am
Commit `0e52df6641a073a93fd225b23c6164cc75050a33`. Der Runnerjob nutzte Go
`1.26.5` auf Linux/arm64; das exportierte Produktbinary nennt als tatsaechlichen
Compiler Go `1.25.12`. Die aeussere Invocation wurde im vergaenglichen
Provider-Sessioncache als `a837a9247a466ddcaf2388e7c4346a71` beobachtet. Der
Job erhielt weder diese aeussere ID noch den Projektnamen; `final.env` traegt
daher fuer beide ehrlich `unknown` ein, statt eine dauerhafte Bindung zu
erfinden.

Das qualifizierte Binary wurde erst nach zwei erfolgreichen Spuren und Cleanup
exportiert. Der laufgebundene Linux-Paketsatz entstand aus dieser Datei und
nicht aus einem zweiten Build. Unabhaengiges Extrahieren bestaetigte denselben
Binary-SHA-256 fuer Rohdatei, TGZ und DEB. Beide extrahierten Paketconfigs
waehlen explizit `profile: us8`. Der vollstaendige DEB-/RPM-/Arch-/TGZ-Satz
wurde in Build-Runner-Session `a353f7d825e2405feb56ddf7ded528d4` atomar
veroeffentlicht; alle vier Checksummen, die TGZ-/DEB-Binaryidentitaet und beide
US8-Configs bestanden in Session `db226b891c0a0e7f984da5358d987914`. RPM und
Arch entstanden aus demselben SHA-geprueften Staging-Input, ihre enthaltenen
Binaries wurden in diesem Runner aber nicht einzeln extrahiert. Das belegt die
Artefaktherkunft; dargestellte UI-Werte und das Verhalten des installierten
Dienstes bleiben eigene Abnahmeschritte.

## Abgeloeste Evidence

Fruehere Laeufe vom Juli 2026 sind nur Diagnosehistorie. Ihnen fehlte
mindestens eine der heutigen Source-, Binary-, Toolchain-, Runtime-Digest- oder
dynamischen Kanalbindungen; sie begruenden daher keinen aktuellen
Kompatibilitaetsclaim.

Firmware-, Gateway-, physische Topologie- und Private-Controller-Research
gehoert ins Companion-Labprotokoll und erweitert diese Produktmatrix nicht. Ein
spaeterer Kompatibilitaetsclaim muss Datenpfad-Gate und direkte UI-Beobachtung
auf dynamischem Stable/RC fuer seine exakten Produktbytes wiederholen; die
Versionen oben bleiben Laufevidence und sind keine Pins.
