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

- `CI`: Validierung, Vulnerability Scan, SBOM, vertrauenswuerdiger
  FreeBSD-Artefaktcheck und `main`-only Linux-Package-Smoke.
- `Package Repositories`: Release-Paketbuild und GitHub-Pages-Deployment fuer
  unsignierte Alpha-Paketquellen.

## Build-Ziele

| Ziel | Toolchain | Einstieg | Output | OS/Arch | Umgebung | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `make build` | Go 1.25 Floor | `go build ./...` | Lokale Binaries im Go-Buildcache | Host OS/Arch | Lokal oder GitHub Ubuntu | Definiert, kein Upload |
| `make check` | Go, golangci-lint, Shell-Policy | `lint`, `validate-config`, `test` | Pass/fail Gate | Host OS/Arch | GitHub Ubuntu und lokal | Primaerer Pre-Merge-Gate |
| `make vulncheck` | Go `govulncheck` | `govulncheck ./...` | Pass/fail Gate | Host OS/Arch | GitHub Ubuntu | CI-Security-Gate |
| `make build-freebsd` | Go Cross-Compile | `cmd/unifi-stubd` | `dist/unifi-stubd_freebsd_<arch>` | FreeBSD `PKG_FREEBSD_GOARCH` | Lokal | Lokales Diagnoseziel, kein Release-CI |
| `make package` | Go plus nFPM | `scripts/package.sh` | `.deb`, `.rpm`, Arch-Paket, Linux-`.tar.gz` | Linux `amd64` oder `arm64` | GitHub Ubuntu | Release-Paketpfad |
| `make package-deb` | Go plus nFPM | `scripts/package.sh deb` | Debian-Paket | Linux `PKG_GOARCH` | Lokal oder GitHub Ubuntu | Definiert |
| `make package-rpm` | Go plus nFPM | `scripts/package.sh rpm` | RPM-Paket | Linux `PKG_GOARCH` | Lokal oder GitHub Ubuntu | Definiert |
| `make package-arch` | Go plus nFPM | `scripts/package.sh archlinux` | Arch-Linux-Paket | Linux `PKG_GOARCH` | Lokal oder GitHub Ubuntu | Definiert |
| `make package-tgz` | Go plus tar | `scripts/package.sh tgz` | OS-spezifischer Tarball | `PKG_GOOS`/`PKG_GOARCH` | Lokal | Definiert |
| `make package-freebsd-tgz` | Go Cross-Compile plus tar | `scripts/package.sh tgz` | FreeBSD-Tarball | FreeBSD `amd64` oder `arm64` | Lokal | Definiert, nicht Release-Workflow |
| `make package-freebsd-pkg-repos` | Go auf FreeBSD-Builder, FreeBSD `pkg`, tar | `scripts/package-freebsd-pkg-repos.sh` | Native FreeBSD-`pkg`-Repos und FreeBSD-Tarballs | `FreeBSD:14/15` fuer `amd64`, `aarch64`, `armv7`; Tarballs default `amd64`, `arm64` | Self-hosted FreeBSD-Builder; optional gemappte `jexec`-Jails | Release-BSD-Pfad |
| `make package-repos` | dpkg, createrepo-c, repo-add, tar | `scripts/build-package-repos.sh` | Statische Paketquellen-Seite | Paketmetadaten fuer definierte Ziele | GitHub Ubuntu | GitHub-Pages-Input |
| `make integration-docker` | Docker Compose und Go-Helfer | `lab/stub/scripts/run-docker-integration.sh` | Pass/fail Smoke | Linux-Container-Lab | Lokal/manuell | Definiert, nicht automatisch in CI |
| Root `Dockerfile` | Docker BuildKit, Go, Alpine | `Dockerfile` | Runtime-/Lab-Image mit Daemon und Inform-Proxy | Linux multi-arch-faehig | Lokal/manuell | Kein Publish-Ziel definiert |

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
| Jobs | `check`, `freebsd-build`, `package` |
| Runner | `ubuntu-latest`; self-hosted `[self-hosted, Linux, X64, freebsd-pkg-builder]` fuer vertrauenswuerdige FreeBSD-Builds |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `anchore/sbom-action@v0`, `actions/upload-artifact@v7` |
| Cache | Kein expliziter Cache. `setup-go`-Defaults gelten, soweit die Action sie bereitstellt. |
| Artefakte | `sbom`, `packages` bei `main`-Push |
| Sicherheitsgrenze | Self-hosted FreeBSD-Job wird fuer Pull Requests uebersprungen. |

`check` fuehrt `make check`, `make vulncheck`, SBOM-Erzeugung und Upload aus.
Auf `main` wird Go `stable` genutzt; andere Refs nutzen die Version aus
`go.mod`.

`freebsd-build` laeuft nur ausserhalb von Pull Requests. Der Job ruft
`make package-freebsd-pkg-repos` auf dem self-hosted BSD-Paketbuilder auf. Der
kanonische Builder nutzt Go-Cross-Compile auf dem FreeBSD-Host plus FreeBSD
`pkg`; `FREEBSD_PKG_BUILD_JAILS` ist optional und nur fuer eigene Builder mit
laufenden `jexec`-Jails gedacht.

`package` laeuft nur bei Pushes auf `main`, nachdem `check` und
`freebsd-build` erfolgreich waren. Der Job ermittelt den neuesten
`v[0-9]*`-Tag, baut Linux-Pakete, installiert das Debian-Paket einmal,
validiert die installierte Config und laedt `dist/packages/` hoch.

### `Package Repositories`

| Bereich | Aktueller Wert |
| --- | --- |
| Datei | `.github/workflows/package-pages.yml` |
| Trigger | `workflow_dispatch`, `v*`-Tag-Push, published/prereleased GitHub Release |
| Default-Permissions | `contents: read` |
| Deploy-Permissions | `contents: read`, `pages: write`, `id-token: write` nur im Deploy-Job |
| Jobs | `build-packages`, `build-freebsd-artifacts`, `build-and-deploy` |
| Runner | Ubuntu fuer Linux-Pakete und Pages-Seite; self-hosted Builder fuer BSD-Artefakte |
| Actions | `actions/checkout@v6`, `actions/setup-go@v6`, `actions/upload-artifact@v7`, `actions/download-artifact@v7`, `actions/configure-pages@v6`, `actions/upload-pages-artifact@v5`, `actions/deploy-pages@v5` |
| Artefakte | `linux-package-artifacts`, `freebsd-package-artifacts`, `freebsd-pkg-repos`, Pages-Artefakt |
| Environment | `github-pages` |
| Concurrency | Eine `github-pages`-Deployment-Gruppe, kein Cancel-in-progress |

Linux-Pakete entstehen auf Ubuntu fuer `amd64` und `arm64`. FreeBSD-Tarballs
und native FreeBSD-`pkg`-Repos entstehen ueber den self-hosted
BSD-Builder-Pfad. Der Deploy-Job fuegt nur bereits gebaute Artefakte zu
`dist/package-site` zusammen und deployed diese statische Seite.

## OS- und Architekturmatrix

| Ziel | Buildpfad | Testpfad | Packaging | Bereitstellung |
| --- | --- | --- | --- | --- |
| Linux amd64 | `PKG_GOARCH=amd64 make package` | `make check`; Debian-Install-Smoke auf `main` | Debian, RPM, Arch, Tarball | Actions-Artefakt und Pages-Paketquelle |
| Linux arm64 | `PKG_GOARCH=arm64 make package` | `make check`; kein Install-Smoke | Debian, RPM, Arch, Tarball | Pages-Paketquelle |
| FreeBSD 14 amd64 | `make package-freebsd-pkg-repos` via FreeBSD-Builder | Build-/Package-Verifikation; Zielhost-Smoke manuell | Native `pkg`-Repo | Pages-Paketquelle |
| FreeBSD 14 aarch64 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Pages-Paketquelle |
| FreeBSD 14 armv7 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Pages-Paketquelle |
| FreeBSD 15 amd64 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Pages-Paketquelle |
| FreeBSD 15 aarch64 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Pages-Paketquelle |
| FreeBSD 15 armv7 | Gleich | Build-/Package-Verifikation | Native `pkg`-Repo | Pages-Paketquelle |
| FreeBSD/OPNsense Tarball amd64 | Aus BSD-Builder-Pfad | Zielhost-Smoke manuell | Tarball | Pages und manuelle Release-Assets |
| FreeBSD/OPNsense Tarball arm64 | Aus BSD-Builder-Pfad | Zielhost-Smoke manuell | Tarball | Pages und manuelle Release-Assets |

Offene Luecken:

- Linux-arm64-Pakete werden gebaut, aber nicht installiert gesmoked.
- FreeBSD-/OPNsense-Runtime-Verhalten braucht weiterhin manuelle Zielhost-
  Evidenz.
- Signierung ist offen; Alpha-Paketquellen bleiben bewusst unsigniert.

## Buildhosts, Runner und Container

GitHub-hosted Ubuntu wird fuer generische Validierung, Linux-Packaging,
Repository-Metadaten und Pages-Deployment genutzt. Der self-hosted Runner
`freebsd-pkg-builder` wird nur fuer vertrauenswuerdige BSD-Arbeit genutzt und
laeuft nicht fuer Pull Requests. Damit wird vermieden, untrusted PR-Code auf
persistenter Projektinfrastruktur auszufuehren.

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

- Root-`Dockerfile`: Daemon und Inform-Proxy in Alpine-Runtime-Image.
- Lab-Gateway-Profil-Dockerfiles: Firmware- und Controller-Lab-Research-Pfade.
- Kein CI-Buildcontainer, kein GHCR-Image, kein Container-Release definiert.

## Packaging und Bereitstellung

| Format | Quelle | Build-Befehl | Name/Output | Validierung | Ziel |
| --- | --- | --- | --- | --- | --- |
| Debian | `packaging/nfpm.yaml`, Linux-Stage | `make package` oder `make package-deb` | `unifi-stubd_<version>-<release>_<arch>.deb` | Install- und Config-Smoke auf `main` fuer amd64 | Artefakt, APT-Repo |
| RPM | `packaging/nfpm.yaml`, Linux-Stage | `make package` oder `make package-rpm` | `unifi-stubd-<version>-<release>.<arch>.rpm` | Repository-Metadaten-Build | Artefakt, RPM-Repo |
| Arch Linux | `packaging/nfpm.yaml`, Linux-Stage | `make package` oder `make package-arch` | `unifi-stubd-<version>-<release>-<arch>.pkg.tar.zst` | Repository-Metadaten-Build | Artefakt, Arch-Repo |
| Linux-Tarball | Linux-Stage | `make package` oder `make package-tgz` | `unifi-stubd_<version>-<release>_linux_<arch>.tar.gz` | Checksums | Artefakt, Pages-Checksum |
| FreeBSD-Tarball | FreeBSD-Stage mit BSD-gebautem Binary | `make package-freebsd-pkg-repos` | `unifi-stubd_<version>-<release>_freebsd_<arch>.tar.gz` | Checksums | Artefakt, Pages-FreeBSD-Pfade |
| FreeBSD `pkg` | FreeBSD-Stage und UCL-Manifest | `make package-freebsd-pkg-repos` | `repo/FreeBSD:*/*` inkl. `packagesite.pkg` | `pkg repo` und Checksums | Pages-FreeBSD/pkg-Pfade |
| SBOM | Repository-Baum | `anchore/sbom-action@v0` | `dist/sbom.spdx.json` | Action-Erfolg | CI-Artefakt |
| Statische Paketseite | Package-Artefakte und FreeBSD-pkg-Repo | `make package-repos` | `dist/package-site/` | `checksums.txt`, Paketmetadaten-Tools | GitHub Pages |

## Deployment- und Release-Grenzen

- Pull Requests laufen nur generische CI-Validierung. Sie veroeffentlichen
  keine Pakete und laufen nicht auf dem self-hosted BSD-Builder.
- Pushes nach `dev` laufen durch Validierung plus vertrauenswuerdige
  BSD-Artefaktchecks, veroeffentlichen aber keine Paketquellen.
- Pushes nach `main` laufen durch Validierung, BSD-Artefaktchecks und den
  Linux-Package-Smoke. Sie laden nur Run-Artefakte hoch.
- `v*`-Tags, published/prereleased GitHub Releases und manueller Dispatch von
  `main` bauen und deployen die GitHub-Pages-Paketquellen.
- Die Workflows laden keine GitHub-Release-Assets hoch. Das bleibt der manuelle
  `gh release create ... dist/releases/...`-Pfad aus der Release-Checkliste.

## Gefundene Probleme und Korrekturen

- CI hatte keine expliziten Token-Rechte. Jetzt gilt `permissions: contents:
  read`.
- Der Pages-Workflow gab Pages/OIDC-Schreibrechte an alle Jobs. Diese Rechte
  liegen jetzt nur noch auf dem Deploy-Job.
- BSD-Tarballs wurden im Release-Workflow auf Ubuntu gebaut. BSD-Release-
  Artefakte kommen jetzt aus `make package-freebsd-pkg-repos`.
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
- Entscheiden, ob GitHub-Release-Asset-Upload automatisiert werden soll.
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
- Vertrauenswuerdiger Push nach `dev`: `CI / check` und `CI / freebsd-build`.
- Push nach `main`: `CI / check`, `CI / freebsd-build` und `CI / package`.
- Release-Paketseite: `v*`-Tag pushen, GitHub Release publishen/pre-releasen
  oder `package-pages.yml` manuell von `main` starten.

Gepruefte Referenzen:

- GitHub Actions Workflow-Syntax und `permissions`:
  `https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax`
- GitHub Pages Custom Workflow Anforderungen:
  `https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages`
- GitHub Secure-Use-Hinweise fuer self-hosted Runner und Action-Pinning:
  `https://docs.github.com/en/actions/reference/security/secure-use`
