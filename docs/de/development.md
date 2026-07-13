# Entwicklungsworkflow

Dieses Projekt nutzt einen gestuften GitHub Flow. Aenderungen bleiben klein und
reviewbar, `dev` ist der Integrationszweig, `main` ist die getestete
Release-Basis, und Pakete werden nur ueber Tags, GitHub-Releases oder einen
expliziten Package-Workflow-Lauf veroeffentlicht.

Das Modell passt zur vorhandenen Repository-Infrastruktur:

- GitHub Actions `CI` fuehrt `make check`, `make coverage`, `make vulncheck`
  und die SBOM-Erzeugung vollstaendig auf GitHub-gehosteten Linux-Runnern aus.
- Bei Pushes auf `main` baut `CI` zusaetzlich Pakete und installiert das
  erzeugte Debian-Paket einmal im Ubuntu-Runner als neutralen Smoke-Test.
- `Package Repositories` baut Linux-Pakete und die Release-SBOM auf Ubuntu,
  deployed die Linux-only-GitHub-Pages-Seite und erstellt oder aktualisiert
  fuer einen `v*`-Tag den GitHub-Pre-Release.
- `Tier 3 FreeBSD Packaging` ist ein separater manueller Workflow. Er nutzt den
  self-hosted FreeBSD-Builder-Pfad nur nach explizitem Dispatch und ist nie
  Bestandteil von `CI` oder des automatischen Package-Workflows.
- Das `github-pages`-Environment erlaubt Deployments bereits nur von `main` und
  `v*`-Tags.

## Go-Toolchain

`go.mod` und `go.work` deklarieren die unterstuetzte Go-Minor-Linie `1.25`.
Patchversionen und eine separate `toolchain`-Direktive werden bewusst nicht
festgeschrieben.

Alle GitHub-Actions-Jobs lesen `go.mod` ueber `actions/setup-go` mit
`check-latest: true`. GitHub waehlt damit den neuesten verfuegbaren Go-`1.25`-
Patch. Der Docker-Build folgt mit `golang:1.25-bookworm` derselben Minor-Linie.

## Branch-Stufen

| Stufe | Branch | Zweck | Automatische Checks | Ergebnis |
| --- | --- | --- | --- | --- |
| Aenderung | kurzlebiger `codex/*`-, `feat/*`- oder `fix/*`-Branch | Eine fokussierte Aenderung | `CI / check` im Pull Request | Keine Paketveroeffentlichung |
| Integration | `dev` | Reviewte Arbeit vor Release-Promotion sammeln | `CI / check` im Pull Request und bei Push | Keine Paketveroeffentlichung |
| Release-Basis | `main` | Stabile, getestete Quelle fuer Tags und oeffentliche Pakete | `CI / check` plus Package-Build und Debian-Install-Smoke bei Push | Paketartefakte im Workflow-Lauf |
| Paketquelle | `v*`-Tag, GitHub-Pre-Release oder manueller Workflow von `main` | Standard-Paketset veroeffentlichen | Linux-Package-Matrix und Release-SBOM | Linux-only-GitHub-Pages und bei einem Tag ein GitHub-Pre-Release |
| FreeBSD-Erweiterung | Manueller Tier-3-Dispatch | FreeBSD-/OPNsense-Paketevidenz bauen oder veroeffentlichen | Self-hosted FreeBSD-Builder-Pfad | Nur Lauf-Evidenz oder FreeBSD-Release-Artefakte und kombinierte Pages |

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
6. Der `main`-Push startet `CI / check`; danach baut der Package-Job
   Linux-Paketformate und installiert das erzeugte Debian-Paket einmal in
   GitHub Actions.
7. Paketquellen nicht von `dev` deployen.

Direkte Pushes auf `main` bleiben expliziten Notfaellen oder Automation
vorbehalten. Der normale Weg ist Pull-Request-Review nach `main`.

## Release Und Paketveroeffentlichung

Versionsnummern kommen aus `v*`-Tags oder einem expliziten Package-Workflow-
Input. `VERSION`, Changelog-Abschnitt und Release-Tag muessen uebereinstimmen.

Der normale Alpha-Release-Pfad ist:

```sh
git switch main
git pull --ff-only origin main
git tag -a v0.2.5-alpha -m "unifi-stubd v0.2.5-alpha"
git push origin v0.2.5-alpha
```

Ein `v*`-Tag startet `Package Repositories`. Linux-Pakete, SPDX-SBOM,
Linux-only-Pages und der GitHub-Pre-Release werden zuerst veroeffentlicht.
Manuelle Linux-Package- und Pages-Builds koennen von `main` laufen:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.5-alpha \
  -f package_release=1
```

Wenn `version` bei einem manuellen Lauf fehlt, liest der Workflow `VERSION`.
Dieses manuelle Event erstellt oder aktualisiert keinen GitHub-Release. Wenn
der Pre-Release selbst repariert werden muss, wird der fehlgeschlagene Tag-
Workflow-Lauf erneut ausgefuehrt.

FreeBSD-Packaging ist ein unabhaengiger Tier-3-Vorgang. Der Input `publish`
steht standardmaessig auf `false`; dieser Modus baut den ausgewaehlten Ref und
behaelt nur sieben Tage gueltige Actions-Artefakte sowie die
Job-Zusammenfassung als Evidenz:

```sh
gh workflow run tier3-freebsd.yml --ref main -f publish=false
```

Um einen bestehenden passenden `v*`-Pre-Release zu erweitern und kombinierte
Linux-/FreeBSD-Pages zu deployen, wird derselbe Tag mit explizit aktivierter
Veroeffentlichung gestartet:

```sh
gh workflow run tier3-freebsd.yml --ref <v-tag-mit-tier3-workflow> -f publish=true
```

Der Publish-Modus fuegt FreeBSD-Tarballs, das native `pkg`-Repository-Archiv
und `SHA256SUMS.freebsd` hinzu. Die automatische Linux-`SHA256SUMS` wird nicht
ersetzt. Tier 3 ist nie Trigger, `needs`-Abhaengigkeit oder erforderliches
Ergebnis von `CI` oder `Package Repositories`.
Legacy-Tags ohne den Tier-3-Workflow oder mit FreeBSD-Eintraegen in
`SHA256SUMS` sind fuer diese getrennte Erweiterung nicht zulaessig.

## Hotfixes

1. Branch von `main` erstellen.
2. Pull Request nach `main` oeffnen.
3. Dieselben `main`-Gates ausfuehren.
4. Erst taggen oder veroeffentlichen, wenn der `main`-CI-Lauf gruen ist.
5. Den Hotfix nach `dev` zurueckfuehren, bevor normale Entwicklung weitergeht.

## Gate-Auswahl

| Aenderungstyp | Erforderlicher lokaler Gate | Zusaetzlicher Gate |
| --- | --- | --- |
| Go-Code, Config-Schema, Profildaten | `make check`, `make coverage`, `git diff --check` | Gezieltes `go test ./tests/...`, wenn sinnvoll |
| Inform, Adoption, Controller-Payload, Profil-Rendering | `make check` | Standardisierter Docker-Gate mit `make integration-docker` vor `dev` nach `main` |
| Paketierte Config, Service-Dateien, Paket-Metadaten | `make check`, `make package` | GitHub-`main`-Package-Install-Smoke |
| FreeBSD- oder OPNsense-Runtime-Verhalten | `make check` | Manueller Lauf von `Tier 3 FreeBSD Packaging` plus FreeBSD-/OPNsense-Smoke mit temporaerem State, wenn Verhalten geaendert wurde |
| Release Notes, Paketveroeffentlichung | `make check` | Linux-first `Package Repositories`; optionale manuelle Tier-3-Veroeffentlichung erst nach vorhandenem passenden Pre-Release |

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
- `tier3-freebsd`-Environment: Freigabe verlangen, nur `main` und `v*`
  erlauben und dem self-hosted Build-Job keine Publish-Rechte geben.
- Runner-Rollen: `docker-dmz` und Docker-Daemon-Zugriff vom alten FreeBSD-Proxy
  entfernen. Auf einem Bare Host ist passwortloses sudo verboten. Ein
  wegwerfbarer, containerisierter Proxy darf containerlokales passwortloses
  sudo nur behalten, wenn das geschuetzte Environment `tier3-freebsd`
  `ALLOW_CONTAINER_LOCAL_SUDO=true` explizit setzt und der Workflow seinen
  Container-Marker, einen aktiven Seccomp-Filter, einen privaten Container-
  Root-Mount sowie die PID-1-Root-Grenze bestaetigt. Bei Einsehbarkeit muss sie
  dem Container-Root entsprechen; andernfalls muss sie durch die Sandbox selbst
  fuer Container-Root unzugaenglich sein. Der Workflow bestaetigt zudem das
  Fehlen gefaehrlicher Root-Capabilities, exponierter Blockgeraete, von Host-
  Root-Mounts und Host-Runtime-Sockets. Tier 2 benoetigt einen separaten Docker-
  Runner; kein Runner traegt gleichzeitig Tier-2- und Tier-3-Rollenlabels.

Der Package-Job laeuft absichtlich nach dem Merge auf `main`, weil er die exakt
committete Release-Basis prueft und Paketartefakte fuer diesen Lauf hochlaedt.

## Quellen

- GitHub Flow: https://docs.github.com/en/get-started/using-github/github-flow
- GitHub-Actions-Events: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
- Protected Branches und Required Checks: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches
- Deployment-Environments: https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments
- Kurzlebige Branches in Trunk-Based Development: https://trunkbaseddevelopment.com/short-lived-feature-branches/
