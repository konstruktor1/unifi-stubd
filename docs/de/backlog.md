# Produkt-Backlog

Stand: 2026-08-01.

Der Backlog ist die einzige Quelle fuer aktuelle Prioritaet und naechste
Arbeit. Er ist nach nutzbarem Produktergebnis geordnet; Runner, Fixtures,
Receipts und Testkataloge sind Bereitstellung, keine Backlog-Ergebnisse.

| Reihenfolge | Produktergebnis | Status | Fertig wenn |
| ---: | --- | --- | --- |
| 1 | Den Nicht-UniFi-`US8` adoptieren und seine echten Metriken zeigen. | Extern blockiert | Derselbe Kandidat bleibt verbunden und zeigt auf dynamisch aufgeloestem Stable und RC sichtbar echten Link, Speed und bidirektionalen Traffic der benannten Quelle. |
| 2 | Den Host ausserhalb der Controllerkontrolle halten. | Offen | Adoption behaelt nur Stub-Identitaet, und Controllerinput kann den Host nicht veraendern. |
| 3 | Die unterstuetzte Linux-Installation liefern. | Offen | Der abgenommene Stub wird auf einem unterstuetzten Linux-Host im Modus `bridge-observe` mit benannter echter Quelle installiert und gestartet. |
| 4 | Den Betreiberpfad veroeffentlichen. | Nach 1–3 | Der Quick Start deckt Installation, Konfiguration, Adoption, sichtbare Metriken, Grenzen und Recovery ohne Lab-Interna ab. |
| 5 | Eine weitere Repraesentation nur fuer einen konkreten Bedarf ergaenzen. | Spaeter | Vor der Umsetzung bestehen benanntes Betreiberergebnis, lokale read-only Quelle, Sicherheitsgrenze und eigene Stable-/RC-Abnahme. |

Aktueller Stand: Stable besitzt das vollstaendige betreibersichtbare Ergebnis.
Die aktuelle RC-UI liefert fuer ein benoetigtes Asset 404; fehlerhaftes
Stub-Verhalten wurde nicht beobachtet, und der strikte RC-Nachweis bleibt extern
blockiert. Siehe [Kompatibilitaet](compatibility.md).

## Naechster Schritt

Den unveraenderten Stable-/RC-Produktgate erneut ausfuehren, sobald Ubiquiti
einen neuen RC veroeffentlicht. Bis dahin den fokussierten Host-Safety-Nachweis
fuer Ergebnis 2 abschliessen. Das Controller-Artefakt nicht patchen, die
sichtbare Metrikanforderung nicht abschwaechen und den Stub ohne Nachweis eines
Produktfehlers nicht aendern. Die Installationsarbeit beginnt erst, wenn die
Ergebnisse 1 und 2 abgeschlossen sind.
