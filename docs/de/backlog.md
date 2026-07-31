# Produkt-Backlog

Stand: 2026-07-31.

Der Backlog ist die einzige Quelle fuer aktuelle Prioritaet und naechste
Arbeit. Er ist nach nutzbarem Produktergebnis geordnet; Runner, Fixtures,
Receipts und Testkataloge sind Bereitstellung, keine Backlog-Ergebnisse.

| Reihenfolge | Produktergebnis | Status | Fertig wenn |
| ---: | --- | --- | --- |
| 1 | Den Nicht-UniFi-`US8` mit echten Metriken abnehmen. | Offen | Derselbe Kandidat ist auf dynamisch aufgeloestem Stable und RC sichtbar, adoptierbar, verbunden und liefert steigende echte Link- und Trafficwerte. Dauerhafte Evidence bindet exakte Source und Binary. |
| 2 | Das betreibersichtbare Ergebnis belegen. | Offen | Beide Controller-Geraeteseiten zeigen Link-State, beobachteten Speed und bidirektionalen Traffic der benannten lokalen Quelle. API-Readback allein reicht nicht. |
| 3 | Den Host-Safety-Nachweis an den Kandidaten binden. | Offen | Fokussierte ausfuehrbare Pruefungen fuer begrenzten Adoption-State und verworfene Hostaktionen bestehen fuer die exakt abgenommene Source. |
| 4 | Die unterstuetzte Linux-Installation liefern. | Offen | Ein Paket mit diesem Binary wird auf einem unterstuetzten Linux-Host als `US8` im Modus `bridge-observe` mit benannter echter Quelle installiert und gestartet. |
| 5 | Den Betreiberpfad veroeffentlichen. | Nach 1–4 | Der Quick Start deckt Installation, Konfiguration, Adoption, sichtbare Metriken, Grenzen und Recovery ohne Lab-Interna ab. |
| 6 | Eine weitere Repraesentation nur fuer einen konkreten Bedarf ergaenzen. | Spaeter | Vor der Umsetzung bestehen benanntes Betreiberergebnis, lokale read-only Quelle, Sicherheitsgrenze und eigene Stable-/RC-Abnahme. |

## Naechster Schritt

Eine zusammenhaengende Kandidatenreise ausfuehren: Eine `US8`-Linux-Bridge mit
ihrer benannten echten Uplinkquelle konfigurieren, das Produkt einmal bauen,
genau dieses Binary ueber das Paket installieren und dieselben Bytes auf
neuestem Stable und neuestem Release Candidate verwenden. Ausfuehrbare
Datenpfad- und Safety-Ergebnisse aufbewahren und die sichtbaren Metriken in
beiden Controller-UIs direkt bestaetigen. Bis diese Reise besteht, ist kein
neues Profil, Gateway-Verhalten oder keine neue Metrikfamilie als Naechstes
dran.
