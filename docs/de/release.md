# Release-Checkliste

Nutze diese Checkliste vor einem veroeffentlichten Tag oder Paket-Set.

Release-Artefakte sind bewusst neutral. Sie duerfen keine echten
Controller-URLs, Site-IP-Adressen, MAC-Adressen, Clientnamen, Credentials,
Adoption-State-Dateien oder host-spezifische Interface-Mappings enthalten.
Solche Dateien gehoeren in eine private Konfigurationsablage ausserhalb dieses
Repositories und werden nach der Paketinstallation auf den jeweiligen Host
kopiert.

## Vor dem Tag

1. `make check` ausfuehren.
2. `make package` ausfuehren.
3. Cross-Architektur-Ziele fuer Releases bauen, z.B. `PKG_GOARCH=arm64 make
   package`. FreeBSD-Release-Artefakte ueber den konfigurierten
   FreeBSD-Builder mit `make package-freebsd-pkg-repos` bauen.
4. `dist/packages/` auf Debian-, RPM-, Arch-Linux- und `.tar.gz`-Ausgaben pruefen.
5. Nach Dependency- oder Tool-Updates `go.mod`, `go.work` und `go.sum` pruefen.
6. Private Lab-Daten suchen:

   ```sh
   sh scripts/check-policy.sh
   ```

7. Pruefen, dass `VERSION`, der neueste datierte Abschnitt in `CHANGELOG.md`
   und der vorgesehene Tag `v<version>` uebereinstimmen.
8. `CREDITS.md` und `NOTICE.md` aktualisieren, wenn neue Quellen, kopierter
   Code, Pakete oder Tools hinzukommen.
9. Pruefen, dass `packaging/linux/etc/unifi-stubd/config.yaml` und `lab/` nur
   Dokumentationsadressen oder neutrale Defaults enthalten.
10. Repository-Metadaten mit `gh repo view --json
    description,homepageUrl,licenseInfo,repositoryTopics` pruefen und
    Paketname, Lizenz, Maintainer und Projekt-URL mit `packaging/nfpm.yaml`
    vergleichen.
11. Signaturstatus pruefen. Alpha-Releases duerfen nur als Pre-Release mit der
    erzeugten unsignierten `SHA256SUMS` veroeffentlicht werden. Stabile Releases
    verlangen einen signierten Tag und eine abgetrennte Signatur fuer
    `SHA256SUMS`.

## Tag setzen

Semantische Versionstags verwenden:

```sh
git tag -a v0.2.4-alpha -m "unifi-stubd v0.2.4-alpha"
git push origin v0.2.4-alpha
```

Nur Commits taggen, die bereits den `main`-CI-Lauf bestanden haben. Der
Package-Repositories-Workflow laeuft fuer `v*`-Tags, deshalb muss das
`github-pages`-Environment Deployments von `main` und `v*`-Tags erlauben.

Der Tag-Push steuert die Alpha-Veroeffentlichung. Der Package-Repositories-
Workflow erstellt den GitHub-Pre-Release und haengt Linux-Pakete, FreeBSD-
Pakete und Repository-Archiv, `SHA256SUMS` und `unifi-stubd.spdx.json` an.
Die persistenten Release-Artefakte nach Workflow-Ende pruefen:

```sh
gh release view v0.2.4-alpha --json isPrerelease,assets,url
```

## Signierung und Repository-Metadaten

Sobald ein Release-Key konfiguriert ist, einen signierten Tag mit `git tag -s`
erstellen, `SHA256SUMS` vom Release laden und eine abgetrennte Signatur wie
`SHA256SUMS.asc` veroeffentlichen. Beides mit `git tag -v <tag>` und `gpg
--verify SHA256SUMS.asc SHA256SUMS` pruefen. Ohne konfigurierten Key bleibt der
Release Alpha, Pre-Release und explizit unsigniert; keine Hochstufung auf
Stable.

Repository-Beschreibung, Homepage, SPDX-Lizenz und Topics muessen gesetzt und
mit `README.md`, `LICENSE` sowie den Paketmetadaten konsistent sein. Release-
Target ist der getestete Commit auf `main`; Tag-SHA und Release-Target-SHA
muessen uebereinstimmen.

## GitHub-Pages-Paketquellen

Alpha-Paketquellen werden unsigniert ueber GitHub Pages veroeffentlicht:

```text
https://konstruktor1.github.io/unifi-stubd/
```

Zuerst alle Paket-Artefakte bauen, danach die statische Repository-Seite
erzeugen:

```sh
PKG_VERSION=0.2.4-alpha PKG_RELEASE=1 PKG_GOARCH=amd64 make package
PKG_VERSION=0.2.4-alpha PKG_RELEASE=1 PKG_GOARCH=arm64 make package
PKG_VERSION=0.2.4-alpha PKG_RELEASE=1 make package-freebsd-pkg-repos
make package-repos
```

Der Package-Repositories-Workflow veroeffentlicht die Paketquelle ueber GitHub
Pages, wenn ein `v*`-Tag gepusht wird. Fuer einen manuellen Rebuild oder Retry
ohne neuen Release den Workflow von `main` aus starten:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.4-alpha \
  -f package_release=1
```

Wenn `version` bei einem manuellen Lauf fehlt, liest der Workflow `VERSION`.

Die SPDX-JSON-SBOM wird an den GitHub-Release angehaengt und verfaellt daher
nicht mit der siebentaegigen Actions-Artefaktaufbewahrung. Die Release-
Checksummen decken SBOM und jedes hochgeladene Paketartefakt ab.

`make package-repos` schreibt `dist/package-site/` mit APT-, RPM-,
Arch-Linux-, FreeBSD-/OPNsense-Tarball-Pfaden und nativen FreeBSD-`pkg`-Repos,
wenn `dist/freebsd-pkg-repos/repo/` existiert. Der
Package-Repositories-Workflow baut FreeBSD-Tarballs und native FreeBSD-Repos
auf dem self-hosted Runner mit Label `freebsd-pkg-builder`. Der kanonische
Builder nutzt Go-Cross-Compile auf dem FreeBSD-Host plus FreeBSD `pkg`;
`FREEBSD_PKG_BUILD_JAILS` ist optional fuer eigene Builder mit laufenden
`jexec`-Jails. Danach wird die kombinierte statische Seite von Ubuntu deployed.
Alpha-Anleitungen bleiben sichtbar
unsigniert (`trusted=yes`, `gpgcheck=0`, `SigLevel = Never`), bis ein
Projekt-Release-Key existiert.

Die generierte Projektseite verlinkt ausserdem auf Source-Repository, Releases,
Wiki, `CREDITS.md` und eine kurze Karte der genutzten Research-/Quellprojekte.
Diese Karte bleibt nur eine Zusammenfassung; die verbindliche Attribution steht
weiterhin in `CREDITS.md`.

## Paket-Metadaten

Der Default-Maintainer fuer Pakete ist `unifi-stubd maintainers <info@spinas.org>`.
Maintainer-Metadaten fuer oeffentliche Paket-Builds ueberschreiben:

```sh
PKG_MAINTAINER='Name <email@example.com>' PKG_GOARCH=amd64 make package
```

## Host-Konfigurationen

Echte Host-Konfigurationen nicht in diesem Repository ablegen. Nutze dafuer ein
privates Verzeichnis oder ein privates Deployment-Repository. Eine sinnvolle
lokale Struktur ist:

```text
unifi-stubd-host-configs/
  hosts/<host>/current-tmp.yaml
  hosts/<host>/installed-config.yaml
  hosts/<host>/process.txt
  hosts/<host>/status-before.json
  hosts/<host>/status-after.json
```

`current-tmp.yaml` dokumentiert die temporaere Laufzeitkonfiguration.
`installed-config.yaml` ist die paketfertige Konfiguration fuer den installierten
Servicepfad. Das Paket selbst bleibt neutral und aendert keine VLANs,
Firewall-Regeln, Routen oder Controller-Netzwerkdefinitionen.
