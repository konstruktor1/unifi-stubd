# Roadmap

## Phase 0: Labor festnageln

- [x] Controller-bekanntes 10G-Profil validieren: `USAGGPRO` online/adopted.
- [x] Betriebsmodi und aktuellen Live-Lab-Stand dokumentieren.
- [ ] UniFi Network Controller Version auswaehlen und dokumentieren.
- [x] Controller in isoliertem Lab betreiben.
- [ ] Eine echte UniFi-Switch-Inform-Sequenz mitschneiden, falls Hardware vorhanden ist.
- [ ] Controller-Logs fuer `inform`, `discover`, `devmgr` auf Debug stellen.

## Phase 1: Discovery

- [x] Discovery-TLV-Builder.
- [x] Broadcast/Multicast-Sender.
- [ ] Controller validieren: erscheint das Fake-Geraet als Pending Adoption?
- [ ] TLV-Diff gegen echten Switch dokumentieren.

## Phase 2: Inform ohne Adoption

- [x] `TNBU` Header-Encoder/Decoder.
- [x] AES-CBC + zlib Grundlage.
- [x] AES-GCM Grundlage.
- [x] Minimaler Inform-Client mit Default-Key.
- [ ] Default-Key gegen Controller-Lab validieren.
- [x] Controller-Antworten dekodieren und loggen.

## Phase 3: Adoption

- [x] `_type: setparam` parsen.
- [x] `mgmt_cfg` in Key/Value zerlegen.
- [x] `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url` persistieren.
- [ ] Nach Adoption zwei schnelle Inform-Requests senden.
- [x] Connected-State mit `USAGGPRO` erreichen.

## Phase 4: Fake-Switch-Payload

- [x] Minimaler Switch-Payload.
- [x] `port_table` mit stabilen Default-Werten erweitern.
- [x] Read-only `observe`-Modus fuer Linux-Bridge-/sysfs-Daten ergaenzen.
- [x] `mac_table` aus Linux Bridge FDB fuellen, wenn konfiguriert.
- [x] Port-Zaehler aus sysfs lesen, wenn konfiguriert.
- [x] Mehrere virtuelle Ports fuer `vmbr0`, `tap*`, `veth*` modellieren.

## Phase 5: Betrieb

- [x] OpenRC-Service.
- [x] systemd Unit.
- [x] YAML-Konfiguration voll verdrahten.
- [x] Paket-Builder fuer Debian, RPM, Arch Linux und tgz.
- [ ] Rotierendes Debug-Log.
- [x] Healthcheck/Status-Command.
- [ ] README mit Lab-Beispiel fuer Proxmox.

## Phase 6: Spaetere Forschung

- [x] Built-in SSH-Adoption fuer `syswrapper.sh set-adopt` und `mca-cli-op set-inform`.
- [ ] Aktiven macvlan/ipvlan-Lifecycle nach validiertem Dry-run-Plan bauen.
- [ ] Passiven LLDP-Import aus `lldpd`.
- [ ] Gateway-Profil `UGW3`/`UXG`.
- [ ] DPI-Felder aus NetFlow/OPNsense/ntopng synthetisieren.
- [ ] Kompatibilitaetsmatrix pro UniFi Network Version.
