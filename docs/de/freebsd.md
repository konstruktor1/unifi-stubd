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
- Optionales `wan_health.source: ping` fuer Gateway-Profile. Es fuehrt lokale
  read-only Pings aus und meldet nur WAN-Telemetrie.
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

## Veroeffentlichte Tarballs installieren

Veroeffentlichte Alpha-Tarballs liegen in GitHub Releases und auf der
GitHub-Pages-Paketeseite. Die Pages-Seite bietet stabile URLs fuer
Paketmanager-Doku und direkte FreeBSD-/OPNsense-Downloads:

```text
https://konstruktor1.github.io/unifi-stubd/freebsd/amd64/
https://konstruktor1.github.io/unifi-stubd/freebsd/arm64/
```

Tarball fuer die Host-Architektur laden. `amd64` passt fuer typische
OPNsense-/x86_64-Installationen, `arm64` fuer ARM-FreeBSD-Hosts:

```sh
ARCH=amd64 # oder arm64
fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${ARCH}/unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep "freebsd/${ARCH}/unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz" checksums.txt
sha256 unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz
tar -tzf unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz
```

Die `sha256`-Ausgabe muss zum Eintrag in `checksums.txt` passen. Der Tarball
enthaelt nur neutrale Defaults. Nach dem Entpacken und vor dem Aktivieren des
Diensts `/usr/local/etc/unifi-stubd/config.yaml` pruefen oder ersetzen.

## Service

Tarball auf dem FreeBSD-/OPNsense-Host installieren oder entpacken, dann den
Dienst ueber rc.conf oder OPNsense-Tunables aktivieren:

```sh
sudo tar -xzf unifi-stubd_0.1.1-alpha-1_freebsd_${ARCH}.tar.gz -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sysrc unifi_stubd_enable=YES
service unifi-stubd start
```

Die Paket-Config liegt im Repository unter
`packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml`. Auf FreeBSD
bleibt `operation_mode: stub` der risikoaermste Paketdefault. `port-map` nur in
isolierten Labs verwenden, in denen jeder Profil-Port explizit als `interface`,
`disabled` oder `unmapped` konfiguriert ist.

Fuer OPNsense-Gateway-Labs muessen UniFi-seitiger Interface-Name und
FreeBSD-Interface-Name getrennt bleiben:

Die UXG-Pro-Profilports sind feste Profildaten:

```text
port 1 -> eth0, Profilrolle wan,  1G RJ45
port 2 -> eth1, Profilrolle lan,  1G RJ45
port 3 -> eth2, Profilrolle wan2, 10G SFP+
port 4 -> eth3, Profilrolle lan2, 10G SFP+
```

```yaml
profile: uxgpro
uplink_port: 3
port_overrides:
  - port: 3
    role: wan
    network_group: WAN
    interface: ixl0
    speed: 10000
    media: SFP+
wan_health:
  source: ping
  interval_seconds: 10
  timeout_ms: 1000
  targets:
    - port: 3
      host: 1.1.1.1
```

Bei `uxgpro` ist physischer Port 3 controllerseitig `ifname: eth2`; `ixl0`
wird nur als `source_interface` gemeldet. Die Ping-Quelle aktualisiert
Connectivity- und Latenz-Telemetrie, fuehrt aber keinen UniFi-Speedtest aus,
erkennt keinen Provider und aendert keine OPNsense-Interfaces, Routen,
Firewall-Regeln oder VLANs. Der Ping folgt der Routing-Tabelle des
OPNsense-Hosts; `targets[].port` waehlt nur die WAN-Telemetriezeile, nicht das
ICMP-Quellinterface.

Nach jeder Config-Aenderung das eine YAML-Dokument validieren, den Dienst neu
starten und den lokalen Status pruefen:

```sh
unifi-stubd -validate -config /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
unifi-stubd -status-json
```

Beim SFP-WAN-Beispiel oben muss der Controller WAN auf Port 3 mit
`ifname: eth2`, `source_interface: ixl0`, `uplink: eth2` und bei erfolgreichem
Ping WAN-Health-Latenz/Connectivity sehen. Host-Namen wie `ixl0`, `igb0` oder
`vtnet0` duerfen nicht in controllerseitigen `ifname`-Feldern erscheinen.
Provider- und ISP-Felder bleiben leer, weil automatische Provider-Erkennung
nicht implementiert ist.

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
