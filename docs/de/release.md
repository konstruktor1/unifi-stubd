# Release-Checkliste

Nutze diese Checkliste vor einem veroeffentlichten Tag oder Paket-Set.

## Vor dem Tag

1. `make check` ausfuehren.
2. `make package` ausfuehren.
3. `dist/packages/` auf Debian-, RPM-, Arch-Linux- und `.tar.gz`-Ausgaben pruefen.
4. Nach Dependency- oder Tool-Updates `go.mod`, `go.work` und `go.sum` pruefen.
5. Private Lab-Daten suchen:

   ```sh
   sh scripts/check-policy.sh
   ```

6. `CHANGELOG.md` aktualisieren.
7. `CREDITS.md` und `NOTICE.md` aktualisieren, wenn neue Quellen, kopierter
   Code, Pakete oder Tools hinzukommen.
8. Pruefen, dass `packaging/linux/etc/unifi-stubd/config.yaml` und `lab/` nur
   Dokumentationsadressen oder neutrale Defaults enthalten.

## Tag setzen

Semantische Versionstags verwenden:

```sh
git tag -a v0.1.0 -m "unifi-stubd v0.1.0"
git push origin v0.1.0
```

Der GitHub-Actions-CI-Workflow baut Pakete fuer `amd64` und `arm64`.

## Paket-Metadaten

Der Default-Maintainer fuer Pakete ist `unifi-stubd maintainers <info@spinas.org>`.
Maintainer-Metadaten fuer oeffentliche Paket-Builds ueberschreiben:

```sh
PKG_MAINTAINER='Name <email@example.com>' PKG_GOARCH=amd64 make package
```
