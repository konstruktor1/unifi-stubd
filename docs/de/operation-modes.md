# Betriebsmodi

`unifi-stubd` zielt zuerst auf Linux-Lab-Hosts, also Proxmox, Alpine und
UTM-Linux-VMs. Der abgenommene Produktscope ist ein `US8` auf einer
Linux-Bridge. FreeBSD/OPNsense und jedes andere Profil bleiben experimentelle
Pfade, die eine eigene Produktentscheidung und Evidence brauchen.

## Aktueller Produktschnitt

Der aktuelle controller-datenpfadqualifizierte Weg ist:

- Profil `us8`;
- `operation_mode: bridge-observe` auf Linux;
- eine explizit benannte Bridge und echte lokale Metrikquelle;
- vor Adoption sichtbar, adoptierbar, verbunden und mit echten Link-, Speed-,
  Counter- und Ratenwerten auf aktuellem Stable und Release Candidate.

Direkte UI-Beobachtung, installierter Dienst und Betreiberpfad bleiben offen.
Die Profile unten beschreiben implementierte experimentelle Identitaeten; sie
erweitern den abgenommenen Produktscope nicht.

`UGW3` ist als experimentelles Gateway-Identitaetsprofil verfuegbar. Es meldet
das Legacy-UniFi-Security-Gateway-Modell und drei 1G-Ports, bleibt aber
Stub-only und emuliert noch keine Router-Dienste.

`UXG` ist ueber das experimentelle `uxg-lite`-Gateway-Identitaetsprofil
verfuegbar. Es meldet ein zweipoliges Gateway-Lite-Layout mit `LAN` und `WAN`,
bleibt aber Stub-only und emuliert noch keine Router-Dienste.

`UXGPRO` ist als experimentelles 10G-Gateway-Identitaetsprofil verfuegbar. Es
behaelt die originale Gateway-artige Zuweisung: `WAN` auf dem primaeren
1G-RJ45-WAN, `LAN` auf dem 1G-RJ45-LAN, `WAN2` auf dem sekundaeren
10G-SFP+-WAN und `LAN2` auf dem 10G-SFP+-LAN. Wenn ein Lab die aktive
Internet-Seite auf SFP+ legt, gehoert das in `uplink_port` und
`port_overrides`.

`UCGF` ist ueber das experimentelle `ucg-fiber`-Cloud-Gateway-Fiber-
Identitaetsprofil verfuegbar. Es meldet Device-Type `udm` mit vier
2.5G-RJ45-LAN-Ports, einem 10G-RJ45-Port `WAN2`, einem 10G-SFP+-Port `WAN`
und einem 10G-SFP+-LAN-Port. Es ist Stub-only und startet weder UniFi OS noch
gebuendelte Controller-Anwendungen.

## YAML-Konfigurationsnotiz

Die Runtime-Config ist ein einzelnes, vom Operator gepflegtes YAML-Dokument.
Block-YAML ist empfohlen. JSON-kompatibles YAML ist nur gueltig, wenn die ganze
Config ein vollstaendiges JSON-Objekt ist; nach einer schliessenden `}` darf
kein Block-YAML wie `wan_health:` angehaengt werden. Nach jeder Aenderung
validieren:

```sh
unifi-stubd -validate -config /etc/unifi-stubd/config.yaml
unifi-stubd -status-json
```

Auf FreeBSD/OPNsense gilt derselbe Ablauf mit
`/usr/local/etc/unifi-stubd/config.yaml`. Controller-`setparam`-/`system_cfg`-
Daten sind keine Runtime-Config-Quelle und duerfen nicht als Autoritaet fuer
Host-Netzwerk verstanden werden.

## Modi

### `stub`

Default-Modus. Der Dienst sendet Discovery- und Inform-Payloads nur aus den
Profildaten. Er liest keinen Host-Bridge-State und aendert kein Host-Netzwerk.
Das ist die verfuegbare FreeBSD-/OPNsense-Basis, kein abgenommener Produktschnitt.

### `bridge-observe`

Read-only Linux-Observation-Modus. Die Fake-Device-Identitaet bleibt gleich,
aber passive Host-Daten werden genutzt, wenn sie verfuegbar sind:

- `/sys/class/net/<interface>/statistics/*` fuer Port-Zaehler
- `/sys/class/net/<interface>/speed` fuer Uplink-Speed
- Interface-State, Speed und Zaehler pro Bridge-Member fuer gemappte Ports
- optional `/proc/net/dev`-Zaehler, wenn `proc_source: procfs` aktiv ist
- `bridge fdb show br <bridge>` fuer gelernte MAC-Table-Eintraege
- optionale passive LLDP-Nachbarn, wenn `lldp_source: lldpd` aktiv ist

FDB-Zeilen werden nach Bridge-Member gruppiert und vor dem Port-Mapping
klassifiziert. `bridge_observe.uplink_interface` wird auf den aktuellen
UniFi-Upstream-Link gelegt. Das Bridge-Device selbst, zum Beispiel `vmbr0`
oder `bridge0`, gilt als
Backplane-Metadatum und belegt keinen UniFi-Port. Proxmox-/VM-Member wie
`tap*`, `veth*`, `fwpr*`, `fwln*`, `fwbr*` und FreeBSD-artige
`epair*`-/`vnet*`-Member sind Access-/Downstream-Ports mit ihren gelernten
MAC-Tabellen. Wenn kein Uplink-Interface konfiguriert ist und genau ein
physisch wirkendes Bridge-Member existiert, wird es als Uplink-Kandidat
behandelt; sonst werden unbekannte Member als normale Ports gemappt.
`bridge_observe.member_port_map` kann ein Member fest auf einen UniFi-Port
pinnen, wenn die deterministische Sortierung nicht reicht.
`bridge_observe.ignored_members` schliesst lokale Bridge-Member komplett aus;
das ist fuer TAP-/epair-Seiten gedacht, die bereits ueber einen expliziten
physischen oder Uplink-Port dargestellt werden.

`observe` bleibt als Migrationsalias gueltig und wird intern auf
`bridge-observe` normalisiert. Bestehende `observe_interface`- und
`observe_bridge`-Konfigurationen werden weiter als Fallback fuer
`bridge_observe.uplink_interface` und `bridge_observe.bridge` gelesen.

```yaml
profile: us8
operation_mode: bridge-observe
bridge_observe:
  bridge: vmbr0
  uplink_interface: eno1
  ignored_members:
    - tap10000i0
  member_port_map:
    - member: tap101i0
      port: 2
```

Auf einem Proxmox-Host kann damit `vmbr0` als dargestellter Switch dienen. Das
Uplink-Member, zum Beispiel `eno1`, liefert Uplink-Zaehler und Link-Speed.
MAC-Eintraege, die auf diesem Uplink gelernt werden, gelten als Remote-Geraete
hinter dem echten Nachbarswitch. Der Dienst erfasst diese Remote-MACs zuerst
und schliesst sie aus allen lokalen Access-Port-MAC-Tabellen aus, auch wenn die
Bridge dieselbe MAC spaeter nochmals in einer anderen FDB-Zeile meldet. VM-
oder Container-Teilnehmer wie `tap101i0` und `veth200i0` werden aus der Bridge
FDB gelernt; ihre MAC-Adressen landen in `port_table[].mac_table`, waehrend
ihre eigene Interface-Speed und ihre Zaehler fuer die gemappten Access-Ports
genutzt werden. Ports ohne gemapptes Bridge-Member werden als getrennt gemeldet
und behalten keinen synthetischen Profil-Link-State. Der Dienst
liest sysfs und FDB pro Heartbeat neu. Netlink-Events werden in dieser Wave noch
nicht abonniert.

Das Profil waehlt den Uplink-Port per Default. `uplink_port` kann auf eine
positive Portnummer gesetzt werden, um die Uplink-Markierung auf einen
bestimmten physischen Port zu legen, ohne dessen Profil-Speed oder Medium zu
aendern. Beispiel: `uplink_port: 1` setzt beim `usaggpro` den Uplink auf einen
10G-SFP+-Port statt auf die Default-25G-SFP28-Uplinkgruppe.

Bei Switch-Profilen mit dedizierten SFP-/SFP+-Uplink-Gruppen bleiben diese
Ports die profilseitig definierten Uplink-Cages. Separat dazu setzt
`bridge-observe` ein explizit konfiguriertes physisches `uplink_interface`
automatisch auf den letzten normalen GE-Port, wenn `uplink_port` nicht gesetzt
ist. Damit erscheint die echte Host-Verbindung als aktiver Kupfer-Upstream,
ohne die SFP-/SFP+-Cages umzudefinieren. Einfache Switch-Profile ohne
dedizierte Uplink-Gruppe behalten ihren Profil-Default.

Wenn der dargestellte Host ueber einen SFP-/SFP+-Link angeschlossen ist, sollte
`uplink_port` explizit auf den passenden Profil-Port gesetzt werden, statt die
GE-Fallback-Logik zu nutzen. Beispiel: Ein 48-Port-Switch-Profil mit SFP+-
Cages kann den Bridge-Uplink auf `uplink_port: 49` melden. Port 48 bleibt dann
getrennt und der aktive Uplink behaelt das SFP+-Medium.

Die Topologie-Richtung wird unabhaengig von der FDB-Sichtbarkeit aufgeloest. Der
aktive `uplink_port` ist upstream; direkte Infrastruktur an anderen Ports ist
downstream. Konfigurierte Nachbarn gewinnen vor LLDP und reale Bridge- oder
NIC-MACs bleiben gueltige Geraete- und Port-Identitaeten.

Bei Proxmox kann die Bridge selbst auch die Management-IP des Hosts tragen. Das
ist fuer Proxmox normal, entspricht aber nicht exakt einem Hardware-UniFi-
Switch: Die Management-Identitaet des Stubs repraesentiert dann die Host-Bridge-
IP, nicht ein isoliertes Switch-Management-Interface. Ein dediziertes Management-
VLAN, macvlan/ipvlan oder eine separate Test-IP bildet einen physischen Switch
sauberer ab.

Profile beschreiben das echte Hardware-Layout: Modell, Portanzahl,
Speed-/Mediengruppen, Default-Portnamen und Default-Gateway-Rollen.
`port_overrides` setzt danach lab-spezifische Zuweisungen und einzelne
Portzustaende:

Externe Profile koennen mit `profile_file` oder `profile_dir` eingelesen werden.
Behandle sie als Lab-Daten, bis sie gegen eine konkrete UniFi-Network-Version
validiert wurden. `-profile-template`, `-profile-validate`, `-profile-export`
und `-validate` erzeugen und pruefen YAML-Profile ohne Discovery- oder
Inform-Traffic.

```yaml
uplink_neighbor:
  mac: 02:aa:bb:cc:dd:01
  vlan: 1
  type: usw
  remote_port: 50

port_neighbors:
  - port: 2
    mac: 02:00:5e:00:53:03
    hostname: lab-host-2
    ip: 192.0.2.52
    vlan: 1
    type: usw
    remote_port: 7

port_overrides:
  - port: 2
    name: lab_lan
    role: lan
    network_group: LAN
    interface: eth1
    ip: 192.0.2.51
    netmask: 255.255.255.0
    ipv6:
      - 2001:db8:51::1/64
    speed: 1000
  - port: 3
    name: backup_wan
    role: wan2
    network_group: WAN2
    wan_uptime_percent: 100
    wan_latency_ms: 7
    wan_connected: true
    speed: 2500
  - port: 4
    speed: 100
  - port: 5
    up: false
```

`port_neighbors` mit dem Standardtyp `client` fuellt
`port_table[].mac_table`. Explizite Nicht-Client-Typen beschreiben direkte
Infrastruktur-Nachbarn. Am aktiven `uplink_port` ist ein solcher Nachbar
upstream, an allen anderen Ports downstream. `remote_port` bezeichnet optional
den Port auf dem Nachbargeraet. `hostname` und `ip` bleiben optionale Metadaten;
`name` ist weiterhin ein YAML-Alias fuer `hostname`. `uplink_neighbor` ist
immer Infrastruktur und bleibt standardmaessig `usw`.

Im Linux-`bridge-observe`-Modus werden gelernte Bridge-FDB-MACs zusaetzlich mit
dem lokalen `/proc/net/arp`-Cache abgeglichen, wenn dieser lesbar ist. So koennen
IPv4-Adressen in `port_table[].mac_table` landen, ohne Host-Netzwerk zu
veraendern. Hostnamen werden nicht per DNS geraten; fuer deterministische Labels
`hostname` oder `name` explizit setzen.

Gateway-Modelle melden beobachtete WAN-/LAN-Link-Fakten ueber `if_table`,
`network_table`, eine read-only physische `port_table`, `config_port_table`,
`ethernet_overrides`, `reported_networks`, `uplink`, `uplink_table` und `wans`.
Firmware-WAN-Health-Wrapper wie `wan1` und `uptime_stats` bleiben ohne ihre
exakten device-originated Provider weg. Explizite Assignment-IDs,
Netzwerknamen und VLAN-Metadaten aus
`port_overrides` werden bei Konfiguration in den Gateway-Porttabellen
gespiegelt. Client-Nachbarn werden ueber `network_table[].host_table` mit
`hostname` und `ip` gemeldet, wenn diese konfiguriert sind; Upstream-Switch-
Nachbarn werden nicht als Gateway-Hosts gerendert.

Gateway-Interface-Namen sind Profildaten, keine Host-Interface-Namen. Das
gewaehlte Profil erzeugt aus `gateway_interface_prefix` und physischem
Profil-Portindex die controllerseitigen Namen (`eth0`, `eth1`, ...).
`port_overrides[].interface` ist nur die lokale Quelle fuer MAC, IPv4/IPv6,
Link, Speed und Counter; dieser Hostname erscheint als `source_interface`. Ein
UXG-Pro-Lab, in dem OPNsense `ixl0` an physischem Port 3 haengt, sollte zum
Beispiel so beschrieben werden:

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
```

Die Gateway-Zeilen melden daraus `ifname: eth2`, weil Profil-Port 3 physisch
`eth2` ist; `source_interface: ixl0` dokumentiert nur, woher die lokalen Fakten
kommen. Rolle und Network Group beschreiben die Funktion des Ports. Sie
benennen das physische Profil-Interface nicht um.

Gateway-Management und Gateway-Datenebene bleiben absichtlich getrennt. Der
Stub kann ein Management- oder Transportnetz nutzen, um den Controller ueber
Top-Level-Felder wie `ip`, `controller_url`, `discovery_interface` und
`discovery_targets` zu erreichen. Daraus wird aber keine LAN- oder WAN-Adresse
eines Gateway-Ports. Gateway-WAN/LAN-Zuordnung kommt aus Profilports plus
`port_overrides`; eine Controller-Management-Adresse soll nicht in
`config_network_wan`, `config_network_lan`, `network_table` oder `port_table`
auftauchen, ausser es ist explizit als Adresse dieses Gateway-Ports
konfiguriert.

Ungenutzte physische Profilports bleiben ebenfalls von routed LAN-State
getrennt. Ports mit `role: unassigned`, deaktivierte Ports oder disconnected
Ports ohne explizite Gateway-Rolle sind nur physische Inventur und erben keine
LAN-IP. So entstehen keine zusaetzlichen LAN-/Gateway-Hinweise in Controller-Web
oder Mobile-App.

`port_overrides[].wan_uptime_percent`, `wan_latency_ms`,
`wan_downtime_seconds` und `wan_connected` sind nur lokale Diagnose-Hinweise.
Sie bleiben im lokalen Status sichtbar, werden dem Controller aber nicht als
Firmware-WAN-Verfuegbarkeit, -Latenz, -Downtime oder -Konnektivitaet gemeldet.

`wan_health` kann diese statischen WAN-Hinweise optional durch aktive,
read-only Ping-Samples ersetzen, nachdem Profil-Ports, Beobachtungsdaten und
`port_overrides` zusammengefuehrt wurden. Das gilt fuer Gateway-Profile in
`stub`, `bridge-observe` und `port-map`. Die Samples sind lokale Diagnosen und
werden nicht auf Firmware-WAN-Health-Felder im Inform abgebildet. Standard ist
`off`, es laeuft also keine Netzwerkmessung ohne explizites Opt-in:

```yaml
wan_health:
  source: ping
  interval_seconds: 10
  timeout_ms: 1000
  targets:
    - port: 3
      host: 192.0.2.1
```

Ping-Ergebnisse aktualisieren nur den lokalen WAN-Diagnosestatus. Der Dienst
leitet aus einem einzelnen Ping weder Firmware-Verfuegbarkeit und -Downtime
noch Speedtest-Ergebnisse ab und aendert weiterhin keine Host-Interfaces,
Routen, VLANs, Firewall-Regeln oder Controller-Provisioning-Daten. Wenn ICMP
blockiert ist oder das lokale `ping`-Binary fehlt, zeigen `-status` und
`-status-json` den letzten Probe-Fehler statt automatisch etwas zu reparieren.
Der Probe nutzt die normale Routing-Tabelle des Hosts. `targets[].port` waehlt
nur den Gateway-Port, dessen Telemetrie aktualisiert wird; das ICMP-Paket wird
nicht an `port_overrides[].interface` gebunden.

`wan_health.source` kennt diese Werte:

- `off`: WAN-Diagnose bleibt inaktiv.
- `static`: dokumentiert, dass nur statische `port_overrides[].wan_*`-Werte
  fuer den lokalen Status genutzt werden sollen. Es wird kein Kommando
  ausgefuehrt.
- `ping`: fuehrt lokale read-only Pings fuer `targets[]` aus und aktualisiert
  den lokalen Status des aufgeloesten Ports. Ziel-Ports muessen nach
  `port_overrides` effektiv `wan` oder `wan2` sein.

Provider- und ISP-Namen werden nicht geraten, und WAN-Diagnosen werden nie zu
einem erfolgreichen Speedtest umgedeutet. Nur das exakt gebundene eingebaute
Profil `uxgpro` / `UXGPRO` / `5.0.16.30689` behaelt die in Device-Captures
beobachteten leeren/idle
`speedtest-status`- und `speedtest-status-udapi`-Bloecke. Andere Gateway-Profile
lassen diese Bloecke weg, bis exakte Device-Belege und ein typisierter lokaler
Provider vorhanden sind.

Bei `UXGPRO` ist die Gateway-`port_table` physische Inventur plus optional
explizit konfigurierte Zuweisungsmetadaten. Der Dienst legt weiterhin keine
VLANs an und wendet keine UniFi-Network-Gateway-Settings auf den Host an.
Felder wie `network_group`, `networkconf_id`, `native_networkconf_id`,
`network_name` und `vlan` sind nur Controller-/Payload-Metadaten.

`port_overrides[].interface` ist read-only. Der Dienst kopiert damit MAC,
IPv4-Adresse, nicht-link-lokale IPv6-CIDR-Adressen, Link-State und verfuegbare
Counter-/Speed-Daten eines bestehenden Host-Interfaces in den Inform-Payload
dieses Ports. Das ist fuer FreeBSD-/OPNsense-Stub-only-Gateway-Tests nuetzlich,
wenn WAN/LAN aus echten Interfaces visualisiert werden sollen, ohne
Host-Netzwerk zu veraendern.

Der Dienst baut Interface-basierte Overrides bei jedem Heartbeat neu aus der
unveraenderten Betreiberkonfiguration auf. MAC, IPv4, IPv6, Link-State, Speed,
Medium und Counter folgen damit auch bei deaktivierten Traffic-Rates dem
aktuellen Interface-Snapshot. Explizit konfigurierte Werte behalten Vorrang;
ein vollstaendiger beobachteter Counter-Snapshot behaelt auch einen gueltigen
Reset auf null bei.

`uplink_neighbor` ist fuer reine Stubs und virtuelle Lab-Ports gedacht, bei
denen es keinen physischen Linkpartner gibt. Der Eintrag fuegt eine konfigurierte
MAC-Table-Zeile auf dem aktuellen Uplink-Port ein.

Wenn eine Quelle fehlt oder nicht lesbar ist, loggt der Dienst eine Warnung und
faellt auf Profilwerte zurueck. Dieser Modus darf keine Interfaces anlegen,
keine Adressen setzen und keine Routen aendern.

### `port-map`

Read-only expliziter Mapping-Modus. Jeder UniFi-Port bekommt genau eine Quelle:

- `interface`: ein physisches OS-Interface in den Payload kopieren.
- `disabled`: den Port administrativ deaktiviert, down, mit Speed `0` und ohne
  gelernte MAC-Eintraege melden.
- `unmapped`: den Profil-Port ohne Sensorquelle belassen.

Der Validate-Pfad prueft, dass explizit gemappte Interfaces lokal existieren.
`disabled` und `unmapped` sind ohne OS-Interface gueltig.
In diesem Modus braucht jeder Profil-Port genau einen `port_mappings[]`-Eintrag.
Interface-Daten kommen ueber die Plattform-Fassade aus `net.InterfaceByName`,
Interface-Adressen, Linux-sysfs-Zaehlern/-Speed/-State, optionalen
`/proc/net/dev`-Zaehlern und Best-Effort-Ausgaben von `ifconfig` und `netstat`.

```yaml
profile: us8
operation_mode: port-map
port_mappings:
  - port: 1
    interface: eno1
  - port: 2
    disabled: true
  - port: 3
    unmapped: true
  # Weiterfuehren, bis jeder Profil-Port einen expliziten Eintrag hat.
```

### `host-direct`

Direkter Host-Identity-Modus. Es wird keine separate MAC oder IP angelegt. Der
Spezialwert `mac: host` ist nur in diesem Modus erlaubt und braucht
`observe_interface`, damit der Dienst die Host-Interface-MAC explizit lesen
kann.

### `macvlan`

In diesem Release nur Planungsmodus. Der Modus ist Linux-only und muss mit
`-dry-run-plan` kombiniert werden. Der Dienst druckt die geplanten
Macvlan-Kommandos, fuehrt sie aber nicht aus.

## Passive Quellen

Passive Quellen sind read-only und haengen hinter `internal/platform`. Sie
reichern Payloads oder Status an, veraendern aber kein Host-Netzwerk.

```yaml
lldp_source: lldpd
traffic_rates_enabled: false
log_source: journalctl
proc_source: procfs
dbus_enabled: false
dbus_bus: system
```

`lldp_source: lldpd` ruft `lldpcli -f json show neighbors` mit Timeout auf und
mappt bekannte lokale Interfaces auf Uplink-, Bridge-Member-, `port-map`- oder
`port_overrides[].interface`-Ports. Fehlendes `lldpcli` wird als Warnung
gemeldet und stoppt den Daemon nicht.

LLDP ist fuer Adoption oder fuer einen manuell konfigurierten Topologie-Hinweis
nicht erforderlich. Wenn LLDP verfuegbar ist, reduziert es manuelle Fehler bei
`uplink_neighbor`, weil Upstream-Chassis und Port vom Host-Interface gelernt
werden koennen. Wenn LLDP fehlt, entscheidet die explizite Konfiguration ueber
den direkten Nachbarn. FDB-Sichtbarkeit wird niemals als Richtungsbeweis
verwendet.

`log_source: journalctl` liest aktuelle Linux-Unit-Logs ueber
`journalctl --output=json`. `log_source: syslog` liest eine konfigurierte
Syslog-Datei, fuer FreeBSD-artige Systeme standardmaessig `/var/log/messages`.
Diese Quellen sind ueber Status/Capabilities sichtbar und bleiben read-only.

`proc_source: procfs` ist Linux-only und ergaenzt Interface-Zaehler aus
`/proc/net/dev`; Link-Speed oder Medium kommen weiterhin aus `/sys/class/net`
oder Plattformtools.

`traffic_rates_enabled: true` meldet read-only RX/TX-Byte-Raten sowie bekannte
Byte-, Paket-, Error-, Link-State-, Speed-, Media- und Source-Interface-Metadaten
fuer gemappte oder beobachtete Interfaces in UniFi-Inform-Payloads. Der Schalter
ist per Default aus, damit bestehende Labs ihre bisherige Controller-Darstellung
behalten. Quelle ist derselbe Interface-Counter-Pfad wie bei
`port_overrides[].interface`, `port-map` und Bridge-Observation; Packet Capture,
NetFlow/IPFIX, DPI oder Paket-/Error-Rate-Felder werden dadurch nicht aktiviert.
Gateway- und gemessene ausgewaehlte Switch-Uplink-Felder `rx_rate` und
`tx_rate` werden als Bit/s-Durchsatz gerendert; `rx_bytes-r`, `tx_bytes-r` und
`bytes-r` bleiben Byte/s-Raten.
Die Runtime-Statusdatei schreibt die letzten bereinigten Inform-Traffic-Felder
mit expliziten Byte/s- und Bit/s-Namen, damit Controller-Graphen gegen die
Host-Interface-Zaehler verglichen werden koennen.

Bei einem Switch beschreiben diese Controller-Spalten unterschiedliche Daten
und duerfen nicht wie dieselbe Messung verglichen werden:

- **Experience** ist der normalisierte Link-Speed des ausgewaehlten aktuellen
  Uplinks (offline des letzten Uplinks), zum Beispiel `10 GbE`. Der Wert ist
  weder momentaner Durchsatz noch ein Satisfaction-Score.
- **Download** und **Upload** sind momentane Geraeteraten, die nur aus dem
  ausgewaehlten Uplink kommen; dadurch werden Downstream-Ports nicht doppelt
  gezaehlt. Der aktuelle Uplink behaelt die physische Richtung: RX ist Download
  und TX ist Upload. Die Inform-Root-Summary nutzt dieselbe Quelle in
  Controller-Summary-Richtung: Physisches RX ist Root-TX, physisches TX ist
  Root-RX.
- **24h Usage** ist das kumulative Abrechnungsfenster des Controllers, nicht
  die aktuelle Transferrate.

Die `*-r`-Felder verwenden Byte/s (`B/s`), `rx_rate` und `tx_rate` verwenden
Bit/s. Die erste Counter-Probe, ein ruhender oder getrennter Link und ein
Counter-Reset melden null, statt eine fruehere Rate stehen zu lassen. Traffic,
der innerhalb
einer Software-Bridge bleibt und das ausgewaehlte Uplink-Interface nicht
durchlaeuft, kann nicht als Uplink-Traffic gemessen werden.

Bei einem gemessenen Switch-Uplink werden kumulative Counter in das vom
Top-Level-`uplink` ausgewaehlte Management-Interface gespiegelt. Gemessene
Byte/s- und Bit/s-Raten stehen sowohl dort als auch am ausgewaehlten physischen
Port. Dieser Pfad ist fuer die offizielle Latest-Statistics-Ausgabe von Network
10.5 erforderlich; bei deaktivierter Messung werden keine Ratenfelder ergaenzt.

Die controller-seitige Releaseabnahme loest das neueste veroeffentlichte Stable
und den neuesten veroeffentlichten Release Candidate dynamisch auf und prueft
dasselbe Produktimage in beiden wegwerfbaren Docker-Controller-Spuren. Der
feste `10.4.57`-Pfad ist eine historische fokussierte Regression, kein
Produkt-Pin. Keiner der Pfade veraendert die Main CI.

`dbus_enabled: true` prueft nur optionale System- oder Session-D-Bus-
Konnektivitaet. D-Bus ist fuer den normalen Stub-Betrieb nicht erforderlich.

Traffic-Metadaten sind aktuell nur `traffic_source: off`; Packet Capture und
DPI sind fuer die erste Observation-Wave absichtlich nicht Teil des Scopes.

`management_lan` ist die Konfiguration fuer Switch-Management-VLANs. Werte
`1..4094` werden im Inform-Payload und im Status ausgegeben, `0` laesst das
Feld ungesetzt:

```yaml
management_lan:
  enabled: true
  vlan: 20
  network_name: Management
  mode: preexisting-interface
  interface: vmbr0.20
  ip: 192.0.2.50
  controller_reachable: off
  adoption_strategy: untagged-first
```

`mode: metadata-only` meldet das VLAN nur an den Controller. `mode:
preexisting-interface` ist der empfohlene erste echte Modus: Das VLAN-Interface
muss bereits existieren, und der Dienst nutzt dessen IPv4-Adresse fuer die
gemeldete Management-IP, als Discovery-Quelle und als gebundene lokale
Inform-Source. `mode: planned-host-vlan` ist nur fuer `-dry-run-plan`. Der
Dienst legt weiterhin keine VLAN-Interfaces an und wendet keine Controller-
Provisionierung auf dem Host an.

Gateway-Profile nutzen `management_lan` nicht fuer WAN-/LAN-Modellierung.
Gateway-Datenebene gehoert in `port_overrides`; Controller-Transport und
Management-Erreichbarkeit bleiben in den Top-Level-Runtime-Feldern.
