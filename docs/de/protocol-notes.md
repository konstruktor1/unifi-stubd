# Protocol Notes

Diese Notizen sind Arbeitsmaterial. Sie sind absichtlich pragmatisch und muessen im Lab gegen konkrete UniFi Network Versionen validiert werden.

## Discovery

Discovery laeuft ueber UDP `10001`. Historische Implementierungen senden an:

- `255.255.255.255:10001`
- `233.89.188.1:10001`

Manche FreeBSD-/OPNsense-Lab-Routen koennen nicht an die All-Ones-Broadcast-
Adresse senden. Fuer diese Faelle kann `discovery_targets` explizite UDP-Ziele
setzen, zum Beispiel die LAN-Broadcast-Adresse `192.0.2.255:10001`. Leere
`discovery_targets` behalten die Defaults oben.

Paketform:

```text
u8  version
u8  packet_type
u16 payload_length_be
TLV payload
```

TLV-Form:

```text
u8  type
u16 length_be
[]  value
```

Interessante TLVs aus alten Implementierungen:

| Type | Bedeutung |
| --- | --- |
| `0x01` | MAC-Adresse |
| `0x02` | MAC + IPv4 |
| `0x0a` | Uptime |
| `0x0b` | Hostname / Name |
| `0x12` | Announcement sequence |
| `0x13` | Serial / MAC |
| `0x15` | Model identifier |
| `0x16` | Firmware version |
| `0x17` | Default/factory flag |

## Inform

Inform laeuft ueber HTTP POST:

```text
POST http://<controller>:8080/inform
Content-Type: application/x-binary
User-Agent: AirControl Agent v1.0
```

Binärheader:

```text
0x00  4 bytes  magic "TNBU"
0x04  4 bytes  packet version, meist 0 oder 1
0x08  6 bytes  device MAC
0x0e  2 bytes  flags
0x10 16 bytes  IV / nonce
0x20  4 bytes  payload version, meist 1
0x24  4 bytes  payload length
0x28  n bytes  payload
```

Bekannte Flags:

| Flag | Bedeutung |
| --- | --- |
| `0x01` | encrypted |
| `0x02` | zlib |
| `0x04` | snappy |
| `0x08` | AES-GCM in neueren UniFi-Versionen |

Historisch war AES-CBC + PKCS#7 + zlib verbreitet. Neuere Controller/Geraete koennen `use_aes_gcm=true` im `mgmt_cfg` setzen.

## Adoption

Minimaler Ablauf:

1. Stub startet im Factory-State.
2. Stub sendet Discovery und/oder Inform mit Default-Key.
3. Controller zeigt `Pending Adoption`.
4. Nach Klick auf Adopt antwortet der Controller mit `_type: "setparam"`.
5. `mgmt_cfg` enthaelt Werte wie `authkey`, `cfgversion`, `stun_url`, `mgmt_url`, `use_aes_gcm`.
6. Stub speichert `authkey` und sendet danach mit diesem Key weiter.
7. Controller schickt spaeter `noop`, `setparam`, Provisioning- oder Restart-Kommandos.

Beobachteter UXG-Pro-5.0.16-Ablauf im Controller-Lab:

| Phase | Request/Response-Form | Stub-Verhalten |
| --- | --- | --- |
| Vor Adoption | Firmware-Informs nutzen den Default-Key und bekommen HTTP `404`, bis ein Admin `Adopt` klickt. | `default=true`, `state=1` behalten. |
| Erster akzeptierter Inform | Controller antwortet HTTP `200` mit `_type: "setparam"` und `mgmt_cfg`. | `authkey`, `cfgversion` und `use_aes_gcm=true` parsen und persistieren. |
| Key-Wechsel | Firmware sendet den naechsten Inform sofort mit Adopted-Key und AES-GCM. | Fuer alle weiteren Informs den Adopted-Key verwenden. |
| Adopted Inform | Firmware meldet `default=false`, danach `state=2`. | Geraet als connected behandeln, sobald der Controller `noop` liefert. |
| System-Konfiguration | Controller sendet ein weiteres `_type: "setparam"` mit `system_cfg`. | Nur sichere Metadaten erfassen; keine Host-User, Firewall, Routen, Zertifikate, Tokens oder Secrets anwenden. |
| Steady State | Controller antwortet mit `_type: "noop"`, `interval` und `include_blocks`. | Mit dem geforderten Intervall weiter informieren; beobachtet wurden 10 Sekunden. |

Das erste im Lab beobachtete `mgmt_cfg` enthielt `cfgversion`, `stun_url`,
`mgmt_url`, `authkey`, `use_aes_gcm=true` und `report_crash=true`. Ein neues
`inform_url` war nicht noetig; die Firmware nutzte die bestehende Inform-URL
weiter und wechselte nur Key/Cipher-Kontext.

Das adoptierte `system_cfg` war ein JSON-String mit den Top-Level-Keys
`ubntconf` und `udapi`. Das `udapi`-Objekt enthielt Interfaces, Services,
System-User, Firewall-Sets/-Filterketten/-Settings, statische Routen und
Radius-Profile. Das ist Provisioning-Datenmaterial fuer ein echtes Gateway,
aber kein sicherer Satz Host-Anweisungen fuer `unifi-stubd`.

Alternative Adoption ueber SSH:

```text
/usr/bin/syswrapper.sh set-adopt <inform_url> <authkey>
```

Fuer `unifi-stubd` waere ein kleiner SSH-Shim moeglich, aber L3-Inform-Adoption ist der bessere MVP.

## Minimaler Switch-Payload

Wichtige Felder:

| Feld | Zweck |
| --- | --- |
| `mac` | stabile Fake-MAC |
| `ip` | sichtbare IP |
| `hostname` | Anzeigename |
| `model` | z.B. `US8`, `US8P60`, `US16P150` |
| `model_display` | Anzeige im Controller |
| `version` | Firmware-Version |
| `serial` | meist MAC ohne Doppelpunkte |
| `num_port` | Anzahl Switch-Ports |
| `cfgversion` | Controller-Konfigurationsversion |
| `uptime` | Status/Connected-State |
| `time` | Device-Zeit |
| `if_table` | Management-Interface |
| `ethernet_table` | Controller-seitige Ethernet-/Portzahl-Tabelle |
| `port_table` | Switch-Ports |
| `port_table[].speed` | Port-Speed in Mbps, z.B. `1000` oder `10000` |
| `port_table[].media` | Medienkennung, z.B. `GE` oder `SFP+` |
| `port_table[].mac_table` | gesehene Clients/VMs |
| `sys_stats` | CPU/RAM/Load |

Mixed-Speed-Switch-Profile sollten das komplette physische Portlayout in
`port_table` melden. `USW-Pro-XG-48` wird zum Beispiel mit 16 2.5G-RJ45-Ports,
32 10G-RJ45-Ports und vier 25G-SFP28-Ports modelliert. Die Management-Speed in
`if_table` kommt vom gewaehlten Uplink-Port.

Profile sollen Hardware-Form behalten. Lab-Zuweisungen wie "dieser Port ist
WAN", "dieser Port ist LAN" oder "dieser Port stellt Backup-WAN dar" gehoeren
in `port_overrides[].role` und `port_overrides[].network_group`.

Bei alten Lab-Laeufen fuehrte ein fehlendes oder zu kleines `uptime` in `mac_table` zu Controller-Problemen. Deshalb sollte jeder MAC-Tabelleneintrag ein plausibles `uptime` besitzen.

Ein Profilwechsel sollte als neue Geraeteidentitaet behandelt werden. Praktisch heisst das: neue Fake-MAC oder `-mac auto`, weil UniFi Modellinformationen pro MAC cached und ein spaeterer Modellwechsel am selben MAC haengen bleiben kann.

## Modellwahl

Fuer den MVP:

- `US8`: simpel, keine PoE-Pflicht.
- `US8P60`: ebenfalls klein, aber PoE-Felder koennen erwartet werden.
- `US16P150`: 18-Port-Profil fuer US-16-150W-aehnliches Verhalten, mit 16
  1G-RJ45-Ports und zwei 1G-SFP-Uplinks.
- `US16XG`: 16-Port-10G-Profil fuer Aggregation-/SFP+-Pruefungen, mit zwoelf
  1/10G-SFP+-Ports und vier 1/10G-RJ45-Ports.
- `USAGGPRO`: groesstes controller-bekanntes 10G-Profil, gegen aeltere UniFi-
  Modelldatenbanken validiert, mit 28 10G-SFP+- und vier 25G-SFP28-Ports.
- `USW-Pro-XG-48`: groesstes eingebautes 10G-Access-Switch-Profil mit gemischten
  2.5G-, 10G- und 25G-SFP28-Portgruppen.
- `UGW3`: experimentelles Legacy-UniFi-Security-Gateway-Identitaetsprofil. Es
  meldet Device-Type `ugw`, Modell `UGW3` und drei 1G-Ports mit den Namen
  `WAN 1`, `LAN 1` und `WAN 2 / LAN 2`.
- `UXGPRO`: experimentelles UniFi-Next-Generation-Gateway-Pro-
  Identitaetsprofil. Es meldet Device-Type `uxg`, Modell `UXGPRO`, Firmware
  `5.0.16.30689`, zwei 1G-RJ45-Ports `WAN` und `LAN` sowie zwei 10G-SFP+-
  Ports `WAN2` und `LAN2`. Der Default-Uplink bleibt `WAN`; SFP+-Internet-
  Labs werden ueber `uplink_port` und `port_overrides` remapped.

`UGW3` und `UXGPRO` sind in diesem Release nur Identitaets-/Profil-Stubs. Fuer
einen vollstaendigen Gateway-Payload fehlen weiterhin WAN/LAN-State, Routing,
DHCP, Firewall, DPI und Health-Felder. Der aktuelle UXG-Pro-Payload sendet
Gateway-Tabellen wie `reported_networks` und `uplink`, aber der Controller kann
Gateway-Ports weiterhin aus seinem internen Modell rendern, statt eine
Switch-artige `port_table` zu uebernehmen. Gateway-Modelle wie `UGW4`,
Cloud-Gateway-Geraete und EFG erst spaeter pruefen.
