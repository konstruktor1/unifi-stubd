# Kompatibilitaetsmatrix

Diese Matrix verfolgt controller-seitige Lab-Validierung. Sie beschreibt UniFi
Network Verhalten, nicht die Firmware-Source-Inventare.

| UniFi Network | Lab-Ziel | Ergebnis | Notizen |
| --- | --- | --- | --- |
| Docker-Lab `lscr.io/linuxserver/unifi-network-application:10.3.58-ls129`, erfasst am 2026-05-20 | `US8` bridge-observe und `UXG`/`uxg-lite` gateway-smoke | Im Lab adoptiert | `make integration-docker` prueft Gateway-Dry-Run-Tabellen, MITM-Inform-Events, Controller-Pending-State, Controller-getriggerte Adoption, lokal persistiertes `STATE=connected` und Controller-`/status` mit `server_version=10.3.58`. |
| Privater echter UniFi-OS-Server-Controller, erfasst am 2026-05-20 | Linux-/Proxmox-`bridge-observe` mit 48-Port-Switch-Profil | Im Lab adoptiert | Validiert wurden echte Controller-Adoption, AES-GCM-Heartbeats nach Adoption, Access-Port-Clients aus Bridge-FDB, getrennte ungenutzte Ports, SFP+-Uplink-Platzierung mit `uplink_port` und `uplink_neighbor`-Topologiemetadaten. Controller-URL, Site-Daten, Device-MACs und Token werden absichtlich nicht in Git erfasst. Tests mit physischer Host-MAC zeigten, dass UniFi Network die Topologie-Richtung umkehren kann, wenn ein echter Upstream-Switch dieselbe MAC bereits meldet. |
| LinuxServer.io `unifi-network-application:latest`, erfasst am 2026-05-17 | UXG-Pro-Firmware `5.0.16.30689` und host-seitige `unifi-stubd`-Diagnostik | Im Lab adoptiert | Die genaue Application-Version war im Compose-File nicht gepinnt; vor Release-tauglicher Kompatibilitaetsaussage pinnen. |
| Kuenftige gepinnte Version | Switch-Profile `US8`, `US16P150`, `US16XG`, `USAGGPRO`, `USW-Pro-XG-48` | Noch nicht erfasst | Eine Zeile pro getesteter UniFi-Network-Version ergaenzen. |

Kompatibilitaetseintraege muessen Controller-/Application-Version, Profil oder
Firmware-Identitaet, Adoption-Ergebnis, Cipher-Modus und alle durch die
Safety-Policy ignorierten Controller-Response-Typen enthalten.
