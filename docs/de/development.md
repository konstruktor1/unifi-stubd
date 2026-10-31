# Entwicklungsworkflow

Dieses Projekt nutzt einen gestuften GitHub Flow. Aenderungen bleiben klein und
reviewbar, `dev` ist der Integrationszweig, `main` ist die getestete
Release-Basis, und Pakete werden nur ueber Tags, GitHub-Releases oder einen
expliziten Package-Workflow-Lauf veroeffentlicht.

Die Workflow-Topologie trennt CI, Linux-Paketveroeffentlichung und die manuelle
FreeBSD-Erweiterung. Allgemeine Checks und Linux-Paketbau verwenden GitHub-hosted
Ubuntu-Worker mit kurzlebigen Build-Runner-Containern; auch der Install-Smoke
nutzt einen wegwerfbaren GitHub-hosted Worker. Der vorhandene Self-hosted Runner
ist ausschliesslich fuer manuelles FreeBSD-Packaging vorgesehen.
Providerquellen-Vorbereitung und offene Handoffs werden
in [CI/CD](ci-cd.md#offene-provider-handoffs) gepflegt.

## Go-Toolchain

`go.mod` und `go.work` deklarieren `go 1.27.0`, die aktuelle stabile Minorserie.
Das separate Linter-Modul verwendet dieselbe Untergrenze.
`golang.org/x/crypto v0.56.0` behebt die gemeldeten SSH-Schwachstellen.
Build Runner waehlt den ausfuehrenden Compiler
unabhaengig davon; Modul und Workspace haben keine `toolchain`-Direktive.

Von Agenten ausgefuehrte Go-Kommandos und Go-basierte Make-Targets verwenden
den konfigurierten Build Runner. Die vorbereitete Consumer-Runtime besitzt den
Aufruf; Host-Go ist auch in eingecheckten Workflows kein Fallback.

## Branch-Stufen

| Stufe | Branch | Zweck | Vorgesehene automatische Checks | Ergebnis |
| --- | --- | --- | --- | --- |
| Aenderung | kurzlebiger `codex/*`-, `feat/*`- oder `fix/*`-Repository-Branch | Eine fokussierte Aenderung | `CI / check` bei Branch-Push | Keine Paketveroeffentlichung |
| Integration | `dev` | Reviewte Arbeit vor Release-Promotion sammeln | `CI / check` bei Branch-Push | Keine Paketveroeffentlichung |
| Release-Basis | `main` | Stabile, getestete Quelle fuer Tags und oeffentliche Pakete | `CI / check` plus Package-Build und Debian-Install-Smoke bei Push | Paketartefakte im Workflow-Lauf |
| Paketquelle | `v*`-Tag, GitHub-Pre-Release oder manueller Workflow von `main` | Standard-Paketset veroeffentlichen | Linux-Package-Matrix und Release-SBOM | Linux-only-GitHub-Pages und bei einem Tag ein GitHub-Pre-Release |
| FreeBSD-Erweiterung | Manueller `FreeBSD Packaging`-Dispatch | Kompatibilitaetsartefakte bis zu einer genehmigten Buildroute | Containerisierter self-hosted Runner und direkter FreeBSD-Builder | Nicht beweiskraeftige Laufartefakte; Veroeffentlichung bleibt blockiert |

Kurzlebige Branches sollten normalerweise nur wenige Tage leben. Wenn eine
Aenderung zu gross wird, wird sie in kleinere Pull Requests geteilt, die jeweils
eigenstaendig durch die Gates kommen.

## Normale Entwicklung

1. Vom aktuellen `dev` starten.

   ```sh
   git switch dev
   git pull --ff-only origin dev
   git switch -c codex/<thema>
   ```

2. Die Aenderung umsetzen und den Patch selbst auf dem Host pruefen.

   ```sh
   git diff --check
   ```

3. Den vollstaendigen Gate `make verify` aus dem vorbereiteten Consumer und
   die erforderliche zusaetzliche Produktabnahme ausfuehren. Einen blockierten
   Check mit Owner und naechstem oeffentlichem Befehl festhalten; er erfuellt
   die Merge-Bedingung nicht. Den Topic-Branch pushen und einen Pull Request
   nach `dev` oeffnen.
4. Nach `dev` nur mergen, wenn die erforderlichen Checks und das Review passen.
5. `dev` gruen halten. Ein kaputter `dev` wird vor fremder Folgearbeit
   repariert.

## Promotion Nach Main

1. Pull Request von `dev` nach `main` oeffnen.
2. Der Pull Request muss `CI / check` mit beiden vorbereiteten Providerbindungen
   auf dem CI-Worker bestehen. Fehlende Runtime-Vorbereitung ist ein Fehler,
   kein Nachweis.
3. Fuer controller-sichtbare Aenderungen den standardisierten Docker-Gate aus
   [Docker-Controller-Lab](docker-lab.md) auf dem exakt promoteten `dev`-Commit
   ausfuehren und die Evidenz festhalten.
4. Den Diff als Release-Candidate-Aenderungsset reviewen, nicht als einzelne
   Feature-Aenderung.
5. Nach `main` mergen.
6. Nach Einrichtung der genehmigten Route startet der `main`-Push `CI / check`;
   danach baut der Package-Job die Linux-Formate und fuehrt den Debian-Install-
   Smoke aus.
7. Paketquellen nicht von `dev` deployen.

Direkte Pushes auf `main` bleiben expliziten Notfaellen oder Automation
vorbehalten. Der normale Weg ist Pull-Request-Review nach `main`.

## Release Und Paketveroeffentlichung

Die [Release-Checkliste](release.md) besitzt Tags, manuelle Workflow-Inputs,
die FreeBSD-Erweiterung, Pruefsummen, Signierung und Pages-Veroeffentlichung.
Diese Seite besitzt nur Branch-Promotion und Gate-Auswahl.

## Hotfixes

1. Branch von `main` erstellen.
2. Pull Request nach `main` oeffnen.
3. Dieselben `main`-Gates ausfuehren.
4. Erst taggen oder veroeffentlichen, wenn der exakte Tree den vollstaendigen
   Gate bestanden hat; der aktuelle `main`-CI-Lauf ist kein Ersatz.
5. Den Hotfix nach `dev` zurueckfuehren, bevor normale Entwicklung weitergeht.

## Gate-Auswahl

| Aenderungstyp | Erforderlicher Produktcheck | Zusaetzliche Produkt- oder Delivery-Evidenz |
| --- | --- | --- |
| Go-Code, Config-Schema, Profildaten | `make verify`, nach hostseitigem `git diff --check` | Zusaetzliche Produktabnahme je nach Aenderung |
| US8-Inform, -Adoption oder controller-sichtbare Metriken | `make verify` | Stable-/RC-Controller-Abnahme; zuerst den [Runtime-Handoff](ci-cd.md#offene-provider-handoffs) aufloesen |
| Paketierte Config, Service-Dateien, Paket-Metadaten | `make verify` auf dem Host; `make package` im Build Runner | `main`-Install-Smoke auf dem wegwerfbaren Worker |
| FreeBSD- oder OPNsense-Runtime-Verhalten | `make verify` | Autorisierte FreeBSD-Buildroute plus Tier-3-Runtime-Smoke; der direkte Host-Go-Builder ist nur Kompatibilitaetspfad und als Evidenz blockiert |
| Release Notes, Paketveroeffentlichung | `make verify` | Linux-first `Package Repositories`; FreeBSD-Veroeffentlichung ist nicht verfuegbar |

Zielhost-Paketinstallation ist kein Standard-Gate der Entwicklung. Sie wird nur
fuer explizite Rollout-Tests genutzt; host-spezifische Configs bleiben ausserhalb
dieses Repositories.

## Empfohlene GitHub-Kontrollen

Diese Kontrollen in GitHub einrichten und pruefen; Repository-Dateien allein
belegen ihre aktiven Einstellungen nicht:

- `main`: Pull Requests verlangen, `CI / check` verlangen, Loeschen blockieren,
  Non-Fast-Forward-Updates blockieren und das Pages-Environment auf `main` und
  `v*`-Tags begrenzen.
- `dev`: Pull Requests und `CI / check` verlangen, sobald der Branch als
  Integrationsstufe genutzt wird.
- `github-pages`-Environment: Deployment nur von `main` und `v*`.
- `freebsd-packaging`-Environment: Freigabe verlangen, nur `main` und `v*`
  erlauben und dem self-hosted Build-Job keine Publish-Rechte geben.
- Runner-Rollen und FreeBSD-Proxy-Safety werden von den Workflows und der
  [Release-Checkliste](release.md) durchgesetzt; sie werden hier nicht
  dupliziert.

## Quellen

- GitHub Flow: https://docs.github.com/en/get-started/using-github/github-flow
- GitHub-Actions-Events: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
- Protected Branches und Required Checks: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- Deployment-Environments: https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments
- Kurzlebige Branches in Trunk-Based Development: https://trunkbaseddevelopment.com/short-lived-feature-branches/
