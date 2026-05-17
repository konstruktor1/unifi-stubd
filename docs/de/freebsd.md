# FreeBSD und OPNsense

FreeBSD-Support ist vorerst Stub-only. Der Dienst kann fuer FreeBSD gebaut und
mit FreeBSD-Pfaden paketiert werden, aber passive Linux-Observation ist dort
noch nicht implementiert.

Im FreeBSD-Paket unterstuetzt:

- `operation_mode: stub`
- UniFi Discovery und Inform-Traffic
- Eingebauter SSH-Shim fuer Advanced Adoption
- Statische Profil-Payloads, Port-Overrides, Uplink-Port-Override und
  konfigurierte Uplink-Neighbor-Eintraege
- rc.d-Service-Artefakt fuer FreeBSD- und OPNsense-aehnliche Systeme

Noch nicht auf FreeBSD unterstuetzt:

- `operation_mode: observe`
- Linux-Bridge-FDB-Import
- `/sys/class/net`-Counter und Speed-Erkennung
- `operation_mode: macvlan`

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
`operation_mode: stub` verwenden, bis ein nativer Observation-Adapter existiert.
