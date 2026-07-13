# CI/CD-, Build-, Packaging- und Deployment-Katalog

Dieser Katalog beschreibt die im Repository definierte Build- und
Veroeffentlichungskette. Grundlage sind Workflows, Makefile-Ziele, Scripts,
Paketmetadaten und Release-Dokumentation. Nicht im Repository definierte Punkte
werden als offen markiert.

## Aktuelle Architektur

GitHub Actions orchestriert die Pipeline. Projektlogik liegt in `Makefile`,
`scripts/package.sh`, `scripts/package-freebsd-pkg-repos.sh` und
`scripts/build-package-repos.sh`. Go-Tools sind ueber `go.mod` als Tool-
Dependencies versioniert und laufen ueber `go tool`. Das Root-`Dockerfile` ist
ein Runtime-/Lab-Image-Pfad, kein CI-Buildcontainer. Ein GHCR-Publish-Workflow
ist nicht definiert.

Die Pipeline ist getrennt in:

- `CI`: Validierung, Vulnerability Scan, SBOM und ein `main`-only Linux-
  Package-Smoke, vollstaendig auf GitHub-hosted Ubuntu.
- `Tier 2 Docker Integration`: manuell gestarteter Controller-Adoption-Gate auf
  dem vertrauenswuerdigen Docker-Lab-Runner, unabhaengig von `CI`.
- `Package Repositories`: Release-Paket- und persistenter SBOM-Build,
  GitHub-Pre-Release-Artefakte sowie Pages-Deployment fuer die Linux-first
  unsignierten Alpha-Paketquellen.
- `Tier 3 FreeBSD Packaging`: manuell gestarteter FreeBSD-Paketbuild. Er kann
  entweder nur sieben Tage aufbewahrte Run-Artefakte erzeugen oder einen
  bestehenden `v*`-Pre-Release ergaenzen und Pages mit FreeBSD-Outputs neu
  bauen.

## Build-Ziele

| Ziel | Toolchain | Einstieg | Output | OS/Arch | Umgebung | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `make build` | Aktueller von GitHub unterstuetzter Go-1.25-Patch | `go build ./...` | Lokale Binaries im Go-Buildcache | Host OS/Arch | Lokal oder GitHub Ubuntu | Definiert, kein Upload |
| `make check` | Go 1.25, golangci-lint, Shell-Policy | `lint`, `validate-config`, `test` | Pass/fail Gate | Host OS/Arch | GitHub Ubuntu und lokal | Primaerer Pre-Merge-Gate |
| `make coverage` | Go-1.25-Coverage-Instrumentierung | Externe Tests unter `tests/...` mit allen produktiven `internal/...`-Paketen | `dist/cover.out`, `dist/coverage.html` und ein Mindest-Gate von 60 Prozent | Host OS/Arch | GitHub Ubuntu und lokal | CI-Qualitaets-Gate |
| `make vulncheck` | Go 1.25 `govulncheck` | `govulncheck ./...` | Pass/fail Gate | Host OS/Arch | GitHub Ubuntu | CI-Security-Gate |
| `make build-freebsd` | Go 1.25 Cross-Compile | `cmd/unifi-stubd` | `dist/unifi-stubd_freebsd_<arch>` | FreeBSD `PKG_FREEBSD_GOARCH` | Lokal | Lokales Diagnoseziel, kein Release-CI |
| `make package` | Go 1.25 plus nFPM | `scripts/package.sh` | `.deb`, `.rpm`, Arch-Paket, Linux-`.tar.gz` | Linux `amd64` oder `arm64` | GitHub Ubuntu | Release-Paketpfad |
| `make package-deb` | Go plus nFPM | `scripts/package.sh deb` | Debian-Paket | Linux `PKG_GOARCH` | Lokal oder GitHub Ubuntu | Definiert |
| `make package-rpm` | Go plus nFPM | `scripts/package.sh rpm` | RPM-Paket | Linux `PKG_GOARCH` | Lokal oder GitHub Ubuntu | Definiert |
| `make package-arch` | Go plus nFPM | `scripts/package.sh archlinux` | Arch-Linux-Paket | Linux `PKG_GOARCH` | Lokal oder GitHub Ubuntu | Definiert |
| `make package-tgz` | Go plus tar | `scripts/package.sh tgz` | OS-spezifischer Tarball | `PKG_GOOS`/`PKG_GOARCH` | Lokal | Definiert |
| `make package-freebsd-tgz` | Go Cross-Compile plus tar | `scripts/package.sh tgz` | FreeBSD-Tarball | FreeBSD `amd64` oder `arm64` | Lokal | Definiert, nicht Release-Workflow |
| `make package-freebsd-pkg-repos` | Go auf FreeBSD-Builder, FreeBSD `pkg`, tar | `scripts/package-freebsd-pkg-repos.sh` | Native FreeBSD-`pkg`-Repos und FreeBSD-Tarballs | `FreeBSD:14/15` fuer `amd64`, `aarch64`, `armv7`; Tarballs default `amd64`, `arm64` | Manueller Tier-3-Runner; optional gemappte `jexec`-Jails | Tier-3-FreeBSD-Pfad |
| `make package-repos` | dpkg, createrepo-c, repo-add, tar | `scripts/build-package-repos.sh` | Statische Paketquellen-Seite | Paketmetadaten fuer vorhandene Ziele | GitHub Ubuntu | Linux-first oder kombinierter GitHub-Pages-Input |
| `make integration-docker` | Docker Compose, gepinntes UniFi Network und Go-Helfer | `lab/stub/scripts/run-docker-integration.sh` | Tier-2-Controller-Adoption-Gate | Linux-Container-Lab | Lokal oder manueller GitHub-Actions-Dispatch | Im eigenstaendigen Tier-2-Workflow definiert |
| Root `Dockerfile` | Docker BuildKit, Go, Alpine | `Dockerfile` | Runtime-/Lab-Image mit Daemon und Inform-/Discovery-Proxy | Linux multi-arch-faehig | Lokal/manuell | Kein Publish-Ziel definiert |

Offene Build-Zieldefinitionen:

- Keine macOS- oder Windows-Buildziele sind definiert.
- Kein Top-Level-`tools/`-Verzeichnis existiert; Helfer liegen unter
  `lab/stub/tools/` und in Lab-Profilen.
- Kein CI-Buildcontainer und kein GitHub-Container-Registry-Publishpfad sind
  definiert.

## GitHub Actions

### `CI`

| Bereich | Aktueller Wert |
| --- | --- |
| Datei | `.github/workflows/ci.yml` |
| Trigger | `workflow_dispatch`, Pushes nach `dev` und `main`, `pull_request` |
| Default-Permissions | `contents: read` |
| Jobs | `check`, `package` |
| Runner | Nur `ubuntu-latest` |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `anchore/sbom-action@v0`, `actions/upload-artifact@v7` |
| Cache | Kein expliziter Cache. `setup-go`-Defaults gelten, soweit die Action sie bereitstellt. |
| Artefakte | `coverage`, `sbom`; `packages` bei `main`-Push, alle sieben Tage aufbewahrt |
| Sicherheitsgrenze | Kein self-hosted Runner und kein Zugriff auf den FreeBSD-Builder. |

`check` fuehrt `make check`, `make coverage`, `make vulncheck` und die SBOM-
Erzeugung aus, laedt Coverage und SBOM fuer sieben Tage hoch und liest die Go-
Version aus `go.mod`.

`package` laeuft nur bei Pushes auf `main`, nachdem `check` erfolgreich war.
Der Job liest den zentralen Wert aus `VERSION`, baut Linux-Pakete, installiert
das Debian-Paket einmal, validiert die installierte Config und laedt
`dist/packages/` fuer sieben Tage hoch.

### `Tier 2 Docker Integration`

| Bereich | Aktueller Wert |
| --- | --- |
| Datei | `.github/workflows/tier2-docker.yml` |
| Trigger | Nur `workflow_dispatch` |
| Rechte | `contents: read` |
| Job | `controller-adoption` |
| Runner | `[self-hosted, Linux, X64, docker-dmz, unifi-stubd]` |
| Befehl | `make integration-docker` |
| State-Grenze | Frisches Compose-Projekt und Volumes vor dem Lauf; bedingungsloser Teardown danach |
| Artefakte | Keine; Captures und Controller-State werden nie hochgeladen |
| Concurrency | Eine Gruppe `unifi-stubd-tier2-docker`, kein Cancel-in-progress |

Der Runner-Preflight prueft Linux/X64, Go, Docker-Daemon-Zugriff, Docker Compose
v2, sichtbare Ressourcen und die kombinierte Lab-Konfiguration. Der Workflow
verwendet das gepinnte Controller-Image `10.4.57-ls135`, schreibt bereinigte
Laufevidenz in die Job-Zusammenfassung und hat keine Abhaengigkeit von oder
Trigger durch den normalen `CI`-Workflow, Package-Workflows, Pull Requests,
Pushes oder Zeitplaene.

### `Tier 3 FreeBSD Packaging`

| Bereich | Aktueller Wert |
| --- | --- |
| Datei | `.github/workflows/tier3-freebsd.yml` |
| Trigger | Nur `workflow_dispatch`, mit erforderlichem Boolean-Input `publish` |
| Default-Rechte | `contents: read` |
| Runner | `[self-hosted, Linux, X64, unifi-stubd, tier3, freebsd-pkg-builder]` nur fuer den Build |
| Build-Befehl | `make package-freebsd-pkg-repos` |
| Build-only-Modus | `publish=false`: ein sieben Tage aufbewahrtes Run-Artefakt `tier3-freebsd-release-<sha>` mit zwei Tarballs, Repository-Archiv und `SHA256SUMS.freebsd` |
| Veroeffentlichungsmodus | `publish=true`: nur auf einem bestehenden passenden `v*`-Pre-Release; laedt FreeBSD-Artefakte und `SHA256SUMS.freebsd` hoch und baut danach kombinierte Pages |
| Concurrency | Eine Build-Gruppe `unifi-stubd-tier3-freebsd`; die Pages-Veroeffentlichung nutzt zusaetzlich die gemeinsame Gruppe `github-pages` |

Der Ubuntu-Validierungsjob ermittelt `VERSION`. Jeder Lauf muss `main` oder den
passenden `v*`-Tag verwenden. Eine Veroeffentlichung muss auf diesem Tag
gestartet werden. Zu diesem Tag muss bereits ein Pre-
Release mit den erwarteten Linux-Paketen, der persistenten SBOM und
`SHA256SUMS` existieren. Der Workflow veraendert weder einen finalen Release
noch einen unvollstaendigen Pre-Release, einen Release mit signierten Checksum-
Artefakten oder eine Legacy-`SHA256SUMS`, die FreeBSD bereits abdeckt. Der
gewaehlte Release-Tag muss diesen Workflow selbst enthalten. Der Tier-3-
Preflight prueft Linux/X64, SSH/SCP/tar, fehlenden Docker-Daemon-Zugriff sowie
den non-root FreeBSD-Buildhost mit nativem `pkg`, `tar` und Go vor dem Build.
Auf einem Bare-Host-Proxy lehnt er passwortloses sudo ab. Bei einem
wegwerfbaren, containerisierten Proxy ist containerlokales passwortloses sudo
nur zulaessig, wenn das geschuetzte Environment `tier3-freebsd` explizit
`ALLOW_CONTAINER_LOCAL_SUDO=true` setzt. Der Preflight muss einen Container-
Marker, einen aktiven Seccomp-Filter, einen privaten Container-Root-Mount und
dasselbe Root-Dateisystem fuer PID 1 bestaetigen. Zudem darf er weder
gefaehrliche Root-Capabilities, exponierte Blockgeraete, Host-Root-Mounts noch
Host-Runtime-Sockets finden.

Tier 3 ist weder Bestandteil von `CI` noch des Tier-2-Controller-Labs oder des
automatischen Tag-Workflows. Ein fehlender oder fehlschlagender FreeBSD-Lauf
blockiert daher weder Linux-Pakete noch die Linux-first Pages-Seite oder die
Erstellung des Pre-Releases.

### `Package Repositories`

| Bereich | Aktueller Wert |
| --- | --- |
| Datei | `.github/workflows/package-pages.yml` |
| Trigger | `workflow_dispatch`, `v*`-Tag-Push |
| Default-Permissions | `contents: read` |
| Deploy-Permissions | `contents: read`, `pages: write`, `id-token: write` nur im Deploy-Job |
| Jobs | `build-packages`, `build-sbom`, `build-and-deploy`, `publish-release` |
| Runner | Nur `ubuntu-latest` |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `anchore/sbom-action@v0`, `actions/upload-artifact@v7`, `actions/download-artifact@v7`, `actions/configure-pages@v6`, `actions/upload-pages-artifact@v5`, `actions/deploy-pages@v5` |
| Artefakte | Sieben Tage aufbewahrte `linux-package-artifacts` und `release-sbom`, Pages-Artefakt, persistente Linux-/SBOM-Release-Artefakte |
| Environment | `github-pages` |
| Concurrency | Eine `github-pages`-Deployment-Gruppe, kein Cancel-in-progress |

Linux-Pakete entstehen auf Ubuntu fuer `amd64` und `arm64`. Der Deploy-Job baut
und veroeffentlicht ein Linux-only `dist/package-site`. Bei einem `v*`-Tag
erstellt oder aktualisiert `publish-release` den Pre-Release mit Linux-Paketen,
`unifi-stubd.spdx.json` und `SHA256SUMS`. Dieser Workflow waehlt nie einen self-
hosted Runner aus und wartet nicht auf Tier 3.

## OS- und Architekturmatrix

| Ziel | Buildpfad | Testpfad | Packaging | Bereitstellung |
| --- | --- | --- | --- | --- |
| Linux amd64 | `PKG_GOARCH=amd64 make package` | `make check`; Debian-Install-Smoke auf `main` | Debian, RPM, Arch, Tarball | Actions-Artefakt und Pages-Paketquelle |
| Linux arm64 | `PKG_GOARCH=arm64 make package` | `make check`; kein Install-Smoke | Debian, RPM, Arch, Tarball | Pages-Paketquelle |
| FreeBSD 14 amd64 | Manueller Tier-3-FreeBSD-Builder | Build-/Package-Verifikation; Zielhost-Smoke manuell | Native `pkg`-Repo | Sieben-Tage-Artefakt oder bestehender Pre-Release und kombinierte Pages mit `publish=true` |
| FreeBSD 14 aarch64 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Gleich |
| FreeBSD 14 armv7 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Gleich |
| FreeBSD 15 amd64 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Gleich |
| FreeBSD 15 aarch64 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Gleich |
| FreeBSD 15 armv7 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Gleich |
| FreeBSD/OPNsense Tarball amd64 | Manueller Tier-3-Builder-Pfad | Zielhost-Smoke manuell | Tarball | Sieben-Tage-Artefakt oder bestehender Pre-Release und kombinierte Pages mit `publish=true` |
| FreeBSD/OPNsense Tarball arm64 | Manueller Tier-3-Builder-Pfad | Zielhost-Smoke manuell | Tarball | Gleich |

Offene Luecken:

- Linux-arm64-Pakete werden gebaut, aber nicht installiert gesmoked.
- FreeBSD-/OPNsense-Runtime-Verhalten braucht weiterhin manuelle Zielhost-
  Evidenz.
- Signierung ist offen; Alpha-Paketquellen bleiben bewusst unsigniert.

## Buildhosts, Runner und Container

GitHub-hosted Ubuntu wird fuer generische Validierung, Linux-Packaging,
Repository-Metadaten, normale Release-Veroeffentlichung und Pages-Deployment
genutzt. Normale CI und der automatische Tag-Workflow waehlen nie einen self-
hosted Runner aus.

Der manuelle Tier-2-Selektor ist `[self-hosted, Linux, X64, docker-dmz,
unifi-stubd]`; er fuehrt nur `lab/stub` aus und benoetigt Docker-Daemon-Zugriff
und Docker Compose v2. Der manuelle Tier-3-Selektor ist `[self-hosted, Linux,
X64, unifi-stubd, tier3, freebsd-pkg-builder]`; er stellt den SSH-
orchestrierten FreeBSD-Paketpfad bereit. Das muessen getrennte Runner sein. Der
alte FreeBSD-Proxy verliert das Label `docker-dmz` und Docker-Daemon-Zugriff;
Tier 2 benoetigt danach einen anderen `docker-dmz`-Runner. Auf einem Bare-Host-
Proxy bleibt passwortloses sudo verboten. Es ist nur innerhalb eines
wegwerfbaren, containerisierten Proxys erlaubt, dessen Workflow-Preflight einen
Container-Marker, einen aktiven Seccomp-Filter und einen privaten Container-
Root-Mount, dasselbe Root-Dateisystem fuer PID 1 sowie das Fehlen gefaehrlicher
Root-Capabilities, exponierter Blockgeraete, von Host-Root-Mounts und Host-
Runtime-Sockets bestaetigt. Diese Ausnahme verlangt zusaetzlich die geschuetzte
Environment-Variable `ALLOW_CONTAINER_LOCAL_SUDO=true`. Kein Runner darf
gleichzeitig Tier-2- und Tier-3-Rollenlabels tragen. Das Environment bekommt
Required Reviewer und Deployment-Regeln nur fuer `main` und `v*`.

Der Rollen-Split wird in dieser Reihenfolge aktiviert: Ersatz-Runner fuer Tier
2 registrieren und preflighten; das geschuetzte Environment `tier3-freebsd`
anlegen; Docker-Zugriff und `docker-dmz` vom alten Proxy entfernen;
passwortloses sudo auf dem Bare Host entfernen oder auf einen wegwerfbaren
Container begrenzen, der die Marker-, Seccomp-, Private-Root-Mount-,
PID-1-Root-, Capability-, Blockgeraete-, Host-Root-Mount- und Runtime-Socket-
Pruefungen besteht; `ALLOW_CONTAINER_LOCAL_SUDO=true` im geschuetzten
Environment nur fuer diesen Container setzen; `tier3` setzen; danach einen
Build von `main` mit `publish=false` starten. So entsteht kein stilles
Zeitfenster ohne geeigneten Docker-Runner.

Der BSD-Builder muss bereitstellen:

- SSH-Alias `unifi-stubd-freebsd-builder`, sofern `FREEBSD_PKG_REMOTE` nicht
  ueberschrieben wird.
- Arbeitsverzeichnis aus `FREEBSD_PKG_REMOTE_DIR`, default
  `/tmp/unifi-stubd-freebsd-pkg`.
- FreeBSD `pkg`, `tar` und ein Go-Kommando namens `go`, `go125` oder `go126`.
- Optional `jexec` und `FREEBSD_PKG_BUILD_JAILS` nur fuer eigene Builder mit
  laufenden Jails. Poudriere-Ziel-Jails werden nicht als laufende `jexec`-
  Buildumgebungen vorausgesetzt.

Container sind nur fuer Runtime und Lab definiert:

- Root-`Dockerfile`: Daemon und Inform-/Discovery-Proxy im Alpine-Runtime-Image.
- Lab-Gateway-Profil-Dockerfiles: Firmware- und Controller-Lab-Research-Pfade.
- Kein CI-Buildcontainer, kein GHCR-Image, kein Container-Release definiert.

## Packaging und Bereitstellung

| Format | Quelle | Build-Befehl | Name/Output | Validierung | Ziel |
| --- | --- | --- | --- | --- | --- |
| Debian | `packaging/nfpm.yaml`, Linux-Stage | `make package` oder `make package-deb` | `unifi-stubd_<version>-<release>_<arch>.deb` | Install- und Config-Smoke auf `main` fuer amd64 | Artefakt, APT-Repo |
| RPM | `packaging/nfpm.yaml`, Linux-Stage | `make package` oder `make package-rpm` | `unifi-stubd-<version>-<release>.<arch>.rpm` | Repository-Metadaten-Build | Artefakt, RPM-Repo |
| Arch Linux | `packaging/nfpm.yaml`, Linux-Stage | `make package` oder `make package-arch` | `unifi-stubd-<version>-<release>-<arch>.pkg.tar.zst` | Repository-Metadaten-Build | Artefakt, Arch-Repo |
| Linux-Tarball | Linux-Stage | `make package` oder `make package-tgz` | `unifi-stubd_<version>-<release>_linux_<arch>.tar.gz` | Checksums | Artefakt, Pages-Checksum |
| FreeBSD-Tarball | FreeBSD-Stage mit BSD-gebautem Binary | `make package-freebsd-pkg-repos` | `unifi-stubd_<version>-<release>_freebsd_<arch>.tar.gz` | Bei Veroeffentlichung `SHA256SUMS.freebsd` | Manuelles Tier-3-Artefakt; optional Pre-Release und kombinierte Pages |
| FreeBSD `pkg` | FreeBSD-Stage und UCL-Manifest | `make package-freebsd-pkg-repos` | `repo/FreeBSD:*/*` inkl. `packagesite.pkg` | `pkg repo` und Checksums | Manuelles Tier-3-Repo-Archiv; optional kombinierte Pages |
| SBOM | Repository-Baum | `anchore/sbom-action@v0` | `unifi-stubd.spdx.json` | Action-Erfolg | CI-Artefakt und persistentes GitHub-Release-Artefakt |
| Statische Paketseite | Linux-Pakete, optional plus veroeffentlichte Tier-3-FreeBSD-Outputs | `make package-repos` | `dist/package-site/` | `checksums.txt`, Paketmetadaten-Tools | Linux-first oder kombinierte GitHub Pages |

## Deployment- und Release-Grenzen

- Pull Requests laufen nur generische CI-Validierung. Sie veroeffentlichen
  keine Pakete und laufen nicht auf dem self-hosted BSD-Builder.
- Pushes nach `dev` fuehren nur generische Validierung aus.
- Pushes nach `main` fuehren Validierung und den Linux-Package-Smoke aus. Sie
  laden nur sieben Tage aufbewahrte Run-Artefakte hoch.
- `v*`-Tags bauen Linux-Pakete und SBOM, erstellen oder aktualisieren den
  GitHub-Pre-Release, haengen den Linux-Release-Satz plus `SHA256SUMS` an und
  deployen die Linux-first Pages-Paketquellen.
- Manueller Dispatch von `main` baut die Pages-Paketquellen neu, ohne einen
  neuen GitHub-Release zu erstellen.
- Ein manueller Dispatch von `Tier 2 Docker Integration` prueft nur den
  ausgewaehlten Ref und startet oder veraendert weder `CI` noch Package-
  Veroeffentlichung, Branch-Regeln oder Required Checks.
- Ein manueller Dispatch von `Tier 3 FreeBSD Packaging` mit `publish=false`
  baut nur sieben Tage aufbewahrte FreeBSD-Run-Artefakte. Mit `publish=true`
  auf einem bestehenden passenden `v*`-Pre-Release ergaenzt er FreeBSD-
  Artefakte und `SHA256SUMS.freebsd` und ersetzt danach die Linux-only Pages-
  Seite durch die kombinierte Linux-/FreeBSD-Seite.

## Gefundene Probleme und Korrekturen

- CI hatte keine expliziten Token-Rechte. Jetzt gilt `permissions: contents:
  read`.
- Der Pages-Workflow gab Pages/OIDC-Schreibrechte an alle Jobs. Diese Rechte
  liegen jetzt nur noch auf dem Deploy-Job.
- BSD-Tarballs waren zuvor Bestandteil automatischer CI- und Release-
  Workflows. Sie kommen jetzt ausschliesslich aus dem manuell gestarteten Tier-
  3-Workflow; Linux-CI, Release-Artefakte und Pages warten nicht mehr auf den
  BSD-Builder.
- Das FreeBSD-Packaging-Script baute Binaries vorher lokal und nutzte den
  BSD-Host nur fuer `pkg create`/`pkg repo`. Jetzt wird der getrackte
  Source-Tree zum Builder uebertragen und jede konfigurierte FreeBSD-ABI dort
  gebaut. Der kanonische Poudriere-Host nutzt Host-seitiges Go-Cross-Compile,
  weil seine ARM-Poudriere-Ziele die Go-Toolchain unter qemu-user-static nicht
  bauen koennen.
- Release- und Entwicklungsdoku beschrieben den alten Cross-Build-Pfad. Die
  Doku beschreibt jetzt den Builder-Pfad und den optionalen `jexec`-Jail-Modus.

## Offene Punkte

- Entscheiden, ob die privaten Poudriere-Host-Repositories zusaetzlich zur
  GitHub-Pages-Paketquelle veroeffentlicht werden sollen.
- Entscheiden, ob Drittanbieter-Actions per vollstaendiger SHA gepinnt werden
  sollen. Aktuell nutzt das Projekt stabile Major-Version-Tags.
- Entscheiden, ob Dependabot fuer GitHub Actions aktiviert werden soll.
- Entscheiden, ob Linux arm64 einen Install-Smoke braucht.
- Entscheiden, ob Tier 3 vor `publish=true` Zielhost-Install-/Upgrade-Smoke-
  Tests erhalten soll.
- Entscheiden, ab wann Paket-/Checksum-Signaturen fuer nicht-Alpha-
  Veroeffentlichungen Pflicht sind.
- Entscheiden, ob ein CI-Buildcontainer notwendig ist. Aktuell erzwingt die
  Projektstruktur keinen.

## Ausloesen und Pruefen

Lokale Gates:

```sh
make check
make vulncheck
make package
make package-freebsd-pkg-repos
make package-repos
```

GitHub-Gates:

- Pull Request: `CI / check`.
- Push nach `dev`: `CI / check`.
- Push nach `main`: `CI / check` und `CI / package`.
- Linux-Release und Paketquellen-Seite: `v*`-Tag pushen oder
  `package-pages.yml` manuell von `main` starten.
- Tier-2-Controller-Gate: `tier2-docker.yml` manuell fuer den exakten Ref
  starten.
- Nur Tier-3-FreeBSD-Build: `gh workflow run tier3-freebsd.yml --ref main -f publish=false`.
- Tier-3-FreeBSD-Veroeffentlichung: nachdem der Linux-Pre-Release existiert,
  `gh workflow run tier3-freebsd.yml --ref <neuer-v-tag-mit-tier3-workflow> -f publish=true` ausfuehren.

Gepruefte Referenzen:

- GitHub Actions Workflow-Syntax und `permissions`:
  `https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax`
- GitHub Pages Custom Workflow Anforderungen:
  `https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages`
- GitHub Secure-Use-Hinweise fuer self-hosted Runner und Action-Pinning:
  `https://docs.github.com/en/actions/reference/security/secure-use`
