# Entwicklungsworkflow

Dieses Projekt nutzt einen gestuften GitHub Flow. Aenderungen bleiben klein und
reviewbar, `dev` ist der Integrationszweig, `main` ist die getestete
Release-Basis, und Pakete werden nur ueber Tags, GitHub-Releases oder einen
expliziten Package-Workflow-Lauf veroeffentlicht.

Das Modell passt zur vorhandenen Repository-Infrastruktur:

- GitHub Actions `CI` fuehrt `make check`, `make vulncheck`,
  SBOM-Erzeugung und FreeBSD-Cross-Build-Checks aus.
- Bei Pushes auf `main` baut `CI` zusaetzlich Pakete und installiert das
  erzeugte Debian-Paket einmal im Ubuntu-Runner als neutralen Smoke-Test.
- `Package Repositories` baut Linux-Pakete, FreeBSD-Tarballs,
  Repository-Metadaten und native FreeBSD-pkg-Repositories und deployed
  GitHub Pages.
- Das `github-pages`-Environment erlaubt Deployments bereits nur von `main` und
  `v*`-Tags.

## Branch-Stufen

| Stufe | Branch | Zweck | Automatische Checks | Ergebnis |
| --- | --- | --- | --- | --- |
| Aenderung | kurzlebiger `codex/*`-, `feat/*`- oder `fix/*`-Branch | Eine fokussierte Aenderung | `CI / check` im Pull Request | Keine Paketveroeffentlichung |
| Integration | `dev` | Reviewte Arbeit vor Release-Promotion sammeln | `CI / check` im Pull Request und bei Push | Keine Paketveroeffentlichung |
| Release-Basis | `main` | Stabile, getestete Quelle fuer Tags und oeffentliche Pakete | `CI / check` plus Package-Build und Debian-Install-Smoke bei Push | Paketartefakte im Workflow-Lauf |
| Paketquelle | `v*`-Tag, GitHub-Pre-Release oder manueller Workflow von `main` | Veroeffentlichtes Paketquellen-Set | Package-Matrix plus FreeBSD-pkg-Repo-Build | GitHub-Pages-Paketquellen |

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

2. Die Aenderung lokal umsetzen und den passenden Gate ausfuehren.

   ```sh
   make check
   git diff --check
   ```

3. Den Topic-Branch pushen und einen Pull Request nach `dev` oeffnen.
4. Nach `dev` nur mergen, wenn die erforderlichen Checks und das Review passen.
5. `dev` gruen halten. Ein kaputter `dev` wird vor fremder Folgearbeit
   repariert.

## Promotion Nach Main

1. Pull Request von `dev` nach `main` oeffnen.
2. Der Pull Request muss `CI / check` bestehen.
3. Fuer controller-sichtbare Aenderungen den standardisierten Docker-Gate aus
   [Docker-Controller-Lab](docker-lab.md) auf dem exakt promoteten `dev`-Commit
   ausfuehren und die Evidenz festhalten.
4. Den Diff als Release-Candidate-Aenderungsset reviewen, nicht als einzelne
   Feature-Aenderung.
5. Nach `main` mergen.
6. Der `main`-Push startet `CI / check`; danach baut der Package-Job alle
   Paketformate und installiert das erzeugte Debian-Paket einmal in GitHub
   Actions.
7. Paketquellen nicht von `dev` deployen.

Direkte Pushes auf `main` bleiben expliziten Notfaellen oder Automation
vorbehalten. Der normale Weg ist Pull-Request-Review nach `main`.

## Release Und Paketveroeffentlichung

Versionsnummern kommen aus `v*`-Tags oder einem expliziten Package-Workflow-
Input. Fuer Release-Builds nicht auf den Makefile-Default vertrauen.

Der normale Alpha-Release-Pfad ist:

```sh
git switch main
git pull --ff-only origin main
git tag -a v0.2.0-alpha -m "unifi-stubd v0.2.0-alpha"
git push origin v0.2.0-alpha
```

Ein `v*`-Tag oder GitHub-Pre-Release startet `Package Repositories`. Manuelle
Retries laufen von `main`:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.0-alpha \
  -f package_release=1
```

Wenn `version` bei einem manuellen Lauf fehlt, ermittelt der Workflow den
neuesten erreichbaren `v[0-9]*`-Tag und entfernt das fuehrende `v`.

## Hotfixes

1. Branch von `main` erstellen.
2. Pull Request nach `main` oeffnen.
3. Dieselben `main`-Gates ausfuehren.
4. Erst taggen oder veroeffentlichen, wenn der `main`-CI-Lauf gruen ist.
5. Den Hotfix nach `dev` zurueckfuehren, bevor normale Entwicklung weitergeht.

## Gate-Auswahl

| Aenderungstyp | Erforderlicher lokaler Gate | Zusaetzlicher Gate |
| --- | --- | --- |
| Go-Code, Config-Schema, Profildaten | `make check`, `git diff --check` | Gezieltes `go test ./tests/...`, wenn sinnvoll |
| Inform, Adoption, Controller-Payload, Profil-Rendering | `make check` | Standardisierter Docker-Gate mit `make integration-docker` vor `dev` nach `main` |
| Paketierte Config, Service-Dateien, Paket-Metadaten | `make check`, `make package` | GitHub-`main`-Package-Install-Smoke |
| FreeBSD- oder OPNsense-Runtime-Verhalten | `make check` | FreeBSD-/OPNsense-Smoke nur mit temporaerem State |
| Release Notes, Paketveroeffentlichung | `make check` | `Package Repositories` per Tag, Release oder `main`-Dispatch |

Zielhost-Paketinstallation ist kein Standard-Gate der Entwicklung. Sie wird nur
fuer explizite Rollout-Tests genutzt; host-spezifische Configs bleiben ausserhalb
dieses Repositories.

## Empfohlene GitHub-Kontrollen

Das Repository nutzt aktuell ein Ruleset fuer den Default-Branch, das Loeschen
und Non-Fast-Forward-Updates blockiert. Das bleibt sinnvoll; als Ziel-Policy
gelten:

- `main`: Pull Requests verlangen, `CI / check` verlangen, Loeschen blockieren,
  Non-Fast-Forward-Updates blockieren und das Pages-Environment auf `main` und
  `v*`-Tags begrenzen.
- `dev`: Pull Requests und `CI / check` verlangen, sobald der Branch als
  Integrationsstufe genutzt wird.
- `github-pages`-Environment: Deployment nur von `main` und `v*`.

Der Package-Job laeuft absichtlich nach dem Merge auf `main`, weil er die exakt
committete Release-Basis prueft und Paketartefakte fuer diesen Lauf hochlaedt.

## Quellen

- GitHub Flow: https://docs.github.com/en/get-started/using-github/github-flow
- GitHub-Actions-Events: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
- Protected Branches und Required Checks: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- Deployment-Environments: https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments
- Kurzlebige Branches in Trunk-Based Development: https://trunkbaseddevelopment.com/short-lived-feature-branches/
