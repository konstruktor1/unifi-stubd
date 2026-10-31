# Release-Checkliste

Nutze diese Checkliste vor einem veroeffentlichten Tag oder Paket-Set.

Release-Artefakte sind bewusst neutral. Sie duerfen keine echten
Controller-URLs, Site-IP-Adressen, MAC-Adressen, Clientnamen, Credentials,
Adoption-State-Dateien oder host-spezifische Interface-Mappings enthalten.
Solche Dateien gehoeren in eine private Konfigurationsablage ausserhalb dieses
Repositories und werden nach der Paketinstallation auf den jeweiligen Host
kopiert.

## Vor dem Tag

1. Den vollstaendigen hostseitigen Gate `make verify` erfolgreich ausfuehren,
   danach die erforderlichen Package-Targets ueber ihre konfigurierten Routen.
   `make check` ist der Go-Produktgate innerhalb des vorbereiteten Runners,
   nicht der vollstaendige Release-Gate; [`AGENTS.md`](../../AGENTS.md)
   dokumentiert die Validierungseinstiege. Der vollstaendige Gate enthaelt
   die Policy-Pruefung auf private Lab-Daten.
2. Die erforderlichen `amd64`- und `arm64`-Releaseartefakte bauen.
3. `dist/packages/` auf die Linux-Ausgaben fuer Debian, RPM, Arch Linux und
   `.tar.gz` pruefen.
4. Nach Dependency- oder Tool-Updates `go.mod`, `go.work` und `go.sum`, das
   isolierte Linter-Modul `golangci-lint.mod` und `golangci-lint.sum` sowie das
   npm-Manifest und Lockfile im Projektroot, `package.json` und
   `package-lock.json`, pruefen.
5. Pruefen, dass `VERSION`, der neueste datierte Abschnitt in `CHANGELOG.md`
   und der vorgesehene Tag `v<version>` uebereinstimmen.
6. `CREDITS.md` und `NOTICE.md` aktualisieren, wenn neue Quellen, kopierter
   Code, Pakete oder Tools hinzukommen.
7. Pruefen, dass `packaging/linux/etc/unifi-stubd/config.yaml` und `lab/` nur
   Dokumentationsadressen oder neutrale Defaults enthalten.
8. Repository-Metadaten mit `gh repo view --json
   description,homepageUrl,licenseInfo,repositoryTopics` pruefen und
   Paketname, Lizenz, Maintainer und Projekt-URL mit `packaging/nfpm.yaml`
   vergleichen.
9. Signaturstatus pruefen. Alpha-Releases duerfen nur als Pre-Release mit der
   erzeugten unsignierten `SHA256SUMS` veroeffentlicht werden. Stabile Releases
   verlangen einen signierten Tag und eine abgetrennte Signatur fuer
   `SHA256SUMS`.

FreeBSD-Paketerzeugung ist kein Pre-Tag-Gate. Der manuelle Workflow `FreeBSD
Packaging` erzeugt nur Kompatibilitaetsartefakte; die Veroeffentlichung bleibt
bis zu einer autorisierten Buildroute deaktiviert. Diese Packaging-Pipeline
ist keine Testklassifizierung als Tier 1, Tier 2 oder Tier 3.

## Tag setzen

Semantische Versionstags verwenden:

```sh
git tag -a v0.2.5-alpha -m "unifi-stubd v0.2.5-alpha"
git push origin v0.2.5-alpha
```

Nur Commits taggen, deren exakter Tree `make verify` bestanden hat. Die allgemeine
CI bereitet aktuelle BR-/CT-Quellen und native Bindungen auf GitHub-hosted
Ubuntu-Workern vor und verwendet danach Build-Runner-Container. Privater
Forgejo-Lesezugriff und die Bereitstellung aktueller Provideraenderungen muessen
verfuegbar sein; vor einem neuen Tag muss ein erfolgreicher CI-Lauf beobachtet
werden. Siehe die [offenen Provider-Handoffs](ci-cd.md#offene-provider-handoffs).
Der Self-hosted Runner bleibt ausschliesslich beim manuellen FreeBSD-Packaging.
Der Package-Repositories-Workflow laeuft fuer `v*`-Tags,
deshalb muss das `github-pages`-Environment Deployments von `main` und
`v*`-Tags erlauben.

Der Tag-Push steuert die erste Alpha-Veroeffentlichung. Der Package-
Repositories-Workflow erstellt den GitHub-Pre-Release, haengt Linux-Pakete,
`SHA256SUMS` und `unifi-stubd.spdx.json` an und deployed die Linux-only-Pages-
Seite. Er startet oder erwartet FreeBSD-Packaging nie. Die persistenten
Linux-Release-Artefakte nach Workflow-Ende pruefen:

```sh
gh release view v0.2.5-alpha --json isPrerelease,assets,url
```

## Manuelles FreeBSD-Kompatibilitaetspackaging

`FreeBSD Packaging` ist ausschliesslich manuell und hat weder Trigger-
noch `needs`-Beziehungen zu `CI` oder `Package Repositories`. FreeBSD-Tarballs,
native `pkg`-Repositories und `SHA256SUMS.freebsd` bleiben sieben Tage als
Laufartefakte erhalten. Sein direkter Host-Go-Builder ist ein
External-Machine-Kompatibilitaetspfad; diese Artefakte sind kein konformer
Build-Nachweis:

```sh
gh workflow run freebsd-packaging.yml --ref main
```

Der ausgewaehlte Ref muss `main` oder ein zu `VERSION` passender `v*`-Tag sein.
Der Workflow hat keinen Publish-Input, Release-Upload oder Pages-Deploy.
Diese Artefakte auch nicht manuell veroeffentlichen. Der FreeBSD-Build-Owner
muss zuerst eine autorisierte VM-Runner- oder gepruefte Split-Build-Route fuer
die Kompilierung bereitstellen. Erst danach darf der Release-Owner die
Veroeffentlichung nach Linux-Publikation wieder integrieren. Die neue Route
muss die Pruefungen auf passenden Tag, vollstaendigen Linux-Pre-Release und
unsignierte Pruefsummen erhalten.

## Signierung und Repository-Metadaten

Sobald ein Release-Key konfiguriert ist, einen signierten Tag mit `git tag -s`
erstellen, `SHA256SUMS` vom Release laden und eine abgetrennte Signatur wie
`SHA256SUMS.asc` veroeffentlichen. Wenn FreeBSD-Artefakte vorhanden
sind, `SHA256SUMS.freebsd` ebenfalls separat signieren. Tag und jede
veroeffentlichte Pruefsummensignatur vor der Promotion pruefen. Ohne
konfigurierten Key bleibt der Release Alpha, Pre-Release und explizit
unsigniert; keine Hochstufung auf Stable. Packaging-Automation darf signierte
Pruefsummenartefakte niemals ersetzen oder erweitern.

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
ueber GitHub Pages, wenn ein `v*`-Tag gepusht wird. Solange die FreeBSD-
Veroeffentlichung blockiert bleibt, meldet der FreeBSD-Bereich, dass die
manuellen Artefakte nicht verfuegbar sind. Fuer einen Rebuild oder Retry dieser Linux-only-
Paketquelle ohne neuen Release den Workflow von `main` aus starten:

```sh
gh workflow run package-pages.yml --ref main \
  -f version=0.2.5-alpha \
  -f package_release=1
```

Wenn `version` bei einem manuellen Lauf fehlt, liest der Workflow `VERSION`.

Ein Lauf von Package Repositories ersetzt eine bestehende kombinierte
Pages-Seite durch die Linux-only-Form. Ihre Wiederherstellung verlangt die
oben beschriebene autorisierte FreeBSD-Build- und Veroeffentlichungsroute.

Die SPDX-JSON-SBOM wird an den GitHub-Release angehaengt und verfaellt daher
nicht mit der siebentaegigen Actions-Artefaktaufbewahrung. `SHA256SUMS` deckt
die automatischen Linux- und SBOM-Artefakte ab. `SHA256SUMS.freebsd` begleitet
die temporaeren FreeBSD-Laufartefakte.

`make package-repos` schreibt `dist/package-site/` mit APT-, RPM-,
Arch-Linux-, FreeBSD-/OPNsense-Tarball-Pfaden und nativen FreeBSD-`pkg`-Repos,
wenn `dist/freebsd-pkg-repos/repo/` existiert. Der Default-Modus verlangt die
FreeBSD-Artefakte; der automatische Workflow erlaubt fuer das Linux-first-
Deployment explizit deren Fehlen. Nur der manuelle Workflow `FreeBSD Packaging`
nutzt die isolierte Packaging-Proxy-Rolle und striktes SSH/SCP zu einer
Non-root-FreeBSD-Builder-Rolle. Seine direkte Host-Go-Kompilierung ist nur
Kompatibilitaet und als Release-Evidenz blockiert; FreeBSD `pkg` besitzt
weiterhin den Aufbau der Ziel-Repositories. `FREEBSD_PKG_BUILD_JAILS` ist
optional fuer eigene Builder mit laufenden `jexec`-Jails. Der FreeBSD-Workflow
baut oder deployed keine kombinierte Seite. Alpha-Anleitungen bleiben sichtbar
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
