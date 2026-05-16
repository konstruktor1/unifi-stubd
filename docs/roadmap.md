# Roadmap

## Phase 0: Labor festnageln

- [ ] UniFi Network Controller Version auswaehlen und dokumentieren.
- [ ] Controller in isoliertem Lab betreiben.
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
- [ ] Default-Key gegen Controller-Lab testen.
- [x] Controller-Antworten dekodieren und loggen.

## Phase 3: Adoption

- [ ] `_type: setparam` parsen.
- [ ] `mgmt_cfg` in Key/Value zerlegen.
- [ ] `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url` persistieren.
- [ ] Nach Adoption zwei schnelle Inform-Requests senden.
- [ ] Connected-State erreichen.

## Phase 4: Fake-Switch-Payload

- [x] Minimaler Switch-Payload.
- [ ] `port_table` mit stabilen Default-Werten erweitern.
- [ ] `mac_table` aus Linux Bridge FDB fuellen.
- [ ] Port-Zaehler aus sysfs lesen.
- [ ] Mehrere virtuelle Ports fuer `vmbr0`, `tap*`, `veth*` modellieren.

## Phase 5: Betrieb

- [ ] systemd Unit.
- [ ] JSON-Konfiguration voll verdrahten.
- [ ] Rotierendes Debug-Log.
- [ ] Healthcheck/Status-Command.
- [ ] README mit Lab-Beispiel fuer Proxmox.

## Phase 6: Spaetere Forschung

- [x] Built-in SSH-Adoption fuer `syswrapper.sh set-adopt` und `mca-cli-op set-inform`.
- [ ] Gateway-Profil `UGW3`/`UXG`.
- [ ] DPI-Felder aus NetFlow/OPNsense/ntopng synthetisieren.
- [ ] Kompatibilitaetsmatrix pro UniFi Network Version.
