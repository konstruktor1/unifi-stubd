# FreeBSD und OPNsense

FreeBSD-Support ist eine experimentelle, konservative Paketbasis ausserhalb
des abgenommenen Produktschnitts. Der Dienst kann fuer FreeBSD gebaut und mit
FreeBSD-Pfaden paketiert werden. Stub-Modus ist implementiert; erste read-only
Observation-Helfer existieren fuer Bridge-FDB-Parsing und explizite Port-Map-
Interface-Metadaten.

Im experimentellen FreeBSD-Paket implementiert:

- `operation_mode: stub`
- UniFi Discovery und Inform-Traffic
- Optionaler SSH-Shim fuer Advanced Adoption, standardmaessig mit
  `ssh_listen: ""` deaktiviert
- Statische Profil-Payloads, Port-Overrides, Uplink-Port-Override und
  konfigurierte Uplink-Neighbor-Eintraege
- Optionales `wan_health.source: ping` fuer Gateway-Profile. Es fuehrt lokale
  read-only Pings aus und meldet nur lokalen WAN-Diagnosestatus.
- `operation_mode: port-map` fuer explizite read-only Portzustaende. Interface-
  Metadaten werden von bestehenden OS-Interfaces kopiert, wenn verfuegbar.
- FreeBSD-Bridge-FDB-Parsing fuer `ifconfig <bridge> addr`
  mit derselben Bridge-Member-Klassifikation wie unter Linux: `bridge0` ist
  Backplane-Metadatum, konfigurierte oder einzelne physisch wirkende Member
  sind Uplink-Kandidaten, und `tap*`-/`epair*`-/`vnet*`-Member sind
  Access-Ports.
- `lldp_source: lldpd`, wenn `lldpcli` installiert und erreichbar ist
- rc.d-Service-Artefakt fuer FreeBSD- und OPNsense-aehnliche Systeme

Noch nicht auf FreeBSD unterstuetzt:

- volle `operation_mode: bridge-observe`-Paritaet mit Linux
- `/sys/class/net`-Counter und Speed-Erkennung
- Linux-`/proc`-Support
- `operation_mode: macvlan`
- native `SIOCGIFMEDIA`-/Event-Subscription-Pfade

## Build

Die Make-Targets `build-freebsd` und `package-freebsd-tgz` laufen ueber den
konfigurierten Build Runner. [`AGENTS.md`](../../AGENTS.md) besitzt den Aufruf;
Host-Go ist kein Fallback.

`package-freebsd-pkg-repos` delegiert die Kompilierung derzeit an einen
direkten FreeBSD-Host-Go-Builder ohne Provider-Consumer-Attribution. Der Pfad
bleibt External-Machine-Kompatibilitaet, kein Build-Runner-Validierungs- oder
Release-Nachweis. Agentengesteuerte Paketveroeffentlichung ist blockiert, bis
eine autorisierte VM-Runner-Route oder gepruefte Split-Build-Route diesen
Kompilierungsschritt ersetzt.

Das Tarball-Target verwendet per Default `amd64`, passend zu typischen
OPNsense-Installationen. Fuer einen ARM-FreeBSD-Host
`PKG_FREEBSD_GOARCH=arm64` setzen. Native `pkg`-Repositories enthalten
`FreeBSD:14`- und `FreeBSD:15`-Builds fuer `amd64`, `aarch64` und `armv7`.
Der Tarball landet unter `dist/packages/`.

Das Kompatibilitaets-Target uebertraegt den Source-Tree auf den konfigurierten
FreeBSD-Builder, baut dort jede konfigurierte FreeBSD-ABI und
schreibt die nativen `pkg`-Repositories plus die `amd64`- und
`arm64`-Tarballs. `FREEBSD_PKG_BUILD_JAILS` nur fuer eigene Builder mit
laufenden `jexec`-Jails setzen. Der aktuelle Kompatibilitaets-Poudriere-
Buildhost baut die FreeBSD-Binaries auf dem FreeBSD-Host mit Go 1.27 oder neuer
(`go` oder `go127`) und erstellt danach ABI-spezifische `pkg`-Repositories mit dem
FreeBSD-`pkg`-Tool.

GitHub-Actions-Veroeffentlichung, Runner-Grenzen, Aufbewahrung und Tag-Regeln
stehen in der
[Release-Checkliste](release.md#manuelles-freebsd-kompatibilitaetspackaging).

Der Tarball enthaelt dieses Layout:

```text
/usr/local/bin/unifi-stubd
/usr/local/etc/unifi-stubd/config.yaml
/usr/local/etc/unifi-stubd/ssh_host_rsa_key
/usr/local/etc/rc.d/unifi-stubd
/var/db/unifi-stubd/adoption.env
/var/db/unifi-stubd/status.json
/var/log/unifi-stubd.log
```

## Veroeffentlichte Tarballs installieren

Der aktuelle manuelle Workflow erzeugt nur Kompatibilitaets-Laufartefakte und
veroeffentlicht keine FreeBSD-Pakete. Diese Download-Schritte nur verwenden,
wenn der gewaehlte Release bereits den passenden veroeffentlichten Tarball und
die Pruefsummen enthaelt.

Tarball fuer die Host-Architektur laden. `amd64` passt fuer typische
OPNsense-/x86_64-Installationen, `arm64` fuer ARM-FreeBSD-Hosts:

```sh
ARCH=amd64 # oder arm64
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${ARCH}/unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep "freebsd/${ARCH}/unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz" checksums.txt
sha256 unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
tar -tzf unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz
```

Die `sha256`-Ausgabe muss zum Eintrag in `checksums.txt` passen. Der Tarball
enthaelt eine nicht beobachtende Startkonfiguration, die `US8` explizit
auswaehlt. Nach dem Entpacken und vor dem Aktivieren des Diensts
`/usr/local/etc/unifi-stubd/config.yaml` pruefen oder ersetzen.

## Veroeffentlichte pkg-Repositories installieren

Diese Schritte nur fuer ein bestehendes veroeffentlichtes natives
Alpha-`pkg`-Repository verwenden. Die Repositories sind nach FreeBSD-ABI
gruppiert. Zuerst die ABI des Zielsystems pruefen:

```sh
pkg config ABI
```

Danach den passenden Repository-Pfad konfigurieren, zum Beispiel fuer
`FreeBSD:14:amd64`:

```sh
sudo mkdir -p /usr/local/etc/pkg/repos
sudo tee /usr/local/etc/pkg/repos/unifi-stubd.conf >/dev/null <<'PKGCONF'
unifi-stubd: {
  url: "https://konstruktor1.github.io/unifi-stubd/freebsd/pkg/FreeBSD:14:amd64",
  enabled: yes,
  signature_type: none
}
PKGCONF
sudo pkg update -r unifi-stubd
sudo pkg install unifi-stubd
```

Diese Repositories sind unsignierte Alpha-Artefakte. Nur in isolierten Lab-
oder Management-Netzen verwenden. Das native `pkg`-Paket markiert
`/usr/local/etc/unifi-stubd/config.yaml` als Paket-Config, damit Upgrades lokale
Aenderungen erhalten und bei nicht automatisch mergebaren neuen Defaults eine
`config.yaml.pkgnew` schreiben. Der Paket-`post-install`-Hook laeuft auch nach
Upgrades und startet eine konservative Config-Migration fuer bekannte
Legacy-Aliasse wie `controller`, `operation_mode: observe` und
Top-Level-`observe_bridge`/`observe_interface`. Der Migrator validiert das
umgeschriebene YAML vor dem Ersetzen und schreibt daneben ein zeitgestempeltes
`.bak.*`-Backup. Widerspruechliche Werte werden gemeldet, brechen das
Paket-Upgrade aber nicht ab.

Migration ohne Schreibzugriff pruefen:

```sh
unifi-stubd -config /usr/local/etc/unifi-stubd/config.yaml -config-migrate-dry-run
```

Migration manuell ausfuehren:

```sh
sudo unifi-stubd -config /usr/local/etc/unifi-stubd/config.yaml -config-migrate
```

## Service

Tarball auf dem FreeBSD-/OPNsense-Host installieren oder entpacken, dann den
Dienst ueber rc.conf oder OPNsense-Tunables aktivieren:

```sh
sudo tar -xzf unifi-stubd_0.2.5-alpha-1_freebsd_${ARCH}.tar.gz -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sysrc unifi_stubd_enable=YES
service unifi-stubd start
```

Die Paket-Config liegt im Repository unter
`packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml`. Auf FreeBSD
bleibt `operation_mode: stub` der risikoaermste Paketdefault. `port-map` nur in
isolierten Labs verwenden, in denen jeder Profil-Port explizit als `interface`,
`disabled` oder `unmapped` konfiguriert ist.

Fuer Source-Mapping, Generierung und Review das
[OPNsense-Generator-How-to](opnsense-generator.md) verwenden.

Nach jeder Config-Aenderung das eine YAML-Dokument validieren, den Dienst neu
starten und den lokalen Status pruefen:

```sh
unifi-stubd -validate -config /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
unifi-stubd -status-json
```
