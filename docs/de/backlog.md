# Produkt-Backlog

Stand: 2026-07-29.

Der Backlog ist nach nutzbarem Produktergebnis geordnet. Bereitstellungs- und
Validierungsmechanik ist keine eigene Produktarbeit.

| Reihenfolge | Ergebnis | Status | Fertig wenn |
| ---: | --- | --- | --- |
| 1 | Den Nicht-UniFi-`US8`-Datenpfad mit echten Metriken belegen. | Verifiziert | Die aktuelle Kompatibilitaetsevidence zeigt denselben Kandidaten auf aktuellem Stable und RC sichtbar, adoptiert, verbunden und mit steigenden echten Metrikwerten. |
| 2 | Belegen, dass ein Betreiber diese Metriken in UniFi Network sieht. | Als Naechstes | Die US8-Geraeteseite auf aktuellem Stable und RC zeigt sichtbar Link, beobachteten Speed und bidirektionale Trafficwerte der echten Quelle. API-Readback allein reicht nicht. |
| 3 | Die Host-Sicherheitsgrenze erhalten. | Implementiert; Evidence offen | Fokussierte Tests decken begrenzten Stub-eigenen Adoption-State und verworfene Shell-, Netzwerk-, Restart-, Upgrade- und Provisioning-Aktionen ab. Ein aufbewahrter Lauf muss diese Pruefungen an die exakt qualifizierte Kandidatensource binden. |
| 4 | Ein installierbares Linux-Paket fuer den unterstuetzten Schnitt liefern. | In Validierung | Das Paket enthaelt das fuer den Datenpfad qualifizierte Binary. Seine Startkonfiguration waehlt `US8` explizit; eine unterstuetzte Bridge-Konfiguration mit benannter echter Metrikquelle muss aber noch auf einem Linux-Host installiert und erfolgreich gestartet werden. |
| 5 | Den unterstuetzten Betreiberpfad nutzbar machen. | Nach 2 und 4 | Der Quick Start fuehrt von Installation und lokaler Metrikquelle ueber Adoption und sichtbare Controllermetriken bis zu Grenzen und Recovery ohne Lab-Interna. |
| 6 | Eine weitere Repraesentation nur fuer einen konkreten Produktbedarf ergaenzen. | Spaeter | Vor der Umsetzung bestehen benanntes Nutzerergebnis, lokale read-only Datenquelle, Sicherheitsgrenze und eigene Stable-/RC-Abnahme. |

## Naechster Schritt

Eintraege 2, 3 und 4 abschliessen. Zuerst die unterstuetzte US8-Bridge-
Konfiguration mit ihrer benannten echten lokalen Metrikquelle
vervollstaendigen; die explizite paketierte US8-Identitaet allein ist nicht
diese Konfiguration. Das laufgebundene Paket auf einem unterstuetzten
Linux-Host installieren, mit der unterstuetzten Konfiguration starten und in
beiden aktuellen Controllerkanaelen bestaetigen, dass ein Betreiber die echten
Metriken sieht. Ein fokussiertes Safety-Ergebnis an genau diese
Kandidatensource binden und aufbewahren. Genau dieser Pfad wird der Quick Start.
Keine neue Metrikfamilie, kein Gateway-Verhalten und kein weiteres Profil kommt
vor diesem nutzbaren Release.
