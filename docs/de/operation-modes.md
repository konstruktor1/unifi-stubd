# Betriebsmodi

`unifi-stubd` zielt zuerst auf Linux-Lab-Hosts, also Proxmox, Alpine und
UTM-Linux-VMs. FreeBSD/OPNsense wird als Stub-only-Ziel unterstuetzt; native
Observation ist dort noch nicht implementiert.

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

### `observe`

Read-only Linux-Observation-Modus. Die Fake-Device-Identitaet bleibt gleich,
aber passive Host-Daten werden genutzt, wenn sie verfuegbar sind:

- `/sys/class/net/<interface>/statistics/*` fuer Port-Zaehler
- `/sys/class/net/<interface>/speed` fuer Uplink-Speed
- `bridge fdb show br <bridge>` fuer gelernte MAC-Table-Eintraege

FDB-Zeilen werden nach Linux-Bridge-Member gruppiert. Das konfigurierte
`observe_interface` wird auf den UniFi-Uplink-Port gelegt, waehrend `tap*`,
`veth*` und andere gelernte Bridge-Member deterministisch auf freie Switch-Ports
mit ihren MAC-Tabellen verteilt werden.

Das Profil waehlt den Uplink-Port per Default. `uplink_port` kann auf eine
positive Portnummer gesetzt werden, um die Uplink-Markierung auf einen
bestimmten physischen Port zu legen, ohne dessen Profil-Speed oder Medium zu
aendern. Beispiel: `uplink_port: 1` setzt beim `usaggpro` den Uplink auf einen
10G-SFP+-Port statt auf die Default-25G-SFP28-Uplinkgruppe.

Profile beschreiben das echte Hardware-Layout: Modell, Portanzahl,
Speed-/Mediengruppen, Default-Portnamen und Default-Gateway-Rollen.
`port_overrides` setzt danach lab-spezifische Zuweisungen und einzelne
Portzustaende:

```yaml
uplink_neighbor:
  mac: 02:aa:bb:cc:dd:01
  vlan: 1
  type: usw

port_neighbors:
  - port: 2
    mac: 02:00:5e:00:53:03
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
    speed: 2500
  - port: 4
    speed: 100
  - port: 5
    up: false
```

`port_neighbors` fuellt `port_table[].mac_table` auf bestimmten Ports. Das ist
nuetzlich, wenn der Controller eine Downstream-Switch- oder Host-MAC auf einem
Nicht-Uplink-Port sehen soll.

Gateway-Modelle melden WAN-/LAN-Zuweisungen ueber `config_port_table`,
`ethernet_overrides`, `network_table` und `reported_networks`.
Switch-artige MAC-Table-Nachbarn koennen vom Controller bei Gateway-
Identitaeten ignoriert werden. Fuer Gateway-Visualisierung daher `role` und
`network_group` nutzen, statt das Hardware-Profil umzubauen.

Bei `UXGPRO` rendert der Controller Gateway-Ports aus seinem Gateway-Modell und
dem gemeldeten WAN-/LAN-State. Er zeigt deshalb nicht dieselbe Switch-
`port_table`-Ansicht wie bei einem UniFi-Switch-Profil.

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

LLDP ist aktuell nur `lldp_source: off`. `lldpd` ist als Quelle geplant, wird
zur Laufzeit aber abgelehnt, bis die Implementierung vorhanden ist.
Traffic-Metadaten sind aktuell nur `traffic_source: off`; Packet Capture und
DPI sind fuer die erste Observation-Wave absichtlich nicht Teil des Scopes.

`management_vlan` ist nur controller-seitige Metadaten. Werte `1..4094` werden
im Inform-Payload und im Status ausgegeben, `0` laesst das Feld ungesetzt. Der
Dienst legt keine VLAN-Interfaces an und wendet keine Controller-Provisionierung
auf dem Host an.
