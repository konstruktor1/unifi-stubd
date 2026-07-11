# Kompatibilitaetsmatrix

Diese Matrix verfolgt controller-seitige Lab-Validierung. Sie beschreibt UniFi
Network Verhalten, nicht die Firmware-Source-Inventare.

| UniFi Network | Lab-Ziel | Ergebnis | Notizen |
| --- | --- | --- | --- |
| Docker-Lab `lscr.io/linuxserver/unifi-network-application:10.4.57-ls135`, validiert am 2026-07-11 | `US8` bridge-observe und `UXG`/`uxg-lite` gateway-smoke | Adoptiert und stabil verbunden | `make integration-docker` erzwingt `server_version=10.4.57` und prueft Discovery, Adoption, persistierten Auth-State, AES-GCM sowie stabile Connected-Samples. Anders als 10.3.58 sendete dieser Controller kein Gateway-Kommando; der Gate belegte vor und nach dem 75-Sekunden-Stabilitaetsfenster null verarbeitete, ausgelieferte, duplizierte oder ausstehende Kommandos bei gesundem controller-gesteuertem Inform-Takt. |
| Docker-Lab `lscr.io/linuxserver/unifi-network-application:10.3.58-ls129`, validiert am 2026-07-10 | `US8` bridge-observe und `UXG`/`uxg-lite` gateway-smoke | Adoptiert und stabil verbunden | `make integration-docker` bootstrapt leere Controller-Volumes, erzwingt `server_version=10.3.58`, erfasst weitergeleitetes Discovery, prueft Pending Adoption und Controller-Adoption, dekodiert bereinigte Default-Key-`setparam`-Evidenz, prueft persistierten Connected-/Auth-/cfgversion-/GCM-State, erfasst mindestens zwei AES-GCM-Informs und verlangt drei aufeinanderfolgende Connected-Controller-Samples. Der Gateway-Gate belegt zusaetzlich ein verarbeitetes `wg-genkeypair` mit WireGuard-geklemmtem X25519-Key, ein ausgeliefertes Ergebnis, kein Duplikat oder ausstehendes Ergebnis und nach 75 Sekunden keine wiederkehrende Task. Er prueft, dass der positive Controller-Intervallhinweis zum lokal begrenzten effektiven und naechsten Takt wird, ohne aktiven Backoff. |
| Privater echter UniFi-OS-Server-Controller, erfasst am 2026-05-20 | Linux-/Proxmox-`bridge-observe` mit 48-Port-Switch-Profil | Im Lab adoptiert | Validiert wurden echte Controller-Adoption, AES-GCM-Heartbeats nach Adoption, Access-Port-Clients aus Bridge-FDB, getrennte ungenutzte Ports, SFP+-Uplink-Platzierung mit `uplink_port` und `uplink_neighbor`-Topologiemetadaten. Controller-URL, Site-Daten, Device-MACs und Token werden absichtlich nicht in Git erfasst. Tests mit physischer Host-MAC zeigten, dass UniFi Network die Topologie-Richtung umkehren kann, wenn ein echter Upstream-Switch dieselbe MAC bereits meldet. |
| LinuxServer.io `unifi-network-application:latest`, erfasst am 2026-05-17 | UXG-Pro-Firmware `5.0.16.30689` und host-seitige `unifi-stubd`-Diagnostik | Im Lab adoptiert | Die genaue Application-Version war im Compose-File nicht gepinnt; vor Release-tauglicher Kompatibilitaetsaussage pinnen. |
| Kuenftige gepinnte Version | Switch-Profile `US8`, `US16P150`, `US16XG`, `USAGGPRO`, `USW-Pro-XG-48` | Noch nicht erfasst | Eine Zeile pro getesteter UniFi-Network-Version ergaenzen. |

Kompatibilitaetseintraege muessen Controller-/Application-Version, Profil oder
Firmware-Identitaet, Adoption-Ergebnis, Cipher-Modus und alle durch die
Safety-Policy ignorierten Controller-Response-Typen enthalten.
