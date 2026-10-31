# Produktbacklog

Aktualisiert: 2026-10-31. Diese Seite besitzt Produktziele und Abhaengigkeiten.
Aktuelle Provider-Integration, CI und Laborpflege stehen im
[Engineering-Backlog](../engineering/backlog.md).

| ID | Ziel | Aktueller Status und Abschlusskriterium |
| --- | --- | --- |
| STUB-01 | Einen reproduzierbaren US8/Linux-Kandidaten festlegen. | Der vollstaendige Projektgate und Linux-Paketbau bestehen auf `main`. Das naechste Controller-Testbinary mit exakter Source und Eingaben binden; ein frueherer Kandidat qualifiziert kein neues Binary. |
| STUB-02 | Den controllerbezogenen Labortestpfad pflegen. | Modernisierung laeuft. Core besitzt die Produkttest-Orchestrierung; Labs besitzt Controller-Vorbereitung und verbindliche Maschinendaten. Zugeordnete Maschinen und aktuelle native Schnittstellen verwenden. Siehe [Controller-Labor](docker-lab.md). |
| STUB-03 | Sicherheitsgrenzen und Metrikfehlerfaelle nachweisen. | Offen. Begrenzte Adoption-Persistenz, lesende Beobachtung, fehlende Quellen, Link-Verlust, erste Samples und Zaehlerresets pruefen. Controllerbefehle fuer Neustart, Upgrade, Shell und Provisioning duerfen den Host nicht veraendern. |
| STUB-04 | Denselben US8-Kandidaten auf aktuellem Stable und RC abnehmen. | Nach STUB-01–03. Beide Kanaele zu Laufbeginn aufloesen; Adoption, Verbindung und sichtbare reale Geschwindigkeit sowie bidirektionalen Traffic der benannten Quelle nachweisen. API-Werte allein belegen keine sichtbare Abnahme. |
| STUB-05 | Die Installation auf einem unterstuetzten Linux-Host abnehmen. | Nach STUB-03–04. CI baut alle vier Linux-Formate und besteht einen echten Debian-Installations-, Versions-, Konfigurations- und Service-Datei-Test. Zur Produktabnahme gehoert noch der installierte Dienst mit benannter Bridge und realem Uplink. |
| STUB-06 | Den englischen und deutschen Betreiberpfad abschliessen. | Nach STUB-05. Installation, Konfiguration, Adoption, sichtbare Metriken, Grenzen und Wiederherstellung gegen den abgenommenen Release pruefen. |
| STUB-07 | Verifizierte Linux-Releases liefern. | Main-CI, Paketbau und Debian-Installationstest bestehen. Veroeffentlichung benoetigt das abgenommene Produkt mit passenden Source-/Binary-Nachweisen. |
| STUB-08 | Eine weitere Repraesentation fuer einen konkreten Bedarf hinzufuegen. | Bis zur Linux-Abnahme zurueckgestellt. Betreiberziel, lesende Quelle, Sicherheitsgrenze und eigene Stable-/RC-Kriterien benennen. FreeBSD/OPNsense bleibt ein separates optionales Ziel. |

Als naechster Produktnachweis folgt STUB-03, danach die gemeinsame Stable-/RC-
Abnahme. Historische Stable-11.0.97-Laborergebnisse gelten fuer ihren erfassten
Kandidaten und ihre Controller-Instanz; RC bleibt ungeprueft. Sie belegen keine
aktuelle gemeinsame Abnahme. Der verbindliche Vertrag und die Anforderungen an
dauerhafte Ergebnisse stehen in [Anforderungen](requirements.md) und
[Kompatibilitaet](compatibility.md).

Aktuelle native BR-/CT-Bindungen fuer die Ausfuehrung verwenden. Historische
Provider-Revisionen, entfernte Snapshots und fruehere Runner-Bereitschaft sind
keine Ausfuehrungsselektoren.
