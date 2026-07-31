# Controller-Kompatibilitaet

Kompatibilitaet wird gegen das zu Laufbeginn aufgeloeste neueste
veroeffentlichte UniFi-Network-Stable und den neuesten veroeffentlichten Release
Candidate bewertet. Beide Spuren muessen dieselben Produktbytes ausfuehren.
Erfasste Controllerversionen sind Evidence dieses Laufs, keine Produkt-Pins.

## Abgenommene Laeufe

Derzeit qualifiziert sich kein Lauf als dauerhafte Produktabnahme.

Im Repository liegt kein aufbewahrter ausfuehrbarer Nachweis, der einen exakten
Kandidaten an alle erforderlichen Controller-Datenpfad-, betreibersichtbaren UI-,
Host-Safety-, Paket- und Installed-Service-Ergebnisse bindet. Session-IDs,
generierte `KEY=value`-Receipts, nur in Prosa wiederholte Hashes und ein API-
Ergebnis ohne UI-Beobachtung belegen diese Ergebnisse nicht.

## Erforderlicher Laufnachweis

Eine abgenommene Zeile muss dauerhafte Evidence fuer Folgendes verlinken:

| Identitaet | Controllerergebnis | Produktergebnis |
| --- | --- | --- |
| Exakter Source-Commit und Digest des Produktbinaries | Aufgeloeste Stable- und RC-Versionen, in beiden mit denselben Bytes | Sichtbarkeit, Adoption, Connected-State, echte Sourcemetriken, direkte UI-Beobachtung, fokussierte Host-Safety-Pruefungen und Start des installierten Pakets |

Die vollstaendigen Abnahmebedingungen gehoeren den
[Produktanforderungen](requirements.md). Diagnose- und abgeloeste Laeufe bleiben
Lab-Historie und erscheinen nicht als Kompatibilitaetsclaims.
