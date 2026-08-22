# Release-Checkliste

Nutze diese Checkliste vor einem veroeffentlichten Tag oder Paket-Set.

Release-Artefakte sind bewusst neutral. Sie duerfen keine echten
Controller-URLs, Site-IP-Adressen, MAC-Adressen, Clientnamen, Credentials,
Adoption-State-Dateien oder host-spezifische Interface-Mappings enthalten.
Solche Dateien gehoeren in eine private Konfigurationsablage ausserhalb dieses
Repositories und werden nach der Paketinstallation auf den jeweiligen Host
kopiert.

## Vor dem Tag

1. Den vollstaendigen Gate `make code-checker` und die erforderlichen
   Package-Targets ueber den konfigurierten Build Runner ausfuehren;
   [`AGENTS.md`](../../AGENTS.md) besitzt den Aufruf.
2. Die erforderlichen `amd64`- und `arm64`-Releaseartefakte bauen.
3. `dist/packages/` auf die Linux-Ausgaben fuer Debian, RPM, Arch Linux und
   `.tar.gz` pruefen.
4. Nach Dependency- oder Tool-Updates `go.mod`, `go.work` und `go.sum` pruefen.
5. Private Lab-Daten suchen:

   ```sh
   sh scripts/check-policy.sh
   ```

6. Pruefen, dass `VERSION`, der neueste datierte Abschnitt in `CHANGELOG.md`
   und der vorgesehene Tag `v<version>` uebereinstimmen.
7. `CREDITS.md` und `NOTICE.md` aktualisieren, wenn neue Quellen, kopierter
   Code, Pakete oder Tools hinzukommen.
8. Pruefen, dass `packaging/linux/etc/unifi-stubd/config.yaml` und `lab/` nur
   Dokumentationsadressen oder neutrale Defaults enthalten.
9. Repository-Metadaten mit `gh repo view --json
    description,homepageUrl,licenseInfo,repositoryTopics` pruefen und
    Paketname, Lizenz, Maintainer und Projekt-URL mit `packaging/nfpm.yaml`
    vergleichen.
10. Signaturstatus pruefen. Alpha-Releases duerfen nur als Pre-Release mit der
    erzeugten unsignierten `SHA256SUMS` veroeffentlicht werden. Stabile Releases
    verlangen einen signierten Tag und eine abgetrennte Signatur fuer
    `SHA256SUMS`.

FreeBSD-Paketerzeugung ist kein Pre-Tag-Gate. Der manuelle Workflow `FreeBSD
Packaging` darf einen Pre-Release erst nach der Linux-Veroeffentlichung und nur
mit vorhandener autorisierter Buildroute erweitern. Der direkte Builder ist
derzeit als Evidenz blockiert. Diese Packaging-Pipeline ist keine
Testklassifizierung als Tier 1, Tier 2 oder Tier 3.

## Tag setzen

Semantische Versionstags verwenden:

```sh
git tag -a v0.2.5-alpha -m "unifi-stubd v0.2.5-alpha"
git push origin v0.2.5-alpha
```

Nur Commits taggen, deren exakter Tree den vollstaendigen Gate bestanden hat.
Der aktuelle `main`-CI-Lauf kann diese Evidenz bis zu seiner genehmigten
CI-sichtbaren Build-Runner-Route nicht liefern; vor dieser Route darf kein neuer
Tag veroeffentlicht werden. Der
Package-Repositories-Workflow laeuft fuer `v*`-Tags, deshalb muss das
`github-pages`-Environment Deployments von `main` und `v*`-Tags erlauben.

Der Tag-Push steuert die erste Alpha-Veroeffentlichung. Der Package-
Repositories-Workflow erstellt den GitHub-Pre-Release, haengt Linux-Pakete,
`SHA256SUMS` und `unifi-stubd.spdx.json` an und deployed die Linux-only-Pages-
Seite. Er startet oder erwartet FreeBSD-Packaging nie. Die persistenten
Linux-Release-Artefakte nach Workflow-Ende pruefen:

```sh
gh release view v0.2.5-alpha --json isPrerelease,assets,url
```

## Manuelle FreeBSD-Packaging-Erweiterung

`FreeBSD Packaging` ist ausschliesslich manuell und hat weder Trigger-
noch `needs`-Beziehungen zu `CI` oder `Package Repositories`. Der Input
`publish` steht standardmaessig auf `false`; dieser Modus behaelt FreeBSD-
Tarballs und native `pkg`-Repositories sieben Tage. Sein direkter Host-Go-
Builder ist ein External-Machine-Kompatibilitaetspfad; diese Artefakte sind
kein konformer Build-Nachweis. Sie duerfen erst veroeffentlicht werden, wenn
eine autorisierte VM-Runner- oder gepruefte Split-Build-Route den
Kompilierungsschritt besitzt:

```sh
gh workflow run freebsd-packaging.yml --ref main -f publish=false
```

Zur Veroeffentlichung wird der exakte Release-Tag mit expliziter Bestaetigung
gestartet:

```sh
gh workflow run freebsd-packaging.yml --ref <v-tag-mit-freebsd-workflow> -f publish=true
```

Die Veroeffentlichung wird nur akzeptiert, wenn der ausgewaehlte Ref ein
`v*`-Tag ist, der Tag ohne `v` mit `VERSION` uebereinstimmt und bereits ein
passender GitHub-Pre-Release mit vollstaendigem Linux-Paketset, SPDX-SBOM und
`SHA256SUMS` existiert. Der Workflow verweigert ausserdem Aenderungen an einem
Release, der bereits eine signierte `SHA256SUMS` oder
`SHA256SUMS.freebsd` enthaelt.
Auch Legacy-Releases, deren `SHA256SUMS` FreeBSD-Artefakte bereits abdeckt,
werden abgelehnt. Der gewaehlte Tag muss diesen Workflow enthalten; die
Trennung beginnt deshalb mit dem ersten neuen Release-Tag nach dieser
Implementierung.

Ein erfolgreicher Publish-Lauf fuegt diese persistenten Artefakte hinzu und
deployed danach eine kombinierte Linux-/FreeBSD-Pages-Seite:

- `unifi-stubd_<version>-1_freebsd_amd64.tar.gz`
- `unifi-stubd_<version>-1_freebsd_arm64.tar.gz`
- `freebsd-pkg-repos.tar.gz`
- `SHA256SUMS.freebsd`

`SHA256SUMS.freebsd` deckt nur die drei FreeBSD-Artefakte ab. Die automatische
Linux-/SBOM-Pruefsummendatei bleibt unveraendert. Ein erneuter Lauf von
`FreeBSD Packaging` fuer denselben Pre-Release ist idempotent, weil die vier
FreeBSD-Artefakte anhand ihres Namens ersetzt werden.

## Signierung und Repository-Metadaten

Sobald ein Release-Key konfiguriert ist, einen signierten Tag mit `git tag -s`
erstellen, `SHA256SUMS` vom Release laden und eine abgetrennte Signatur wie
`SHA256SUMS.asc` veroeffentlichen. Wenn FreeBSD-Artefakte vorhanden
sind, `SHA256SUMS.freebsd` ebenfalls separat signieren. Tag und jede
veroeffentlichte Pruefsummensignatur vor der Promotion pruefen. Ohne
konfigurierten Key bleibt der Release Alpha, Pre-Release und explizit
unsigniert; keine Hochstufung auf Stable. FreeBSD-Artefakte vor dem Signieren
hinzufuegen, weil der FreeBSD-Packaging-Workflow einen Release mit signierten
Pruefsummenartefakten nicht mehr veraendert.

Repository-Beschreibung, Homepage, SPDX-Lizenz und Topics muessen gesetzt und
mit `README.md`, `LICENSE` sowie den Paketmetadaten konsistent sein. Release-
Target ist der getestete Commit auf `main`; Tag-SHA und Release-Target-SHA
muessen uebereinstimmen.

## GitHub-Pages-Paketquellen

Alpha-Paketquellen werden unsigniert ueber GitHub Pages veroeffentlicht:

```text
https://konstruktor1.github.io/unifi-stubd/
```

Der Package-Repositories-Workflow veroeffentlicht eine Linux-only-Paketquelle
ueber GitHub Pages, wenn ein `v*`-Tag gepusht wird. Bis zur FreeBSD-Packaging-
Veroeffentlichung meldet der FreeBSD-Bereich, dass die manuellen Artefakte nicht
verfuegbar sind. Fuer einen manuellen Rebuild oder Retry dieser Linux-only-
Paketquelle ohne neuen Release den Workflow von `main` aus starten:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.5-alpha \
  -f package_release=1
```

Wenn `version` bei einem manuellen Lauf fehlt, liest der Workflow `VERSION`.

Ein spaeterer Linux-only-Lauf von Package Repositories ersetzt die kombinierte
Pages-Seite wieder durch die Linux-only-Form. `freebsd-packaging.yml` auf dem
passenden Tag erneut mit `publish=true` ausfuehren, um die kombinierte Seite aus
den persistenten Release-Artefakten wiederherzustellen.

Die SPDX-JSON-SBOM wird an den GitHub-Release angehaengt und verfaellt daher
nicht mit der siebentaegigen Actions-Artefaktaufbewahrung. `SHA256SUMS` deckt
die automatischen Linux- und SBOM-Artefakte ab; `SHA256SUMS.freebsd` deckt die
separat veroeffentlichten FreeBSD-Artefakte ab.

`make package-repos` schreibt `dist/package-site/` mit APT-, RPM-,
Arch-Linux-, FreeBSD-/OPNsense-Tarball-Pfaden und nativen FreeBSD-`pkg`-Repos,
wenn `dist/freebsd-pkg-repos/repo/` existiert. Der Default-Modus verlangt die
FreeBSD-Artefakte; der automatische Workflow erlaubt fuer das Linux-first-
Deployment explizit deren Fehlen. Nur der manuelle Workflow `FreeBSD Packaging`
nutzt die isolierte Packaging-Proxy-Rolle und striktes SSH/SCP zu einer
Non-root-FreeBSD-Builder-Rolle. Seine direkte Host-Go-Kompilierung ist nur
Kompatibilitaet und als Release-Evidenz blockiert; FreeBSD `pkg` besitzt
weiterhin den Aufbau der Ziel-Repositories. `FREEBSD_PKG_BUILD_JAILS` ist
optional fuer eigene Builder mit laufenden
`jexec`-Jails. Nach dem FreeBSD-Build wird die kombinierte statische Seite auf
Ubuntu zusammengestellt und deployed. Alpha-Anleitungen bleiben sichtbar
unsigniert (`trusted=yes`, `gpgcheck=0`, `SigLevel = Never`), bis ein
Projekt-Release-Key existiert.

Die generierte Projektseite verlinkt ausserdem auf Source-Repository, Releases,
`CREDITS.md` und eine kurze Karte der genutzten Research-/Quellprojekte.
Diese Karte bleibt nur eine Zusammenfassung; die verbindliche Attribution steht
weiterhin in `CREDITS.md`.

## Paket-Metadaten

Der Default-Maintainer fuer Pakete ist `unifi-stubd maintainers <info@spinas.org>`.
Ein autorisierter Runner-Aufruf darf ihn mit `PKG_MAINTAINER` ueberschreiben.

## Host-Konfigurationen

Echte Host-Konfigurationen nur ausserhalb dieses Repositories in privatem
Deployment-Speicher ablegen. Das Paket aendert keine VLANs, Firewall-Regeln,
Routen oder Controller-Netzwerkdefinitionen.
