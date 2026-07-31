# Produkt-Roadmap

Stand: 2026-07-29.

Die Roadmap beschreibt das Stub-Produkt. Labs, Runner, Pakete und Tests
unterstuetzen die Auslieferung; sie sind keine Produktphasen.

## 1. US8 mit sichtbaren echten Metriken — aktuell

Eine Nicht-UniFi-Linux-Bridge ausliefern, die ein Betreiber in UniFi Network
als `US8` hinzufuegen kann.

Produktergebnis:

- sie wird gefunden und kann adoptiert werden;
- sie bleibt verbunden;
- sie meldet echten Link-State, Geschwindigkeit, Counter und Raten aus der
  benannten lokalen Quelle;
- Controllerinput kann den Host nicht provisionieren.

Die aktuelle [Kompatibilitaetsevidence](compatibility.md) verifiziert Discovery,
Adoption, Verbindung und den echten Metrikdatenpfad auf dynamisch aufgeloestem
Stable und Release Candidate. Fokussierte Tests decken die
Host-Sicherheitsgrenze ab, aber ein aufbewahrtes Ergebnis ist noch nicht an die
exakt qualifizierte Source gebunden. Auch direkte Evidence, dass ein Betreiber
diese Werte auf beiden Controller-Geraeteseiten sieht, fehlt; diese
Produktphase ist daher nicht abgenommen.

## 2. Nutzbarer Linux-Release — danach

Dieses Verhalten in ein Paket und einen Betreiberpfad ueberfuehren,
die auf einem unterstuetzten Linux-Host funktionieren.

Exit:

- das Paket enthaelt das fuer den Datenpfad qualifizierte Produktbinary;
- die unterstuetzte Paketkonfiguration waehlt US8-Bridge-Observation und eine
  echte lokale Metrikquelle;
- Installation und paketierter Servicestart bestehen auf einem unterstuetzten
  Linux-Host;
- der Quick Start fuehrt einen Betreiber von der lokalen Metrikquelle bis zu
  Adoption und Controller-Readback;
- die Grenzen nennen US8 auf einer Linux-Bridge als unterstuetzten Schnitt.

Der binary-gebundene Paketsatz existiert. Seine Startkonfiguration waehlt nun
explizit `US8`, nutzt aber weiterhin den nicht beobachtenden `stub`-Modus ohne
benannte echte Metrikquelle und ist daher nicht die unterstuetzte
Abnahmekonfiguration. Hostspezifische Bridge-Konfiguration, installierter
Dienst und Betreiberpfad sind noch offen.

## 3. Weitere Repraesentationen — nur bei Bedarf

Ein weiteres Profil oder eine Plattform kommt nur mit einem konkreten
Betreiberergebnis, einer benannten lokalen read-only Quelle, einer expliziten
Sicherheitsentscheidung und eigener Stable-/RC-Abnahme hinzu. Evidence eines
benachbarten Modells erweitert das Produkt nicht automatisch.

## Unmittelbare Reihenfolge

1. Die unterstuetzte US8-Bridge-Konfiguration mit einer benannten echten
   lokalen Metrikquelle vervollstaendigen; die explizite paketierte
   US8-Identitaet allein ist keine Abnahmeevidence.
2. Das laufgebundene Paket auf einem unterstuetzten Linux-Host installieren und
   mit dieser Konfiguration starten.
3. Auf aktuellem Stable und Release Candidate adoptieren und direkt
   bestaetigen, dass ein Betreiber echte Link-, Speed- und Trafficwerte sieht.
4. Das fokussierte Host-Safety-Ergebnis fuer genau diese Kandidatensource
   aufbewahren.
5. Genau diesen Installations-, Adoptions-, Metriksource-, Grenzen- und
   Recovery-Pfad vor der Paketveroeffentlichung zum Betreiber-Quick-Start
   machen.

Neue Gateway-Telemetrie, WAN-Health-Semantik und weitere Profile bleiben
spaetere Optionen; sie kommen nicht vor dem nutzbaren Releasepfad.
