# Controller-Kompatibilitaet

Kompatibilitaet wird gegen das zu Laufbeginn aufgeloeste neueste
veroeffentlichte UniFi-Network-Stable und den neuesten veroeffentlichten Release
Candidate bewertet. Beide Spuren muessen dieselben Produktbytes ausfuehren.
Erfasste Controllerversionen sind Evidence dieses Laufs, keine Produkt-Pins.

## Abgenommene Laeufe

Derzeit qualifiziert sich kein Lauf als dauerhafte Produktabnahme.

Im Repository liegt kein aufbewahrter ausfuehrbarer Nachweis, der einen exakten
Kandidaten an das erforderliche Stable-/RC-Produktergebnis und fokussierte
Host-Safety-Pruefungen bindet. Der aktuelle Release Candidate akzeptiert und
verbindet den Stub und empfaengt dessen echte Metriken, seine Geraeteseite ist
jedoch nicht verfuegbar. Damit fehlt das vollstaendige betreibersichtbare
Ergebnis; fehlerhaftes Stub-Verhalten ist dadurch nicht belegt.

## Erforderlicher Laufnachweis

Eine abgenommene Zeile muss dauerhafte Evidence fuer Folgendes verlinken:

| Identitaet | Controllerergebnis | Produktergebnis |
| --- | --- | --- |
| Exakte Source- und Produktbinary-Identitaet | Aufgeloeste Stable- und RC-Versionen, in beiden mit denselben Bytes | Adoption, Connected-State, echte sichtbare Sourcemetriken und fokussierte Host-Safety-Pruefungen |

Die vollstaendigen Abnahmebedingungen gehoeren den
[Produktanforderungen](requirements.md). Diagnose- und abgeloeste Laeufe bleiben
Lab-Historie und erscheinen nicht als Kompatibilitaetsclaims.
Paketinstallation ist ein getrenntes Phase-2-Releaseergebnis.
