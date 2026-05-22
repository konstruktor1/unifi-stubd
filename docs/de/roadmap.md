# Roadmap

## Phase 0: Labor festnageln

- [x] Controller-bekanntes 10G-Profil validieren: `USAGGPRO` online/adopted.
- [x] Betriebsmodi und aktuellen Live-Lab-Stand dokumentieren.
- [x] UniFi-Network-Kompatibilitaetsmatrix starten.
- [x] Controller in isoliertem Lab betreiben.
- [ ] Eine echte UniFi-Switch-Inform-Sequenz mitschneiden, falls Hardware vorhanden ist.
- [ ] Echte UXG-/Cloud-Gateway-Inform-Baselines erfassen, sobald Hardwarezugriff
  vorhanden ist.
- [ ] Controller-Logs fuer `inform`, `discover`, `devmgr` auf Debug stellen.
- [ ] Anonymisierte Capture-Review-Checkliste ergaenzen, bevor Payload-Fixtures
  committed werden.

## Phase 1: Discovery

- [x] Discovery-TLV-Builder.
- [x] Broadcast/Multicast-Sender.
- [x] Controller validieren: Docker-Lab bestaetigt Fake-Geraet als Pending Adoption.
- [x] Explizite `discovery_targets` fuer geroutete oder FreeBSD-/OPNsense-Labs.
- [x] Optionales `discovery_interface`-Source-Binding.
- [ ] TLV-Diff gegen echten Switch dokumentieren.
- [ ] Discovery ueber geroutete Management-Netze ohne All-Ones-Broadcast
  validieren.
- [ ] Entscheiden, ob STUN-Metadaten fuer unterstuetzte Stub-Profile noetig sind.

## Phase 2: Inform ohne Adoption

- [x] `TNBU` Header-Encoder/Decoder.
- [x] AES-CBC + zlib Grundlage.
- [x] AES-GCM Grundlage.
- [x] Minimaler Inform-Client mit Default-Key.
- [x] Default-Key-Inform-Pfad gegen Docker-Controller-Lab validieren.
- [x] Controller-Antworten dekodieren und loggen.
- [x] Sichere Metadaten fuer ignorierte Provisioning-Antworten aufzeichnen.
- [ ] Fixture-Abdeckung fuer `include_blocks`-Antworten wie `gw_caps`,
  `dns_shield_servers`, aktive Leases und WAN-Status-Bloecke ergaenzen.
- [ ] Retry-/Backoff- und Timeout-Policy fuer Controller-Inform-Fehler testen.
- [ ] Begrenzte Status-Historie fuer letzte Inform-Antworttypen ergaenzen, ohne
  Authkeys oder Controller-Secrets zu leaken.

## Phase 3: Adoption

- [x] `_type: setparam` parsen.
- [x] `mgmt_cfg` in Key/Value zerlegen.
- [x] `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url` persistieren.
- [x] Post-Adoption-Connected-State im Docker-Controller-Lab verifizieren.
- [x] Connected-State mit `USAGGPRO` erreichen.
- [x] Forget/Delete/Remove/Restore-Default-Antworten als lokalen Stub-Reset
  behandeln.
- [x] Built-in SSH-Adoption-Kompatibilitaet fuer eingeschraenkte `set-adopt`-
  und `set-inform`-Command-Shapes.
- [ ] Adoption-Failure-Diagnostik ergaenzen: falscher Key, stale Model, stale
  MAC, unerreichbarer Controller und Controller-seitige Ablehnung unterscheiden.
- [ ] Migrationstests fuer Adoption-State beim Wechsel von `/tmp`-Runs zu
  paketverwalteten Services.

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
- [x] Gelernte MAC-Eintraege mit lokalen ARP-IPv4-Metadaten anreichern, wenn
  verfuegbar.
- [x] Konfigurierte `port_neighbors` mit deterministischen Hostname- und
  IP-Metadaten unterstuetzen.
- [x] Ungenutzte Ports disconnected halten, statt controller-sichtbaren
  Fake-Traffic auf leeren Ports zu erzeugen.
- [ ] Explizite `uplink_neighbor.remote_port`-Metadaten ergaenzen und bei
  bekanntem Wert in den Switch-Payload uebernehmen.
- [ ] Uplink-Nachbar automatisch aus passivem LLDP ableiten, mit manuellem
  `uplink_neighbor` als deterministischem Override.
- [ ] LLDP-/CDP-/FDP-Nachbarn fuer lokales Interface, Chassis, Remote-Port,
  VLAN und Systemnamen normalisieren.
- [ ] Topologie-Regressionstests fuer Physical-MAC-Kollision gegen synthetische
  locally administered Stub-MACs.
- [ ] Profilabdeckung fuer weitere typische UniFi-Switch-Modelle aus Homelabs.

## Phase 5: Betrieb

- [x] OpenRC-Service.
- [x] systemd Unit.
- [x] YAML-Konfiguration voll verdrahten.
- [x] Paket-Builder fuer Debian, RPM, Arch Linux und tgz.
- [x] Stub-only FreeBSD-/OPNsense-tgz und rc.d-Artefakt.
- [x] GitHub-Pre-Release mit neutralen Paket-Artefakten und Checksums.
- [x] GitHub-Pages-Alpha-Repositories fuer APT, RPM, Arch Linux und
  FreeBSD-/OPNsense-Tarballs.
- [x] Externe private Host-Konfigurationsstruktur fuer echte Lab-Konfigs und
  Installationsprotokolle.
- [ ] Rotierendes Debug-Log.
- [x] Healthcheck/Status-Command.
- [x] Plattform-Capability-Status fuer LLDP, Logs, procfs, D-Bus und Traffic-Quellen.
- [x] README und Betriebsmodus-Doku fuer Bridge-Observe-/Port-Map-Lab-Nutzung.
- [x] Docker-Controller-Integration-Smoke-Test.
- [ ] Projekt-Release-GPG-Key und signierte APT-/RPM-/Arch-Paketmetadaten.
- [ ] FreeBSD-`pkg`-Paket und Repository, sobald echte FreeBSD-Paketartefakte
  den aktuellen Tarball-Pfad ersetzen.
- [ ] Getrennte Paketkanaele `alpha`, `testing` und `stable`.
- [ ] Package-Manager-Install-Smoke-Tests in CI fuer APT-, RPM- und Pacman-Repos.
- [ ] Pages-Repository-Freshness-Check in CI fuer APT `Packages.gz`, RPM
  `repomd.xml`, Arch `unifi-stubd.db`, FreeBSD-Tarballs und Checksums.
- [ ] Release-SBOM und Provenance-Artefakte neben Checksums veroeffentlichen.
- [ ] Paket-Upgrade-, Downgrade-, Uninstall- und Purge-Tests mit explizitem
  Adoption-State-Retention-Verhalten ergaenzen.
- [ ] Package-Rollback-Runbook ergaenzen, das paketverwaltete Services wieder
  auf einen erfassten temporaeren Startbefehl zurueckfuehrt.
- [ ] GitHub-Actions-Update fuer den Node-24-Runner-Uebergang.
- [ ] Verschluesselte Backup- oder Rotations-Policy fuer die externe private
  Host-Konfigurationsablage.
- [ ] Live-Controller-Post-Install-Gate: keine doppelten Geraete und keine
  `ADOPT_FAILED`-Schleife nach Paket-Rollout.
- [ ] Operator-Runbook fuer sichere Sammlung von `status-json`, Service-Logs
  und Paketversionen ohne private Controller-Daten.

## Phase 6: Spaetere Forschung

- [x] Built-in SSH-Adoption fuer `syswrapper.sh set-adopt` und `mca-cli-op set-inform`.
- [ ] Aktiven macvlan/ipvlan-Lifecycle nach validiertem Dry-run-Plan bauen.
- [ ] Schmalen lokalen Adapter fuer `planned-host-vlan` pruefen; getrennt von
  Controller-Provisioning halten.
- [x] Passiven LLDP-Import aus `lldpd`.
- [ ] LLDP-VLAN-/MED-Details, CDP/FDP und Event-Subscriptions.
- [ ] Topologie-Richtung dokumentieren und testen, wenn ein Stub eine physische
  Host-MAC nutzt, die auch fuer einen echten Upstream-UniFi-Switch sichtbar ist.
- [ ] Unterstuetztes Deployment-Muster fuer synthetische Stub-MACs bei
  Proxmox-Bridge-Darstellungen ergaenzen.
- [x] Experimentelle Gateway-Identitaetsprofile `UGW3`, `UXG-Lite`, `UXGPRO`
  und `UCGF`.
- [x] Gateway-WAN-/LAN-Port-Reporting fuer UXG-artige Lab-Stubs.
- [x] Per Konfiguration schaltbares Traffic-Rate-Reporting aus read-only
  Interface-Countern.
- [ ] Vollstaendiger Gateway-Statuspayload fuer `UGW3`/`UXG`.
- [ ] Gateway-UI-Validierung fuer Settings > Internet, Devices und Ports ueber
  gepinnte UniFi-Network-Versionen.
- [ ] WAN-Health-Felder fuer Uptime, Latency, Peak Utilization und ISP-Label,
  sofern der Controller read-only Werte akzeptiert.
- [ ] Active-Lease- und Host-Table-Shape fuer Gateway-Client-Sichtbarkeit, ohne
  erfundene Per-Client-Traffic-Counter.
- [ ] UniFi-UI-Verifikation fuer Switch- und Gateway-Traffic-Activity nach
  paketbasierter Installation.
- [ ] Breitere WAN-Activity- und Link-State-Kompatibilitaetstests ueber mehrere
  UniFi-Network-Versionen.
- [ ] `traffic_source`-Adapter erst nach stabilem Counter-Pfad implementieren;
  Kandidaten sind NetFlow/IPFIX, OPNsense-Telemetrie und ntopng.
- [ ] DPI-Felder aus NetFlow/OPNsense/ntopng synthetisieren.
- [ ] FreeBSD-`SIOCGIFMEDIA`-Reader fuer Interface-Media und Speed.
- [ ] FreeBSD-Interface-Counter-Paritaet fuer `port-map` und spaeteres
  `bridge-observe`.
- [ ] Volle FreeBSD-`bridge-observe`-Paritaet, sobald native Counter-/Media-
  Reader vorhanden sind.
- [ ] Grosse Firmware-Research-Labs in Companion-Repo oder klar getrenntes Research-Paket verschieben.
- [ ] UXG-Lite- und UCG-Fiber-ARM64-Firmware-Wrapper mit `strace` oder breiterem
  LD_PRELOAD-Shim debuggen.
- [ ] Deterministische UDM-Pro-SE-Netdev- und Switch-Driver-Mock-Oberflaeche
  fuer das Firmware-Referenzlab.
- [ ] UGW3-Legacy-Board-Identity- und EEPROM-Mock-Layer fuer den Firmware-Runner.
- [ ] GPL-/Source-Bundle-Prozess klaeren, bevor externe firmware-abgeleitete
  Quellen oder strukturierte Daten in Projektcode uebernommen werden.
- [x] Kompatibilitaetsmatrix pro UniFi Network Version.
- [x] JSON-Schema fuer YAML-Konfiguration.
- [ ] Config-Referenzdoku aus JSON-Schema generieren, damit README, paketierte
  Beispiele und Schema synchron bleiben.
