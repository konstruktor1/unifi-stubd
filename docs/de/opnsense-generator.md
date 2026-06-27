# OPNsense-API-Generator How-to

`unifi-stubd-opnsense` ist ein Companion-Generator. Er laeuft nicht im
`unifi-stubd`-Daemon und aendert keine OPNsense-Interfaces, Routen,
Firewall-Regeln oder VLANs. Er liest OPNsense ueber read-only API-Aufrufe und
erstellt eine normale `unifi-stubd`-YAML-Datei zur Pruefung.

Die aktuellen Pakete installieren `unifi-stubd`. Bis Packaging fuer das
Companion-Tool ergaenzt ist, `unifi-stubd-opnsense` separat bauen und auf den
OPNsense-Host kopieren, wenn der Generator dort laufen soll.
Fuer Feldverhalten, Merge-Regeln, Endpoints und Troubleshooting siehe die
[OPNsense-API-Generator Referenz](opnsense-generator-reference.md).

## 1. OPNsense-API-Key erstellen

In der OPNsense-WebGUI:

1. `System > Access > Users` oeffnen.
2. Den User fuer diese Integration auswaehlen oder einen dedizierten
   read-only Lab-User erstellen.
3. Im API-Key-Bereich des Users einen neuen Key erzeugen.
4. Die generierte Key/Secret-Datei einmalig herunterladen und privat halten.

OPNsense-API-Keys sind Key/Secret-Paare. Der Key ist der HTTP-Basic-Auth-
Username, das Secret ist das Passwort.
Die OPNsense-Dokumentation beschreibt denselben API-Key-Ablauf unter
[Use the API](https://docs.opnsense.org/development/how-tos/api.html) und
[Local Users & Groups](https://docs.opnsense.org/manual/how-tos/user-local.html).

## 2. Companion-Binary auf OPNsense installieren

Das Binary auf einem Entwicklungsrechner aus diesem Repository bauen:

```sh
GOOS=freebsd GOARCH=amd64 go build \
  -o dist/unifi-stubd-opnsense \
  ./cmd/unifi-stubd-opnsense
```

Fuer einen ARM-FreeBSD-/OPNsense-Host `GOARCH=arm64` verwenden. Danach das
Binary kopieren:

```sh
scp dist/unifi-stubd-opnsense root@opnsense.example:/usr/local/bin/
ssh root@opnsense.example chmod 0755 /usr/local/bin/unifi-stubd-opnsense
```

Auf der OPNsense-Shell pruefen:

```sh
/usr/local/bin/unifi-stubd-opnsense -h
```

## 3. API-Credentials auf OPNsense ablegen

Auf der OPNsense-Shell private Dateien fuer API-Key und Secret erstellen:

```sh
mkdir -p /usr/local/etc/unifi-stubd
chmod 700 /usr/local/etc/unifi-stubd
umask 077
ee /usr/local/etc/unifi-stubd/opnsense-api-key
ee /usr/local/etc/unifi-stubd/opnsense-api-secret
chmod 600 /usr/local/etc/unifi-stubd/opnsense-api-key
chmod 600 /usr/local/etc/unifi-stubd/opnsense-api-secret
```

In `opnsense-api-key` nur den rohen Key einfuegen und in
`opnsense-api-secret` nur das rohe Secret. Key und Secret nicht in die
Source-YAML-Datei schreiben.

## 4. OPNsense-Interfaces bestimmen

In der OPNsense-UI `Interfaces > Overview` verwenden oder auf der Shell:

```sh
ifconfig -l
ifconfig ixl0
ifconfig vtnet0
```

Bei einem UXG-Pro-foermigen Lab-Profil sind die dargestellten UniFi-Ports feste
Profildaten:

```text
port 1 -> eth0, Profilrolle wan,  1G RJ45
port 2 -> eth1, Profilrolle lan,  1G RJ45
port 3 -> eth2, Profilrolle wan2, 10G SFP+
port 4 -> eth3, Profilrolle lan2, 10G SFP+
```

Wenn echtes OPNsense-WAN `ixl0` ist und als UniFi-Port 3 erscheinen soll, dann
`port: 3` auf `interface: ixl0` mappen. Das controllerseitige `ifname` bleibt
`eth2`; `ixl0` wird als `source_interface` generiert.

## 5. OPNsense-Source-Datei erstellen

Auf OPNsense `/usr/local/etc/unifi-stubd/opnsense-source.yaml` erstellen:

```sh
ee /usr/local/etc/unifi-stubd/opnsense-source.yaml
chmod 600 /usr/local/etc/unifi-stubd/opnsense-source.yaml
```

Beispiel:

```yaml
base_url: https://127.0.0.1
api_key_file: /usr/local/etc/unifi-stubd/opnsense-api-key
api_secret_file: /usr/local/etc/unifi-stubd/opnsense-api-secret
api_key_env: ""
api_secret_env: ""
ca_file: ""
insecure_skip_verify: false
timeout_ms: 2000
uplink_port: 3
gateway_status: true
interfaces:
  - port: 3
    interface: ixl0
    name: WAN SFP+
    role: wan
    network_group: WAN
    network_name: opnsense_wan
    vlan: 3
  - port: 4
    interface: vtnet0
    name: LAN SFP+ to server-lan1
    role: lan
    network_group: LAN
    network_name: opnsense_lan
    vlan: 1
wan_health:
  source: static
  interval_seconds: 10
  timeout_ms: 1000
  targets: []
```

`insecure_skip_verify: true` nur fuer einen ausdruecklichen Lab-Endpunkt mit
self-signed Zertifikat verwenden, wenn die Validierung per `ca_file` nicht
moeglich ist.

## 6. Lokal ohne API-Aufruf validieren

```sh
/usr/local/bin/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml \
  -validate
```

Das prueft Basisconfig, Source-YAML und Credential-Laden. OPNsense wird dabei
nicht per API kontaktiert.

## 7. Config generieren und pruefen

Die generierte Config zuerst in eine temporaere Datei schreiben:

```sh
/usr/local/bin/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml \
  > /tmp/unifi-stubd.generated.yaml
```

Die generierte `unifi-stubd`-Config validieren:

```sh
/usr/local/bin/unifi-stubd \
  -validate \
  -config /tmp/unifi-stubd.generated.yaml
```

Die generierten Port-Overrides vor der Uebernahme ansehen:

```sh
grep -n "port_overrides" /tmp/unifi-stubd.generated.yaml
grep -n "source_interface\\|interface: ixl0\\|role: wan" /tmp/unifi-stubd.generated.yaml
```

## 8. Generierte Config uebernehmen

Erst ein Backup behalten, dann die Service-Config nach Review ersetzen:

```sh
cp -p /usr/local/etc/unifi-stubd/config.yaml \
  /usr/local/etc/unifi-stubd/config.yaml.before-opnsense
install -m 0600 /tmp/unifi-stubd.generated.yaml \
  /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
/usr/local/bin/unifi-stubd -status-json
```

Der Generator ist kein Live-Sync-Dienst. Wenn sich OPNsense-Interface-
Zuweisungen aendern, Generator erneut laufen lassen und das generierte YAML vor
dem Ersetzen der Daemon-Config wieder pruefen.

## Variante vom Entwicklungsrechner

Der Generator kann auch von einem Entwicklungsrechner laufen. Dann die aktuelle
Stub-Config von OPNsense kopieren, API-Source und Credential-Dateien lokal auf
dem Entwicklungsrechner halten und `base_url` auf die OPNsense-Management-
Adresse setzen:

```sh
scp root@opnsense.example:/usr/local/etc/unifi-stubd/config.yaml ./config.opnsense.yaml
go run ./cmd/unifi-stubd-opnsense \
  -config ./config.opnsense.yaml \
  -source ./opnsense-source.yaml \
  > generated.yaml
scp generated.yaml root@opnsense.example:/tmp/unifi-stubd.generated.yaml
```

Danach `/tmp/unifi-stubd.generated.yaml` auf OPNsense wie oben validieren und
uebernehmen.
