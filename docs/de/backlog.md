# Produkt-Backlog

Stand: 2026-08-01.

Der Backlog ist die einzige Quelle fuer aktuelle Prioritaet und naechste
Arbeit. Er ist nach nutzbarem Produktergebnis geordnet; Runner, Fixtures,
Receipts und Testkataloge sind Bereitstellung, keine Backlog-Ergebnisse.

| Reihenfolge | Produktergebnis | Status | Fertig wenn |
| ---: | --- | --- | --- |
| 1 | Den Nicht-UniFi-`US8` adoptieren und seine echten Metriken zeigen. | In Arbeit | Derselbe Kandidat bleibt verbunden und zeigt auf dynamisch aufgeloestem Stable und RC sichtbar echten Link, Speed und bidirektionalen Traffic der benannten Quelle. |
| 2 | Den Host ausserhalb der Controllerkontrolle halten. | Offen | Adoption behaelt nur Stub-Identitaet, und Controllerinput kann den Host nicht veraendern. |
| 3 | Die unterstuetzte Linux-Installation liefern. | Offen | Der abgenommene Stub wird auf einem unterstuetzten Linux-Host im Modus `bridge-observe` mit benannter echter Quelle installiert und gestartet. |
| 4 | Den Betreiberpfad veroeffentlichen. | Nach 1–3 | Der Quick Start deckt Installation, Konfiguration, Adoption, sichtbare Metriken, Grenzen und Recovery ohne Lab-Interna ab. |
| 5 | Eine weitere Repraesentation nur fuer einen konkreten Bedarf ergaenzen. | Spaeter | Vor der Umsetzung bestehen benanntes Betreiberergebnis, lokale read-only Quelle, Sicherheitsgrenze und eigene Stable-/RC-Abnahme. |

Aktueller Stand: Das Stable-Produktergebnis ist vollstaendig. Auf dem aktuellen
Release Candidate ist der Stub sichtbar, adoptiert und verbunden und liefert
echte Metriken; die Controller-Geraeteseite ist jedoch nicht verfuegbar. Der
betreibersichtbare RC-Nachweis bleibt deshalb offen.

## Naechster Schritt

Das fehlende betreibersichtbare Ergebnis auf dem aktuellen Release Candidate
abschliessen und das Stable-Ergebnis fuer denselben Produktkandidaten
bestaetigen. Akzeptiert, verbindet und meldet der Release Candidate den Stub
weiterhin, waehrend seine Geraeteseite nicht verfuegbar bleibt, wird dies als
RC-Kompatibilitaetsgrenze festgehalten. Der Stub wird ohne Nachweis eines
fehlerhaften Produktverhaltens nicht geaendert. Phase 2 beginnt erst nach
Abschluss von Phase 1.
