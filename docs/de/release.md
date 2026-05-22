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
   package` und `PKG_FREEBSD_GOARCH=amd64 make package-freebsd-tgz`.
4. `dist/packages/` auf Debian-, RPM-, Arch-Linux- und `.tar.gz`-Ausgaben pruefen.
5. Nach Dependency- oder Tool-Updates `go.mod`, `go.work` und `go.sum` pruefen.
6. Private Lab-Daten suchen:

   ```sh
   sh scripts/check-policy.sh
   ```

7. `CHANGELOG.md` aktualisieren.
8. `CREDITS.md` und `NOTICE.md` aktualisieren, wenn neue Quellen, kopierter
   Code, Pakete oder Tools hinzukommen.
9. Pruefen, dass `packaging/linux/etc/unifi-stubd/config.yaml` und `lab/` nur
   Dokumentationsadressen oder neutrale Defaults enthalten.
10. Fuer oeffentliche Releases Pakete oder Checksums mit dem Projekt-Release-Key
   signieren, sobald dieser existiert; stabile Releases nicht als unsignierte
   Artefakte veroeffentlichen.

## Tag setzen

Semantische Versionstags verwenden:

```sh
git tag -a v0.1.0 -m "unifi-stubd v0.1.0"
git push origin v0.1.0
```

Der GitHub-Actions-CI-Workflow baut Pakete fuer `amd64` und `arm64`.

Alpha-Paketsets als Pre-Release veroeffentlichen:

```sh
gh release create v0.1.1-alpha --prerelease \
  --title "unifi-stubd v0.1.1-alpha" \
  --notes-file dist/releases/v0.1.1-alpha/release-notes.md \
  dist/releases/v0.1.1-alpha/*
```

## GitHub-Pages-Paketquellen

Alpha-Paketquellen werden unsigniert ueber GitHub Pages veroeffentlicht:

```text
https://konstruktor1.github.io/unifi-stubd/
```

Zuerst alle Paket-Artefakte bauen, danach die statische Repository-Seite
erzeugen:

```sh
PKG_VERSION=0.1.1-alpha PKG_RELEASE=1 PKG_GOARCH=amd64 make package
PKG_VERSION=0.1.1-alpha PKG_RELEASE=1 PKG_GOARCH=arm64 make package
PKG_VERSION=0.1.1-alpha PKG_RELEASE=1 PKG_FREEBSD_GOARCH=amd64 make package-freebsd-tgz
PKG_VERSION=0.1.1-alpha PKG_RELEASE=1 PKG_FREEBSD_GOARCH=arm64 make package-freebsd-tgz
make package-repos
```

`make package-repos` schreibt `dist/package-site/` mit APT-, RPM-,
Arch-Linux- und FreeBSD-/OPNsense-Tarball-Pfaden. Der Generator benoetigt
`dpkg-scanpackages`, `createrepo_c` und `repo-add`; der GitHub-Pages-Workflow
installiert diese Tools auf Ubuntu vor dem Publish. Alpha-Anleitungen bleiben
sichtbar unsigniert (`trusted=yes`, `gpgcheck=0`, `SigLevel = Never`), bis ein
Projekt-Release-Key existiert.

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
