# Betriebsmodi

`unifi-stubd` zielt zuerst auf Linux-Lab-Hosts, also Proxmox, Alpine und
UTM-Linux-VMs.

## Aktuell validierter Stand

Das validierte Live-Lab-Geraet ist:

- Host: `10.0.0.151`
- Profil: `usaggpro`
- Controller-Modell: `USAGGPRO` / `USW Pro Aggregation`
- MAC: `32:c1:80:4f:7e:bc`
- Controller-State: online und adopted
- Ports: 28 10G-SFP+-Ports und vier 25G-SFP28-Ports
- Uplink: Port 1 durch Live-Override `uplink_port`; Profil-Default ist Port 29

`USAGGPRO` ist aktuell das validierte grosse 10G-Profil. `USWProXG48` bleibt
experimentell, weil der aktuelle Lab-Controller es nicht als bekanntes Pending-
Adoption-Modell angenommen hat.

## Modi

### `stub`

Default-Modus. Der Dienst sendet Discovery- und Inform-Payloads nur aus den
Profildaten. Er liest keinen Host-Bridge-State und aendert kein Host-Netzwerk.

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

LLDP wird aktuell als `lldp_source: off` oder `lldp_source: lldpd` akzeptiert,
aber nur `off` hat heute Runtime-Verhalten. Traffic-Metadaten sind aktuell nur
`traffic_source: off`; Packet Capture und DPI sind fuer die erste
Observation-Wave absichtlich nicht Teil des Scopes.
