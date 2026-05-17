# Research

Stand: 2026-05-16

## Kernthese

UniFi zeigt echte UniFi-Geraete nicht wegen LLDP im `Devices`-Tab an, sondern weil sie das Ubiquiti-Discovery- und Inform-Protokoll sprechen. Fuer eine reine Anzeige einer Proxmox-Bridge oder OPNsense-VM braucht `unifi-stubd` deshalb einen minimalen Fake-UniFi-Device-Lifecycle:

1. UDP Discovery senden.
2. `/inform` gegen den Controller sprechen.
3. Adoption/Authkey akzeptieren.
4. Regelmaessig Status-Payloads liefern.
5. Provisioning-Kommandos ignorieren oder als erfolgreich quittieren, ohne Host-Konfiguration zu aendern.

## Wichtigste Fundstellen

Die genaue Herkunftsmatrix steht in [CREDITS.md](../../CREDITS.md). Die
Eintraege unten sind die Quellen, die Protokoll- und Produktrichtung gepraegt
haben.

- [Unofficial UniFi Guide: Discovery](https://jrjparks.github.io/unofficial-unifi-guide/protocols/discovery.html)
- [Unofficial UniFi Guide: Inform](https://jrjparks.github.io/unofficial-unifi-guide/protocols/inform.html)
- [Unofficial UniFi Guide: Adoption](https://jrjparks.github.io/unofficial-unifi-guide/adoption.html)
- [jeffreykog/unifi-inform-protocol](https://github.com/jeffreykog/unifi-inform-protocol)
- [fxkr/unifi-protocol-reverse-engineering](https://github.com/fxkr/unifi-protocol-reverse-engineering)
- [Tamarack: Reverse Engineering the UniFi Inform Protocol](https://tamarack.cloud/blog/reverse-engineering-unifi-inform-protocol)
- [Ubiquiti: UniFi Required Ports Reference](https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference)
- [Ubiquiti: Remote Adoption / Layer 3](https://help.ui.com/hc/en-us/articles/204909754-Remote-Adoption-Layer-3)
- [Ubiquiti: UniFi Security Gateway Datasheet](https://dl.ui.com/datasheets/unifi/UniFi_Security_Gateway_DS.pdf)
- [Ubiquiti: UniFi Security Gateway Quick Start Guide](https://dl.ui.com/qsg/USG/USG_EN.html)
- [Ubiquiti: UXG-Pro Tech Specs](https://techspecs.ui.com/unifi/advanced-hosting/uxg-pro?s=me)

## Attribution-Grenze

Research-Repositories werden als Quellen fuer Protokollfakten, historischen
Kontext und Designideen genannt. Die Implementierung in `unifi-stubd` ist
eigenstaendiger Go-Code. Code aus Research-Repositories darf nur uebernommen
werden, wenn die Lizenz vorher geprueft und die Attribution aktualisiert wurde.

## Alte Projekte

### wvengen/unifi-controllable-switch

[wvengen/unifi-controllable-switch](https://github.com/wvengen/unifi-controllable-switch) ist der naechste Vorfahr fuer dieses Projekt. Es patchte einen TOUGHswitch so, dass er im UniFi Controller als Switch sichtbar und adoptierbar wurde. Besonders relevant:

- `devel/unifi_announce.py`: Discovery-TLVs.
- `devel/unifi_inform.py`: altes Inform-Paket mit `TNBU`, AES-CBC und Payload-Feldern.
- `src/syswrapper.sh`: Adoption ueber `set-adopt <inform_url> <authkey>`.
- `src/unifi-inform-status`: Beispiel fuer `if_table`, `port_table`, `sys_stats`.

Das Projekt scheiterte langfristig an Controller-Version-Drift und Firmware-Hack-Komplexitaet, nicht an der Grundidee.

### stephanlascar/unifi-gateway

[stephanlascar/unifi-gateway](https://github.com/stephanlascar/unifi-gateway) war ein pfSense/USG-Emulator-PoC. Es ist nicht produktionsreif, enthaelt aber nuetzliche Gateway-Payload-Ideen wie:

- `dpi-clients`
- `dpi-stats`
- `dpi-stats-table`
- `config_port_table`
- WAN/LAN-Konfiguration

Gateway-Emulation ist fuer dieses Projekt ein spaeteres Forschungsziel, nicht der MVP.

### jda/pixiedust

[jda/pixiedust](https://github.com/jda/pixiedust) analysiert Inform-Traffic und nutzt PCAPs. Nuetzlich fuer:

- Authkey-Extraktion.
- Beobachtung von `setparam`.
- Vergleich frischer UniFi-Switch-Payloads.

### ZAP-Quebec/unifi-inform

[ZAP-Quebec/unifi-inform](https://github.com/ZAP-Quebec/unifi-inform) ist eine alte Go-Implementierung des Inform-Protokolls. Nuetzlich fuer Header-Struktur, Flags und Nachrichtenmodell.

## Offizielle Grenzen

UniFi dokumentiert Drittanbieter-Geraete nur eingeschraenkt. Third-Party-Gateways koennen in VLAN-/Routing-Szenarien vorkommen, aber Traffic Identification/DPI ist stark an UniFi Gateways gebunden. Fuer dieses Projekt bedeutet das:

- Switch-artige Sichtbarkeit ist realistisch.
- Port-Traffic und MAC-Tabellen sind realistisch.
- Vollstaendige DPI ohne UniFi Gateway ist wahrscheinlich nicht realistisch.
- Simulierte DPI-Werte waeren ein eigenes Reverse-Engineering-Projekt.
