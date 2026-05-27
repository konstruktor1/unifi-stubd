# Betriebsmodi

`unifi-stubd` zielt zuerst auf Linux-Lab-Hosts, also Proxmox, Alpine und
UTM-Linux-VMs. FreeBSD/OPNsense wird konservativ ueber Stub-Modus,
explizites `port-map`, Bridge-FDB-Parsing und read-only Syslog-Metadaten
unterstuetzt.

## Aktuell validierter Stand

Das validierte Live-Lab-Geraet ist:

- Host: `192.0.2.151`
- Profil: `usaggpro`
- Controller-Modell: `USAGGPRO` / `USW Pro Aggregation`
- MAC: `02:00:5e:00:53:51`
- Controller-State: online und adopted
- Ports: 28 10G-SFP+-Ports und vier 25G-SFP28-Ports
- Uplink: Port 1 durch Live-Override `uplink_port`; Profil-Default ist Port 29

`USAGGPRO` ist aktuell das validierte grosse 10G-Profil. `USWProXG48` bleibt
experimentell, weil der aktuelle Lab-Controller es nicht als bekanntes Pending-
Adoption-Modell angenommen hat.

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

## Modi

### `stub`

Default-Modus. Der Dienst sendet Discovery- und Inform-Payloads nur aus den
Profildaten. Er liest keinen Host-Bridge-State und aendert kein Host-Netzwerk.
Das ist der unterstuetzte FreeBSD-/OPNsense-Modus.

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

Die Topologie-Richtung haengt von der realen Controller-Sicht auf die gemeldete
Device-MAC ab. Wenn der Stub die physische Bridge- oder NIC-MAC nutzt, kann ein
Upstream-UniFi-Switch diese MAC bereits auf einem eigenen Port melden. Der
Controller kann dann diese reale Beobachtung bevorzugen und den Link falsch
herum darstellen, selbst wenn `uplink_neighbor` konfiguriert ist. Eine
synthetische lokal administrierte Stub-MAC vermeidet diese Kollision und ist fuer
reine Darstellungs-Tests meist sauberer. Die physische MAC sollte nur verwendet
werden, wenn genau diese Controller-Heuristik getestet werden soll.

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

port_neighbors:
  - port: 2
    mac: 02:00:5e:00:53:03
    hostname: lab-host-2
    ip: 192.0.2.52
    vlan: 1
    type: usw

port_overrides:
  - port: 2
    name: lab_lan
    role: lan
    network_group: LAN
    interface: eth1
    ip: 192.0.2.51
    netmask: 255.255.255.0
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

`port_neighbors` fuellt `port_table[].mac_table` auf bestimmten Ports. Das ist
nuetzlich, wenn der Controller eine Downstream-Switch- oder Host-MAC auf einem
Nicht-Uplink-Port sehen soll. `hostname` und `ip` sind optionale Client-
Metadaten; `name` wird in YAML als Alias fuer `hostname` akzeptiert. Wenn
`type` fehlt, werden Port-Nachbarn als `client` gemeldet; `uplink_neighbor`
bleibt standardmaessig `usw`.

Im Linux-`bridge-observe`-Modus werden gelernte Bridge-FDB-MACs zusaetzlich mit
dem lokalen `/proc/net/arp`-Cache abgeglichen, wenn dieser lesbar ist. So koennen
IPv4-Adressen in `port_table[].mac_table` landen, ohne Host-Netzwerk zu
veraendern. Hostnamen werden nicht per DNS geraten; fuer deterministische Labels
`hostname` oder `name` explizit setzen.

Gateway-Modelle melden beobachtete WAN-/LAN-Link-Fakten ueber `if_table`,
`network_table`, eine read-only physische `port_table`, `config_port_table`,
`ethernet_overrides`, `reported_networks`, `uplink`, `uplink_table` und `wan1`.
WAN-aehnliche Ports melden zusaetzlich `uptime_stats`-Zeilen. Explizite
Assignment-IDs, Netzwerknamen und VLAN-Metadaten aus
`port_overrides` werden bei Konfiguration in den Gateway-Porttabellen
gespiegelt. Client-Nachbarn werden ueber `network_table[].host_table` mit
`hostname` und `ip` gemeldet, wenn diese konfiguriert sind; Upstream-Switch-
Nachbarn werden nicht als Gateway-Hosts gerendert.

Gateway-Interface-Namen sind Profildaten, keine Host-Interface-Namen. Das
gewaehlte Profil erzeugt aus `gateway_interface_prefix` und physischem
Profil-Portindex die controllerseitigen Namen (`eth0`, `eth1`, ...).
`port_overrides[].interface` ist nur die lokale Quelle fuer MAC, IP, Link,
Speed und Counter; dieser Hostname erscheint als `source_interface`. Ein
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

Fuer Gateway-Lab-Anzeigen sind `port_overrides[].wan_uptime_percent`,
`wan_latency_ms`, `wan_downtime_seconds` und `wan_connected` nur deterministische
Status-Hinweise. Damit kann der Controller ein konfiguriertes WAN/WAN2 als
online, degradiert oder down sehen, ohne Routen anzulegen, Interfaces zu
aendern oder Provisioning vom Controller anzunehmen.

`wan_health` kann diese statischen WAN-Hinweise optional durch aktive,
read-only Ping-Samples ersetzen, nachdem Profil-Ports, Beobachtungsdaten und
`port_overrides` zusammengefuehrt wurden. Das gilt fuer Gateway-Profile in
`stub`, `bridge-observe` und `port-map`; Switch-Profile rendern keine
WAN-Health-Payload-Felder. Standard ist `off`, es laeuft also keine
Netzwerkmessung ohne explizites Opt-in:

```yaml
wan_health:
  source: ping
  interval_seconds: 10
  timeout_ms: 1000
  targets:
    - port: 3
      host: 192.0.2.1
```

Ping-Ergebnisse setzen nur WAN-Telemetrie wie Connected-State, Latenz, Downtime
und Uptime-Prozent. Der Dienst aendert weiterhin keine Host-Interfaces, Routen,
VLANs, Firewall-Regeln oder Controller-Provisioning-Daten. Wenn ICMP blockiert
ist oder das lokale `ping`-Binary fehlt, zeigen `-status` und `-status-json` den
letzten Probe-Fehler statt automatisch etwas zu reparieren.

`wan_health.source` kennt diese Werte:

- `off`: WAN-Health bleibt inaktiv. Der Payload nutzt Link-State und explizite
  `wan_*`-Hints, die bereits am Port konfiguriert sind.
- `static`: dokumentiert, dass nur statische `port_overrides[].wan_*`-Werte
  genutzt werden sollen. Es wird kein Kommando ausgefuehrt.
- `ping`: fuehrt lokale read-only Pings fuer `targets[]` aus und ueberschreibt
  danach nur WAN-Health-Felder am bereits aufgeloesten Port. Ziel-Ports muessen
  nach `port_overrides` effektiv `wan` oder `wan2` sein.

Provider- und ISP-Namen werden nicht geraten. Der Gateway-Payload kann
`speedtest-status.latency` und erfolgsartige Statuswerte aus WAN-Health melden,
fuehrt aber keinen UniFi-Speedtest-Dienst aus und befuellt weder
`speedtest-status.server.provider` noch `isp_name` oder `isp_info`.

Bei `UXGPRO` ist die Gateway-`port_table` physische Inventur plus optional
explizit konfigurierte Zuweisungsmetadaten. Der Dienst legt weiterhin keine
VLANs an und wendet keine UniFi-Network-Gateway-Settings auf den Host an.

`port_overrides[].interface` ist read-only. Der Dienst kopiert damit MAC,
IPv4-Adresse, Link-State und verfuegbare Counter-/Speed-Daten eines bestehenden
Host-Interfaces in den Inform-Payload dieses Ports. Das ist fuer
FreeBSD-/OPNsense-Stub-only-Gateway-Tests nuetzlich, wenn WAN/LAN aus echten
Interfaces visualisiert werden sollen, ohne Host-Netzwerk zu veraendern.

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
werden koennen. Wenn LLDP fehlt, sollte `uplink_neighbor` explizit gesetzt
bleiben. Die Topologie-Richtung bleibt dann aber controller-abhaengig: Der
Controller kann weiterhin bevorzugen, was echte UniFi-Switches ueber dieselbe
MAC melden.

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
