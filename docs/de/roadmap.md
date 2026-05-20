# Roadmap

## Phase 0: Labor festnageln

- [x] Controller-bekanntes 10G-Profil validieren: `USAGGPRO` online/adopted.
- [x] Betriebsmodi und aktuellen Live-Lab-Stand dokumentieren.
- [x] UniFi-Network-Kompatibilitaetsmatrix starten.
- [x] Controller in isoliertem Lab betreiben.
- [ ] Eine echte UniFi-Switch-Inform-Sequenz mitschneiden, falls Hardware vorhanden ist.
- [ ] Controller-Logs fuer `inform`, `discover`, `devmgr` auf Debug stellen.

## Phase 1: Discovery

- [x] Discovery-TLV-Builder.
- [x] Broadcast/Multicast-Sender.
- [x] Controller validieren: Docker-Lab bestaetigt Fake-Geraet als Pending Adoption.
- [ ] TLV-Diff gegen echten Switch dokumentieren.

## Phase 2: Inform ohne Adoption

- [x] `TNBU` Header-Encoder/Decoder.
- [x] AES-CBC + zlib Grundlage.
- [x] AES-GCM Grundlage.
- [x] Minimaler Inform-Client mit Default-Key.
- [x] Default-Key-Inform-Pfad gegen Docker-Controller-Lab validieren.
- [x] Controller-Antworten dekodieren und loggen.

## Phase 3: Adoption

- [x] `_type: setparam` parsen.
- [x] `mgmt_cfg` in Key/Value zerlegen.
- [x] `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url` persistieren.
- [x] Post-Adoption-Connected-State im Docker-Controller-Lab verifizieren.
- [x] Connected-State mit `USAGGPRO` erreichen.

## Phase 4: Fake-Switch-Payload

- [x] Minimaler Switch-Payload.
- [x] `port_table` mit stabilen Default-Werten erweitern.
- [x] Read-only `bridge-observe`-Modus fuer Linux-Bridge-/sysfs-Daten ergaenzen.
- [x] `observe` als Migrationsalias fuer `bridge-observe` behalten.
- [x] Read-only `port-map`-Modus fuer explizite Interface-/Disabled-/Unmapped-Portquellen ergaenzen.
- [x] `mac_table` aus Linux Bridge FDB fuellen, wenn konfiguriert.
- [x] Port-Zaehler aus sysfs lesen, wenn konfiguriert.
- [x] Optionale `/proc/net/dev`-Counterquelle ueber `proc_source: procfs`.
- [x] Mehrere virtuelle Ports fuer `vmbr0`, `tap*`, `veth*` modellieren.
- [x] Bridge-Member vor dem Mapping als Bridge/Uplink/Access klassifizieren.
- [x] Remote-MACs, die auf dem physischen Uplink gelernt wurden, aus lokalen
  Access-Port-MAC-Tabellen herausfiltern.
- [x] Ungenutzte Bridge-Observe-Ports als disconnected melden statt
  synthetisch up.
- [x] Explizite SFP-/SFP+-Uplink-Platzierung ueber `uplink_port` fuer
  Bridge-Observe-Profile unterstuetzen, wenn der physische Link nicht der
  GE-Fallback ist.
- [ ] Explizite `uplink_neighbor.remote_port`-Metadaten ergaenzen und bei
  bekanntem Wert in den Switch-Payload uebernehmen.
- [ ] Uplink-Nachbar automatisch aus passivem LLDP ableiten, mit manuellem
  `uplink_neighbor` als deterministischem Override.

## Phase 5: Betrieb

- [x] OpenRC-Service.
- [x] systemd Unit.
- [x] YAML-Konfiguration voll verdrahten.
- [x] Paket-Builder fuer Debian, RPM, Arch Linux und tgz.
- [x] Stub-only FreeBSD-/OPNsense-tgz und rc.d-Artefakt.
- [ ] Rotierendes Debug-Log.
- [x] Healthcheck/Status-Command.
- [x] Plattform-Capability-Status fuer LLDP, Logs, procfs, D-Bus und Traffic-Quellen.
- [x] README und Betriebsmodus-Doku fuer Bridge-Observe-/Port-Map-Lab-Nutzung.
- [x] Docker-Controller-Integration-Smoke-Test.

## Phase 6: Spaetere Forschung

- [x] Built-in SSH-Adoption fuer `syswrapper.sh set-adopt` und `mca-cli-op set-inform`.
- [ ] Aktiven macvlan/ipvlan-Lifecycle nach validiertem Dry-run-Plan bauen.
- [x] Passiven LLDP-Import aus `lldpd`.
- [ ] LLDP-VLAN-/MED-Details, CDP/FDP und Event-Subscriptions.
- [ ] Topologie-Richtung dokumentieren und testen, wenn ein Stub eine physische
  Host-MAC nutzt, die auch fuer einen echten Upstream-UniFi-Switch sichtbar ist.
- [ ] Unterstuetztes Deployment-Muster fuer synthetische Stub-MACs bei
  Proxmox-Bridge-Darstellungen ergaenzen.
- [x] Experimentelle Gateway-Identitaetsprofile `UGW3`, `UXG` und `UXGPRO`.
- [ ] Vollstaendiger Gateway-Statuspayload fuer `UGW3`/`UXG`.
- [ ] DPI-Felder aus NetFlow/OPNsense/ntopng synthetisieren.
- [ ] Grosse Firmware-Research-Labs in Companion-Repo oder klar getrenntes Research-Paket verschieben.
- [x] Kompatibilitaetsmatrix pro UniFi Network Version.
- [x] JSON-Schema fuer YAML-Konfiguration.
