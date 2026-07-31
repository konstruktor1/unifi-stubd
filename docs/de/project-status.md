# Projektstand

Zuletzt aktualisiert: 2026-07-29.

## Produktergebnis

`unifi-stubd` soll einem Betreiber erlauben, einen Nicht-UniFi-Linux-Host oder
eine Bridge als kleinen Switch in UniFi Network einzufuegen und echte
read-only Link- und Trafficmetriken einer benannten lokalen Schnittstelle zu
sehen.

Der erste unterstuetzte Schnitt ist bewusst eng: eine `US8`-Repraesentation auf
einer lokalen Linux-Bridge. Gateway-Identitaeten, andere Switchmodelle,
FreeBSD und weitere physische Topologien sind teilweise implementiert oder
erforscht, aber ohne eigene Produktentscheidung und Evidence nicht
abgenommener Produktscope.

## Evidenzstand

| Produktbedingung | Status | Evidence oder Luecke |
| --- | --- | --- |
| Vor Adoption sichtbar, adoptierbar und verbunden | Verifiziert | Derselbe Kandidat bestand aktuelles Stable und Release Candidate. |
| Echter Link, beobachteter Speed, Zaehler und Raten erreichen den Controller | Verifiziert | Controller-Datenabfragen zeigten die benannte Quelle, positive Raten und steigende Zaehler in beiden Kanaelen. |
| Betreiber sieht diese Werte auf der US8-Geraeteseite | Offen | Es wurde noch keine direkte UI-Beobachtung aufbewahrt. API-Readback belegt die Darstellung nicht. |
| Adoption kann den Host weder provisionieren noch Aktionen ausfuehren | Evidence offen | Fokussierte Tests decken begrenzte Persistenz und verworfene Shell-, Netzwerk-, Restart-, Upgrade- und Provisioning-Aktionen ab, aber die aktuelle Produktevidence bindet keinen Safety-Lauf an die exakt qualifizierte Source. |
| Linux-Paket enthaelt das qualifizierte Binary | Verifiziert | Der laufgebundene Paketsatz entsteht aus dem exportierten Kandidaten statt aus einem zweiten Build. |
| Installiertes Paket startet in der unterstuetzten US8-/Bridge-/Metrik-Konfiguration | Offen | Die paketierte Startkonfiguration waehlt `US8` nun explizit, bleibt aber im nicht beobachtenden `stub`-Modus ohne benannte echte Metrikquelle. |
| Betreiber-Quick-Start erreicht das sichtbare Metrikergebnis | Offen | Installation, unterstuetzte Konfiguration, UI-Bestaetigung, Grenzen und Recovery brauchen noch einen zusammenhaengenden Pfad. |

Exakte Lauf-IDs, Controllerversionen und Binary-Identitaet gehoeren in die
[Kompatibilitaetsevidence](compatibility.md), nicht in diese Zusammenfassung.

## Produktgrenze

Das unterstuetzte Verhalten bleibt read-only gegenueber dem dargestellten
Host. Controllerinput darf den begrenzten Adoption-State des Stubs herstellen,
aber keine beliebigen Shell-Kommandos ausfuehren und weder Host-Netzwerk,
Dienste, Firmware, Routen, Firewall, User noch Zeit veraendern.

Der Stub stellt nur Werte bereit, die an der konfigurierten lokalen Quelle
beobachtet wurden. Fehlende Beobachtungen bleiben fehlend und werden nicht zu
synthetischen Erfolgswerten. Discovery- und Inform-Traffic bleiben opt-in und
gehoeren nur in isolierte Lab- oder Management-Netze.

## Naechste Produktarbeit

1. Die unterstuetzte US8-Bridge-Konfiguration mit einer benannten echten
   lokalen Metrikquelle vervollstaendigen; die explizite paketierte
   US8-Identitaet allein ist keine Abnahmeevidence.
2. Das laufgebundene Paket auf einem unterstuetzten Linux-Host installieren und
   mit dieser Konfiguration starten.
3. Denselben Kandidaten auf dynamisch aufgeloestem Stable und Release Candidate
   adoptieren und direkt bestaetigen, dass ein Betreiber echte Link-, Speed-
   und bidirektionale Trafficwerte sieht.
4. Das fokussierte Host-Safety-Ergebnis fuer genau diese Kandidatensource
   aufbewahren.
5. Genau diesen Installations-, Konfigurations-, Adoptions-, Grenzen- und
   Recovery-Pfad zum Quick Start machen.

Erst nach dieser Abnahme ist der erste nutzbare Release abgeschlossen. Neue
Metrikfamilien, Gateway-Verhalten und weitere Profile bleiben spaetere
Produktentscheidungen.

## Evidence-Hygiene

Firmware-Images, Captures, Adoption-Keys, Controller-Tokens, private
Controller-URLs, SSH-Host-Keys, MAC-Tabellen und Client-Daten bleiben aus Git.
Labs, Runner, Pakete und Tests unterstuetzen die Auslieferung; sie sind keine
eigenen Produktfeatures oder Roadmap-Phasen.
