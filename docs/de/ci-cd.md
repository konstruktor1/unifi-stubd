# Build, Pruefungen und Auslieferung

`Makefile`, Skripte und Workflows sind die ausfuehrbaren Quellen der Wahrheit.
Diese Seite ordnet nur ihre stabilen Einstiege zu; sie spiegelt keine einzelnen
Jobs, Tool-Versionen, Runner-Labels oder Artefaktlisten.

## Consumer-Einstiege

Alle von Agenten ausgefuehrten Go-Arbeiten verwenden den konfigurierten Build
Runner gemaess [`AGENTS.md`](../../AGENTS.md). Host-Go oder ein Go-basiertes
Make-Target ist kein Fallback.

| Ergebnis | Einstieg |
| --- | --- |
| Vollstaendiger Pre-Merge-Gate | `make code-checker` |
| Runner-lokaler Produkt-Subgate | `make check` |
| Aktueller controller-sichtbarer Stable-/RC-Gate | `make integration-nonunifi-metrics` |
| Linux-Pakete | `make package` |
| FreeBSD-Paketsatz (nur Kompatibilitaet; als Evidenz blockiert) | `make package-freebsd-pkg-repos` |
| Statische Paketquellen | `make package-repos` |

`make code-checker` ist der einzige vollstaendige Repository-Gate. Er fuehrt
Produkt- und Security-Subgates in einer vorbereiteten Build-Runner-Instanz aus
und schliesst mit Code Checker ab; seine Komponenten sind keine getrennten
Verifikationskandidaten.

`make integration-nonunifi-metrics` ist der einzige controller-sichtbare Gate.
Er loest das neueste veroeffentlichte Stable und den neuesten Release Candidate
zu Laufbeginn auf und verwendet in beiden wegwerfbaren Spuren denselben
Kandidaten. Details und Bedingungen stehen im
[Docker-Controller-Lab](docker-lab.md). Die vollstaendige Produktabnahme
verlangt zusaetzlich die fokussierten Host-Safety- und No-Fabrication-Belege
aus den [Requirements](requirements.md).

## Verantwortung der Automatisierung

- [CI](../../.github/workflows/ci.yml) besitzt Repository-Pruefungen,
  Diagnostik und den Linux-Package-Smoke-Pfad des Main-Branches.
- [Controller Integration](../../.github/workflows/tier2-docker.yml) besitzt den
  manuell gestarteten Controller-Gate. Der Owner des Self-hosted Runners muss
  die Build-Runner-Bindung des ausgecheckten Consumers nach dem Checkout
  vorbereiten.
- [Package Repositories](../../.github/workflows/package-pages.yml) besitzt
  Linux-Releaseartefakte und die Paketseite.
- [FreeBSD Packaging](../../.github/workflows/freebsd-packaging.yml) besitzt den
  getrennt betriebenen FreeBSD-Builder und den optionalen Publikationspfad.

Die eingecheckten CI- und Package-Workflows enthalten noch Host-Go-Schritte und
liefern bis zu einer genehmigten CI-sichtbaren Build-Runner-Route weder
Validierungs- noch Release-Evidenz. Der direkte FreeBSD-Builder ist ein
getrennter External-Machine-Kompatibilitaetspfad, kein Build-Runner-Nachweis.

Releaseentscheidungen und Betreiberschritte stehen im
[Entwicklungsworkflow](development.md) und in der
[Release-Checkliste](release.md).
