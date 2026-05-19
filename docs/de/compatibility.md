# Kompatibilitaetsmatrix

Diese Matrix verfolgt controller-seitige Lab-Validierung. Sie beschreibt UniFi
Network Verhalten, nicht die Firmware-Source-Inventare.

| UniFi Network | Lab-Ziel | Ergebnis | Notizen |
| --- | --- | --- | --- |
| LinuxServer.io `unifi-network-application:latest`, erfasst am 2026-05-17 | UXG-Pro-Firmware `5.0.16.30689` und host-seitige `unifi-stubd`-Diagnostik | Im Lab adoptiert | Die genaue Application-Version war im Compose-File nicht gepinnt; vor Release-tauglicher Kompatibilitaetsaussage pinnen. |
| Kuenftige gepinnte Version | Switch-Profile `US8`, `US16P150`, `US16XG`, `USAGGPRO`, `USW-Pro-XG-48` | Noch nicht erfasst | Eine Zeile pro getesteter UniFi-Network-Version ergaenzen. |

Kompatibilitaetseintraege muessen Controller-/Application-Version, Profil oder
Firmware-Identitaet, Adoption-Ergebnis, Cipher-Modus und alle durch die
Safety-Policy ignorierten Controller-Response-Typen enthalten.
