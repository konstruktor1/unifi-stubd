# Produktanforderungen

Stand: 2026-07-31.

## Produktziel

`unifi-stubd` laesst einen Nicht-UniFi-Linux-Host oder eine Bridge in UniFi
Network als kleinen Switch erscheinen. Der erste unterstuetzte Produktschnitt
ist ein `US8`: Ein Betreiber kann ihn im Controller hinzufuegen, adoptieren,
verbunden halten und echte Link- und Trafficmetriken der benannten lokalen
Schnittstelle sehen.

Lab-Infrastruktur, Runner, Fixtures, interne Statusschemata und Testkataloge
sind Bereitstellung. Sie sind keine Produktanforderungen.

## Vertrag

| ID | Anforderung |
| --- | --- |
| `PR-STUB` | Eine lokale Linux-Bridge mit laufendem Stub erscheint als `US8`, kann adoptiert werden und bleibt verbunden. |
| `PR-METRICS` | Der Stub sendet echten read-only Link-State, Geschwindigkeit, kumulative RX-/TX-Zaehler und aktuelle bidirektionale Raten der konfigurierten lokalen Quelle an den Controller. UniFi Network muss einem Betreiber die resultierenden Link-, Speed- und Trafficwerte darstellen; rohe kumulative Zaehler muessen nicht wortwoertlich gerendert werden. Fehlende Daten bleiben abwesend und werden nie synthetisch ersetzt. |
| `PR-ADOPTION` | Adoption persistiert nur den eng begrenzten Stub-State. Controllerinput darf den Host weder provisionieren noch anderweitig veraendern. |
| `PR-COMPAT` | Die Produktabnahme loest zu Laufbeginn das neueste veroeffentlichte Stable und den neuesten veroeffentlichten Release Candidate auf. Dieselben Produktbytes muessen beide bestehen; keine Controllerversion wird dauerhaft festgeschrieben. |
| `PR-SCOPE` | Der abgenommene Schnitt ist `US8` auf einer lokalen Linux-Bridge. Andere Profile, Gateways, Betriebssysteme und physische Topologien brauchen eine eigene Produktentscheidung und Evidence. |

## Abnahmebedingungen

Der Produktkandidat ist erst angenommen, wenn beide aktuellen Controllerkanaele
fuer dasselbe Binary Folgendes zeigen:

1. Der Controller meldet die dynamisch aufgeloeste Version.
2. Der US8 ist vor der Adoption sichtbar und akzeptiert die Controller-Adoption.
3. Beide Controller-Metrikabfragen melden den US8 adoptiert und verbunden.
4. Der ausgewaehlte Uplink nennt die lokale Quelle, ist aktiv und meldet deren
   beobachtete Geschwindigkeit.
5. Beide Root-Zaehler steigen waehrend begrenztem bidirektionalem Traffic
   zwischen Controllerabfragen an; aktuelle Uplink-/Port-Raten sind positiv.
6. In beiden Controllerkanaelen kann ein Betreiber die resultierenden Link-,
   Speed- und Trafficwerte auf der US8-Geraeteseite sehen. Die Beobachtung muss
   echte Werte von leerem Feld, Platzhalter oder veraltetem synthetischem Wert
   unterscheiden.

Einmal pro Kandidat muessen fokussierte ausfuehrbare Pruefungen zusaetzlich
belegen, dass Adoption nur die begrenzte Stub-Identitaet persistiert und
Controllerinput keine Host-Provisionierung, Shell-, Restart- oder
Upgrade-Aktion ausloesen kann. Das sind Kandidatenbedingungen, keine
doppelten Assertions in jeder Controllerspur.

Controller-API-Readback belegt den Datenpfad, aber nicht seine Darstellung.
Eine direkte betreibersichtbare Beobachtung belegt die Darstellung; beides ist
erforderlich. Die genaue Pixelposition ist keine Produktanforderung.

## Abnahmeevidence

Vier Ergebnisse haben verschiedene Bedeutungen:

- Die Controller-Datenpfadevidence belegt Sichtbarkeit, Adoption, Verbindung
  und echte Metrikwerte auf aktuellem Stable und Release Candidate.
- UI-Evidence belegt, dass ein Betreiber diese Werte tatsaechlich sieht.
- Fokussierte Sicherheitspruefungen belegen begrenzte Persistenz und das Fehlen
  controller-getriggerter Hostaktionen.
- Die Artefaktbindung belegt, dass ein Paket das angenommene Binary enthaelt.

Keines ersetzt die anderen. Insbesondere beweist ein erzeugtes Paket noch
nicht, dass der installierte Dienst und die Betreiberanleitung auf einem
unterstuetzten Host funktionieren.
