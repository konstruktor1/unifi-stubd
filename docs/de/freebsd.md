# FreeBSD und OPNsense

FreeBSD-Support bleibt konservativ. Der Dienst kann fuer FreeBSD gebaut und mit
FreeBSD-Pfaden paketiert werden. Stub-Modus ist die unterstuetzte Basis;
erste read-only Observation-Helfer existieren fuer Bridge-FDB-Parsing und
explizite Port-Map-Interface-Metadaten.

Im FreeBSD-Paket unterstuetzt:

- `operation_mode: stub`
- UniFi Discovery und Inform-Traffic
- Eingebauter SSH-Shim fuer Advanced Adoption
- Statische Profil-Payloads, Port-Overrides, Uplink-Port-Override und
  konfigurierte Uplink-Neighbor-Eintraege
- `operation_mode: port-map` fuer explizite read-only Portzustaende. Interface-
  Metadaten werden von bestehenden OS-Interfaces kopiert, wenn verfuegbar.
- FreeBSD-Bridge-FDB-Parsing fuer `ifconfig <bridge> addr`
  mit derselben Bridge-Member-Klassifikation wie unter Linux: `bridge0` ist
  Backplane-Metadatum, konfigurierte oder einzelne physisch wirkende Member
  sind Uplink-Kandidaten, und `tap*`-/`epair*`-/`vnet*`-Member sind
  Access-Ports.
- `log_source: syslog` fuer read-only Status-/Log-Metadaten aus einer
  konfigurierten Datei, standardmaessig `/var/log/messages`; OPNsense nutzt
  haeufig `/var/log/system/latest.log`, das per Default root-only ist
- `lldp_source: lldpd`, wenn `lldpcli` installiert und erreichbar ist
- optionale D-Bus-Verfuegbarkeitspruefung; D-Bus ist nicht erforderlich
- rc.d-Service-Artefakt fuer FreeBSD- und OPNsense-aehnliche Systeme

Noch nicht auf FreeBSD unterstuetzt:

- volle `operation_mode: bridge-observe`-Paritaet mit Linux
- `/sys/class/net`-Counter und Speed-Erkennung
- Linux-`/proc`-Support
- `operation_mode: macvlan`
- native `SIOCGIFMEDIA`-/Event-Subscription-Pfade

## Build

FreeBSD-Binary cross-builden:

```sh
make build-freebsd
```

FreeBSD-/OPNsense-Stub-Paket als Tarball bauen:

```sh
make package-freebsd-tgz
```

Die FreeBSD-Targets verwenden per Default `amd64`, passend zu typischen
OPNsense-Installationen. Fuer einen ARM-FreeBSD-Host
`PKG_FREEBSD_GOARCH=arm64` setzen. Der Tarball landet unter `dist/packages/`
und enthaelt dieses Layout:

```text
/usr/local/bin/unifi-stubd
/usr/local/etc/unifi-stubd/config.yaml
/usr/local/etc/unifi-stubd/ssh_host_rsa_key
/usr/local/etc/rc.d/unifi-stubd
/var/db/unifi-stubd/adoption.env
/var/db/unifi-stubd/status.json
/var/log/unifi-stubd.log
```

## Service

Tarball auf dem FreeBSD-/OPNsense-Host installieren oder entpacken, dann den
Dienst ueber rc.conf oder OPNsense-Tunables aktivieren:

```sh
sysrc unifi_stubd_enable=YES
service unifi-stubd start
```

Die Paket-Config liegt im Repository unter
`packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml`. Auf FreeBSD
bleibt `operation_mode: stub` der risikoaermste Paketdefault. `port-map` nur in
isolierten Labs verwenden, in denen jeder Profil-Port explizit als `interface`,
`disabled` oder `unmapped` konfiguriert ist.

Fuer OPNsense-Logmetadaten `log_source: off` lassen, solange der Service-User
die ausgewaehlte Datei nicht lesen kann. Ein read-only Runtime-Test auf
OPNsense 26.1/FreeBSD 14.3 validierte `syslog_path:
/var/log/system/latest.log`, wenn das Binary als root lief. Der Default
`/var/log/messages` bleibt der generische FreeBSD-Pfad.

## Native Helfer

Go bleibt die Implementierungssprache. C++ oder native Helper-Binaries werden
nicht eingesetzt. Falls FreeBSD spaeter Daten braucht, die Go nicht sauber lesen
kann, sollte ein kleiner reviewbarer Helper erst nach nachgewiesenem OS-Limit
erganzt werden. Der naechste naheliegende FreeBSD-Ausbau ist ein
`x/sys/unix`-Pfad fuer `SIOCGIFMEDIA`.
