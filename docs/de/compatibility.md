# Controller-Kompatibilitaet

Kompatibilitaet wird gegen das zu Laufbeginn aufgeloeste neueste
veroeffentlichte UniFi-Network-Stable und den neuesten veroeffentlichten Release
Candidate bewertet. Beide Spuren muessen dieselben Produktbytes ausfuehren.
Erfasste Controllerversionen sind Evidence dieses Laufs, keine Produkt-Pins.

## Abgenommene Laeufe

Derzeit qualifiziert sich kein Lauf als dauerhafte Produktabnahme.

Im Repository liegt kein aufbewahrter ausfuehrbarer Nachweis, der einen exakten
Kandidaten an das erforderliche Stable-/RC-Produktergebnis und fokussierte
Host-Safety-Pruefungen bindet.

## Diagnostischer Kanalbefund

Am 2026-08-01 bestand derselbe Stub-Kandidat auf Stable 10.5.67 den
vollstaendigen sichtbaren Produktpfad. RC 10.5.66 akzeptierte, adoptierte und
verband ihn und empfing echte 10-GbE- sowie bidirektionale Trafficmetriken.
Seine UI forderte `swai.b54a...js` an; ein direkter Controllerabruf dieses
Pfads lieferte 404. Das RC-Ergebnis ist damit an der Controller-UI blockiert,
waehrend kein fehlerhaftes Stub-Verhalten beobachtet wurde. Der
Upstream-Kontext steht in der
[10.5.66-Releasediskussion](https://community.ui.com/releases/UniFi-Network-Application-10-5-66/9fd6a940-8f43-4071-a36a-28edcc7e5032)
und im [10.5.67-Release](https://community.ui.com/releases/UniFi-Network-Application-10-5-67/375288b9-a4b4-46f1-a19d-5c787d342c2b).

## Erforderlicher Laufnachweis

Eine abgenommene Zeile muss dauerhafte Evidence fuer Folgendes verlinken:

| Identitaet | Controllerergebnis | Produktergebnis |
| --- | --- | --- |
| Exakte Source- und Produktbinary-Identitaet | Aufgeloeste Stable- und RC-Versionen, in beiden mit denselben Bytes | Adoption, Connected-State, echte sichtbare Sourcemetriken und fokussierte Host-Safety-Pruefungen |

Die vollstaendigen Abnahmebedingungen gehoeren den
[Produktanforderungen](requirements.md). Diagnose- und abgeloeste Laeufe bleiben
Lab-Historie und erscheinen nicht als Kompatibilitaetsclaims.
Paketinstallation ist ein getrenntes Phase-2-Releaseergebnis.
