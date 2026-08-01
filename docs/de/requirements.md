# Produktanforderungen

Stand: 2026-08-01.

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
| `PR-METRICS` | UniFi Network zeigt echten Link-State, beobachteten Speed und bidirektionalen Traffic der konfigurierten lokalen Quelle. Der Stub liest diese Quelle, ohne den Host zu veraendern, und erfindet keine fehlenden Werte. |
| `PR-ADOPTION` | Adoption persistiert nur den eng begrenzten Stub-State. Controllerinput darf den Host weder provisionieren noch anderweitig veraendern. |
| `PR-COMPAT` | Derselbe Produktkandidat funktioniert mit dem neuesten veroeffentlichten Stable und dem neuesten veroeffentlichten Release Candidate; keine Controllerversion wird dauerhaft festgeschrieben. |
| `PR-SCOPE` | Der abgenommene Schnitt ist `US8` auf einer lokalen Linux-Bridge. Andere Profile, Gateways, Betriebssysteme und physische Topologien brauchen eine eigene Produktentscheidung und Evidence. |

## Abnahmebedingungen

Der Produktkandidat ist erst angenommen, wenn derselbe Kandidat auf beiden
aktuellen Controllerkanaelen Folgendes zeigt:

1. Der US8 ist sichtbar, kann adoptiert werden und bleibt verbunden.
2. Ein Betreiber sieht echten Link-State, beobachteten Speed und
   bidirektionalen Traffic der benannten lokalen Quelle.
3. Fehlende Beobachtungen bleiben abwesend, statt durch synthetische Werte
   ersetzt zu werden.
4. Controllerinput kann den Host weder provisionieren noch neu starten,
   upgraden oder Kommandos auf ihm ausfuehren.

Das betreibersichtbare Ergebnis ist erforderlich; controllerinterne Daten
allein sind keine Abnahme. Die genaue Pixelposition ist keine
Produktanforderung. Packaging und Installed-Service-Abnahme gehoeren zum
nutzbaren Linux-Release.
