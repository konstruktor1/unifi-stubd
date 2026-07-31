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

`discovery_interface` kann das lokale Quellinterface fuer Discovery-Sends
festlegen. Das ist absichtlich explizit; der Dienst raet nicht, welches Lab-
oder Management-Netz Discovery-Traffic sehen soll.

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

Die konfigurierte Inform-URL ist eine explizite Vertrauensgrenze. Der Client
folgt HTTP `301`, `302`, `307` oder `308` nicht, behaelt die erste Antwort als
Nicht-200-Lifecycle-Ergebnis, entschluesselt beziehungsweise dispatcht ihren
Body nicht und uebernimmt niemals `Location`. Nur ein validiertes
`mgmt_cfg.inform_url` darf die persistierte Management-URL aendern. Fuer einen
vom Controller gelieferten Wert bedeutet validiert: hoechstens 2.048 Bytes,
eine absolute `http`- oder `https`-URL mit Host und weder Userinfo, Query noch
Fragment. Diese URL-Policy ist eine bewusste Stub-Sicherheitsentscheidung und
kein rekonstruierter Firmwareclaim. Sie gilt fuer gelernte und persistierte
Management-URLs; sie ist weder Host-/CIDR-Allowlist noch Pfadredaktion und
definiert die betreiberkonfigurierte initiale Controller-URL nicht neu. Pfade
duerfen deshalb keine Secrets tragen.

Binärheader:

```text
0x00  4 bytes  magic "TNBU"
0x04  4 bytes  packet version, exakt 0
0x08  6 bytes  device MAC
0x0e  2 bytes  flags
0x10 16 bytes  IV / nonce
0x20  4 bytes  payload version, exakt 1
0x24  4 bytes  payload length
0x28  n bytes  payload
```

Alle Integerfelder sind Big Endian. Der Decoder verlangt exakt
`len(packet) == 40 + payload_length`; kurze Header, nicht unterstuetzte
Versionen, uebergrosse Laengenangaben, abgeschnittene Payloads und angehaengte
Bytes werden vor Entschluesselung oder JSON-Verarbeitung abgelehnt.

Bekannte Flags:

| Flag | Bedeutung |
| --- | --- |
| `0x01` | encrypted |
| `0x02` | zlib |
| `0x04` | snappy |
| `0x08` | AES-GCM in neueren UniFi-Versionen |

Historisch war AES-CBC + PKCS#7 + zlib verbreitet. Die implementierte CBC-Form
verwendet AES-128-CBC mit einem exakten 16-Byte-Key und fuer jedes Paket einem
neuen kryptografisch zufaelligen 16-Byte-IV. Plaintext wird strikt nach PKCS#7
aufgefuellt; bereits ausgerichteter Plaintext erhaelt einen vollstaendigen
zusaetzlichen Padding-Block. CBC-Ciphertext muss deshalb ein nichtleeres
Vielfaches von 16 Bytes sein; fehlerhaftes oder inkonsistentes Padding wird
abgelehnt.

Kompression ist eine explizite, bereits aufgeloeste Wire-Policy-Eingabe. Wenn
sie aktiviert ist, laeuft zlib vor PKCS#7-Padding und Verschluesselung;
andernfalls traegt CBC vor dem Padding rohes JSON. Der Low-Level-Inform-Client
leitet zlib nicht aus Nullwert-Optionen ab. Durch zlib expandierter Plaintext
und Encoder-Input sind auf 16 MiB begrenzt. Die exakte Build-zu-Wire-Auswahl
bleibt ein separater Profilaktivierungsvertrag.

Neuere Controller/Geraete koennen `use_aes_gcm=true` im `mgmt_cfg` setzen. Der
AES-128-GCM-Envelope des Produkts verwendet den protokollspezifischen
16-Byte-Nonce, authentifiziert den vollstaendigen finalen 40-Byte-Header als
Additional Data, haengt einen 16-Byte-Tag an und zaehlt diesen Tag in
`payload_length` mit. Manipulationen an Header, Nonce, Ciphertext oder Tag
werden geschlossen abgelehnt; ein konventioneller 12-Byte-GCM-Nonce ist fuer
dieses Wireformat ungueltig.

Controller-Antworten verwenden eine richtungsspezifische Empfangs-Policy. Der
Client bindet das Response-Flag `0x08` an den bereits fuer den Request
ausgewaehlten Ciphermodus und lehnt beide CBC/GCM-Abweichungen ab, statt aus dem
Response-Header neu zu verhandeln. Nach Authentisierung und Entschluesselung im
passenden Modus geht der Plaintext direkt an die JSON-Verarbeitung:
Response-zlib und Response-Snappy werden nicht akzeptiert. Ein Policyfehler
legt keinen Plaintext offen. Das begrenzte Client-Ergebnis darf HTTP-Status,
rohe Response-Bytes und kopierte TNBU-Paketmetadaten behalten; der persistierte
Runtime-Status behaelt nur Body-Presence, Response-Flags und den bereinigten
Fehler. Der richtungsneutrale Decoder bleibt fuer Device-Requests und
Protokollwerkzeuge erhalten, wo zlib-Request-Bodies gueltig sind.

Der Empfangspfad besitzt zusaetzlich explizite lokale Ressourcengrenzen. Ein
TNBU-Header darf hoechstens 16 MiB plus 64 KiB codierte Payload angeben; dies
deckt begrenztes zlib-Framing und Cipher-Overhead ab. Groessere Angaben werden
vor dem Kopieren des Bodys abgelehnt. Entschluesselung oder Dekompression duerfen
hoechstens 16 MiB Plaintext liefern. Bevor Controller-JSON materialisiert wird,
ist es auf 4 MiB, 64 verschachtelte Objekte/Arrays und 65.536 Tokens (Namen,
Werte und Delimiter) begrenzt. Dieselben JSON-Grenzen gelten erneut fuer ein als
String codiertes JSON-Dokument in `system_cfg`. Diese Werte sind lokale
Stub-Sicherheitsregeln und keine aus Firmware abgeleiteten Schwellen; eine
ueberschrittene Grenze verwirft die gesamte Antwort ohne Adoption-State-Update.

Jeder Daemonlauf schreibt eine begrenzte, bereinigte `inform_session` in den
Runtime-Status. Die geordneten Events enthalten exakte Request-Flags, die
nicht-geheime Keyquelle (`default` oder `adopted`), HTTP-Status, Body-Presence,
ob eine TNBU-Antwort dekodiert und ihr JSON geparst wurde, Response-Flags/-Typ
sowie den gewaehlten Cipher. Die CBC-Fallback-Kompatibilitaetsfelder bleiben
false, weil ein Heartbeat genau einen lokal gewaehlten Cipherrequest sendet.
Weder Key, Paketbody, MAC, Secret noch beliebiger roher Controllertext wird dort
aufbewahrt. Feste allowlist-basierte Responsetypen und begrenzte
Strukturmetadaten bleiben sichtbar; daraus folgt nicht, dass saemtliche vom
Controller gelieferten Werte auf jeder Produktoberflaeche abwesend sind. Ein
HTTP `404` oder ein fehlerhaftes HTTP `200` ist als fehlgeschlagener Austausch
sichtbar; ein leeres HTTP `200` ist ein erfolgreicher Heartbeat ohne dekodierte
Protokollevidenz.

Transport-, Response-Decode-, JSON-Syntax-, Dispatcher- und alle Nicht-200-
Ergebnisse erhoehen denselben begrenzten Backoff. Fehlende, nicht-stringfoermige
und unbekannte `_type`-Werte sind Dispatcherfehler mit festem Text; nicht
vertrauenswuerdiger Discriminator-Text wird nicht gespeichert. Nur ein leeres
HTTP 200 oder ein 200 mit akzeptiertem Controllerobjekt setzt die Fehlerfolge
zurueck. Ein Fehler loest niemals einen zweiten Request mit anderem Cipher aus.
Die aktuelle Produktabnahme wird durch die benannten Produktrequirements
definiert; sie verlangt keine prozessinterne Controllerreise.

## Grenze zwischen Evidence und Produkt

Firmware- und Reverse-Engineering-Matrizen sind Researchinput, keine
Produktregistry. Sie bleiben im Companion-Researchprotokoll und erzeugen fuer
sich allein weder Runtime-Verhalten, Tests noch Kompatibilitaetsclaims.

Nur vom Produktvertrag verlangtes Verhalten wird implementiert. Unbekannte
Buildidentitaeten, Profil-Overrides, nicht unterstuetzte Controllerwerte und
nicht erkannte Kommandos fallen geschlossen aus, statt Verhalten eines
Nachbargeraets zu uebernehmen. Controller-Intervallhinweise gelten ebenfalls
nur ueber eine exakt gepruefte Built-in-Identitaet; andernfalls bleibt der vom
Betreiber gewaehlte Takt aktiv.

## Adoption

Minimaler Ablauf:

1. Stub startet im Factory-/Default-Wire-State `1`.
2. Stub sendet Discovery und/oder Inform mit Default-Key.
3. Controller zeigt `Pending Adoption`.
4. Nach Klick auf Adopt antwortet der Controller mit `_type: "setparam"`.
5. `mgmt_cfg` enthaelt Werte wie `authkey`, `cfgversion`, `stun_url`, `mgmt_url`, `use_aes_gcm`.
6. Stub speichert `authkey`; der unmittelbar folgende Request nutzt diesen Key,
   GCM und Wire-State `0`.
7. Im erfassten Pfad schliesst ein akzeptiertes `noop` den Uebergang zum
   Managed-Wire-State `2` ab.
8. Das gespeicherte Capture enthaelt spaeter `setparam(system_cfg)` und `noop`.
   Weitere Command-Typen stammen aus statischer Analyse, nicht aus diesem Ablauf.

Das Inform-Wire-Vokabular ist exakt:

| Wert | Name | Produktbedeutung |
| --- | --- | --- |
| `0` | `Unknown` | Ein `setparam` enthielt einen Authkey oder aenderte die aktive Inform-URL; diese Identitaet wurde noch nicht durch eine spaetere erfolgreiche Antwort bestaetigt. |
| `1` | `Default` | Factory- oder Pending-Adoption-Identitaet mit oeffentlichem Kompatibilitaets-Key. |
| `2` | `Managed` | Ein erfolgreiches `noop` oder `setparam` hat die aktive Management-Identitaet bestaetigt. |
| `3` | `Selfrun` | Nur Statusvokabular; daraus folgt keine automatische Lifecycle- oder Host-Aktion. |
| `4` | `Upgrade` | Nur Statusvokabular; daraus folgt keine Firmware-, Service- oder Host-Aktion. |

Der dynamisch beobachtete Hauptpfad lautet damit `1 -> 0 -> 2`, nicht direkt
`1 -> 2`. Die statische Firmwareanalyse ergaenzt nur die allgemeine Regel: Ein
erfolgreiches `noop` oder `setparam` setzt zuerst Managed; ein `setparam` mit
Authkey oder eine Inform-URL-Aenderung fuehrt den Reporter danach wieder zu
Unknown. Das Capture belegt keine Restart-, Same-Identity- oder keylose
`noop`-Semantik.

Fuer den in UXG Pro 5.0.16 beobachteten und in UDM Pro SE 5.1.19 bestaetigten
Gateway-Pfad ist der oeffentliche Kompatibilitaets-Key der 16-Byte-Wert mit der
Hexdarstellung `ba86f2bbe107c7c57eb5f2690775c712`. Eine unabhaengige
Known-Answer-Berechnung zeigt die Gleichheit mit MD5 ueber den vier ASCII-Bytes
`ubnt`; dies beweist weder eine allgemeine Firmware-Key-Derivation noch den Key
fuer andere Firmwarefamilien.

Der Daemon waehlt diesen Wert explizit nur, solange der validierte lokale State
die Pending-/Default-Wire-Identitaet abbildet. Dazu gehoert URL-only Advanced
Adoption als keyloser `provisioning`-State; er sendet weiterhin `default=true`
und `state=1`. Der Low-Level-Inform-Client setzt den oeffentlichen Wert niemals
als Ersatz fuer einen fehlenden Key ein, und dieselben Bytes werden als
individueller Adopted-Key abgelehnt. Korrupter State, Adopted-State ohne
gueltigen individuellen Key und Failed-State ohne Key stoppen vor jedem
HTTP-Request.

Die Behandlung von Controllerantworten bleibt bewusst klein:

- Akzeptierte `setparam`-Daten duerfen nur begrenzte Adoption-Identity, Cipher,
  Konfigurationsversion und validierten Inform-URL-State aktualisieren.
- `system_cfg` ist nicht vertrauenswuerdiges Provisioning-Material und wird nie
  auf Host-User, Services, Firewall, Routen, Zertifikate oder Netzwerk angewandt.
- `_type: "upgrade"`, `_type: "cmd"` und unbekannte Responsetypen werden
  ignoriert, ohne ein Kommando auszufuehren oder die advertierte Produktversion
  zu veraendern.
- `noop`-Metadaten duerfen nur begrenztes prozesslokales Reporting und
  Scheduling beeinflussen. Sie duerfen weder die Hostuhr setzen noch
  Hostautoritaet erlangen.

Die Heartbeat-Policy bleibt lokal. Controller-Intervallhinweise brauchen eine
exakt gepruefte Built-in-Identitaet und werden auf Betreibergrenzen begrenzt;
jede andere Identitaet behaelt den vom Betreiber gewaehlten Basistakt.
Transport-, Decode-, JSON-, Dispatcher- und Nicht-200-Fehler nutzen begrenzten
Backoff. Nur ein leeres HTTP 200 oder ein akzeptiertes Controllerobjekt setzt
den Fehlerzaehler zurueck.

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
| `if_table[].management_vlan` | optionale konfigurierte Management-VLAN-Metadaten |
| `if_table[].rx_bytes`, `tx_bytes`, Paket-/Error-Counter | kumulative physische Counter, vom ausgewaehlten Uplink gespiegelt |
| gemessene `if_table[].bytes-r`, `rx_bytes-r`, `tx_bytes-r` | Raten des ausgewaehlten Uplinks in Byte/s fuer den normalisierten aktuellen Uplink |
| gemessene `if_table[].rx_rate`, `tx_rate` | Uplink-Raten in Bit/s fuer die Live-Statistik von Network 10.5 |
| `ethernet_table` | Controller-seitige Ethernet-/Portzahl-Tabelle |
| `port_table` | Switch-Ports |
| `port_table[].speed` | Port-Speed in Mbps, z.B. `1000` oder `10000` |
| `port_table[].media` | Medienkennung, z.B. `GE` oder `SFP+` |
| gemessene Uplink-`port_table[].rx_rate`, `tx_rate` | physische RX-/TX-Rate in Bit/s; bei Downstream-Ports nicht gesendet |
| `port_table[].mac_table` | gesehene Clients/VMs, optional mit `hostname`- und `ip`-Metadaten |
| `port_table[].last_connection` | direkte Downstream-Infrastruktur oder zuletzt verbundener Client pro Port |
| Top-Level-`uplink` | Management-Interface-Selektor wie `eth0`; UniFi Network kombiniert ihn mit `port_table[].is_uplink` zum aktuellen Uplink-Objekt |
| `uplink.speed` | Link-Speed des ausgewaehlten aktuellen Uplinks in Mbps; Quelle fuer die Switch-Spalte Experience |
| `last_uplink.uplink_mac` | explizite direkte Upstream-MAC fuer Switch-Payloads |
| `last_uplink.uplink_remote_port` | optionale konfigurierte oder per LLDP gelernte Upstream-Portnummer |
| Root-`bytes`, `rx_bytes`, `tx_bytes` | kumulative Traffic-Summary des ausgewaehlten Uplinks |
| Root-`bytes-r`, `rx_bytes-r`, `tx_bytes-r` | Raten des ausgewaehlten Uplinks in Byte/s (`B/s`) |
| Root-`rx_rate`, `tx_rate` | Raten des ausgewaehlten Uplinks in Bit/s |
| `sys_stats` | CPU/RAM/Load |

Mixed-Speed-Switch-Profile sollten das komplette physische Portlayout in
`port_table` melden. `USW-Pro-XG-48` wird zum Beispiel mit 16 2.5G-RJ45-Ports,
32 10G-RJ45-Ports und vier 25G-SFP28-Ports modelliert. Die Management-Speed in
`if_table` kommt vom gewaehlten Uplink-Port.

Die kumulativen Byte-, Paket- und Error-Counter des ausgewaehlten physischen
Uplinks werden zusaetzlich in die vom `uplink` benannte Management-`if_table`-
Zeile gespiegelt. Das ist die native Switch-Form und behaelt die physische
Richtung. Bei expliziter Ratenmessung erhaelt diese Zeile gemessene `*-r`-Raten
in Byte/s sowie `rx_rate`/`tx_rate` in Bit/s; der ausgewaehlte `port_table`-
Eintrag erhaelt ebenfalls die gemessenen Bit/s-Felder. Network 10.5 liest seine
normalisierte Live-Uplink-Rate aus diesem ausgewaehlten Interface-Pfad. Ohne
echte Messquelle werden die Management-Raten und ausgewaehlten Port-Bit/s-
Felder weggelassen und niemals synthetisch erzeugt. Legacy-synthetische
`port_table`-`*-r`-Werte sind nur Stub-Kompatibilitaetsdaten. Sie sind keine
gemessene Telemetrie und erfuellen `PR-METRICS` nicht; die Abnahme
nutzt ausschliesslich die ausgewaehlte lokale Interfacequelle einschliesslich
ihrer echten Nullsamples.

Die Switch-Root-Traffic-Summary wird nur bei aktiviertem Rate-Reporting
gesendet und kommt aus demselben ausgewaehlten Uplink. Die Root-Richtung ist
gegenueber den physischen Interface-Countern gedreht: Physisches Uplink-RX wird
zu Root-TX (Download), physisches Uplink-TX zu Root-RX (Upload). Die erste
Probe, ruhende/getrennte Links und Counter-Resets melden null; Traffic, der den
ausgewaehlten Uplink nicht durchlaeuft, ist nicht Teil dieser Summary.

Bei Switches ist der Inform-Wire-Wert der String-Selektor `uplink: "eth0"`
und kein Objekt. UniFi Network normalisiert die ausgewaehlte `port_table`-Zeile
zum aktuellen REST-`uplink`-Objekt. Network 10.4.57 behaelt darin die
kumulativen Root-Zaehler, laesst Root-Ratenfelder in der normalisierten Antwort
aber weg. Network 10.5.62 uebernimmt die gemessenen Interface-Raten in Byte/s
und Bit/s in den aktuellen Uplink und die offizielle Latest-Statistics-API;
dessen Root-Felder bleiben null oder fehlen. `source_interface` bleibt lokale
Payload-Metadaten in der logischen und der Port-Ansicht; es ist keine
Controller-Beobachtung.

Profile sollen Hardware-Form behalten. Lab-Zuweisungen wie "dieser Port ist
WAN", "dieser Port ist LAN" oder "dieser Port stellt Backup-WAN dar" gehoeren
in `port_overrides[].role` und `port_overrides[].network_group`.

`management_lan` wird zuerst als sichere Payload-Metadaten modelliert. Die
Konfiguration kann Switch-Management zusaetzlich auf ein bereits existierendes
VLAN-Interface legen. In diesem Modus wird dessen IPv4-Adresse als
Management-IP gemeldet, Discovery ueber dieses Interface gesendet und
Inform-HTTP lokal an dieselbe Adresse gebunden. Der Dienst legt weiterhin keine
getaggten Host-Interfaces an und uebernimmt keine VLAN-Aenderungen aus
Controller-Provisioning.

Bei alten Lab-Laeufen fuehrte ein fehlendes oder zu kleines `uptime` in
`mac_table` zu Controller-Problemen. Deshalb sollte jeder MAC-Tabelleneintrag
ein plausibles `uptime` besitzen. Inform-Payloads sollten zusaetzlich
monoton steigende Top-Level-`uptime`, `sys_stats.uptime`,
`system-stats.uptime` sowie Gateway-`time_ms`-/`uptime_str`-Felder melden,
damit UI-Freshness im Controller nicht von einem statischen First-Boot-Payload
abhaengt. Synthetische Ports ohne OS-Quelle duerfen Kompatibilitaets-
Freshnessdaten behalten, diese werden aber weder als Produktmetrik beschrieben
noch so getestet. Ports mit real beobachtetem Interface behalten die echten
OS-Counter und -Raten.

Direkte Infrastruktur und Client-MAC-Zeilen werden getrennt aufgeloest. Der
aktive `uplink_port` waehlt die Upstream-Kante, die Switch-Payloads ueber
`uplink` und `last_uplink` melden; Infrastruktur an anderen Ports bleibt
downstream. Konfigurierte Nachbarn gewinnen vor LLDP, waehrend FDB-
Beobachtungen niemals die Richtung bestimmen. `remote_port` liefert optional
`uplink_remote_port`. Reale Geraete- und Interface-MACs bleiben bei dieser
Projektion unveraenderte Identitaeten.

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
- `UXG`: experimentelles UniFi-Gateway-Lite-Identitaetsprofil, verfuegbar ueber
  das Profil `uxg-lite`. Es meldet Device-Type `uxg`, Modell `UXG`, Firmware
  `5.0.16.30689` und zwei 1G-Ports `LAN` und `WAN`.
- `UXGPRO`: experimentelles UniFi-Next-Generation-Gateway-Pro-
  Identitaetsprofil. Es meldet Device-Type `uxg`, Modell `UXGPRO`, Firmware
  `5.0.16.30689`, zwei 1G-RJ45-Ports `WAN` und `LAN` sowie zwei 10G-SFP+-
  Ports `WAN2` und `LAN2`. Der Default-Uplink bleibt `WAN`; SFP+-Internet-
  Labs werden ueber `uplink_port` und `port_overrides` remapped.
- `UCGF`: experimentelles UniFi-Cloud-Gateway-Fiber-Identitaetsprofil,
  verfuegbar ueber das Profil `ucg-fiber`. Es meldet Device-Type `udm`,
  Firmware `5.0.16`, vier 2.5G-RJ45-LAN-Ports, einen 10G-RJ45-Port `WAN2`,
  einen 10G-SFP+-Port `WAN` und einen 10G-SFP+-LAN-Port.

`UGW3`, `UXG`, `UXGPRO` und `UCGF` sind in diesem Release nur
Identitaets-/Profil-Stubs. Fuer einen vollstaendigen Gateway-Payload fehlen
weiterhin WAN/LAN-State, Routing, DHCP, Firewall, DPI und Health-Felder. Der
aktuelle Gateway-Payload sendet Gateway-Tabellen wie `network_table`, eine
read-only physische `port_table`, `uplink`, `uplink_table`,
`config_port_table` und `ethernet_overrides`. Explizit konfigurierte
WAN-/LAN-Zuweisungsmetadaten koennen gespiegelt werden; sie werden aber nicht
aus Controller-Provisioning gelernt und nicht auf Host-Netzwerk angewandt.
Gateway-Modelle wie `UGW4`, andere Cloud-Gateway-Geraete und EFG erst spaeter
pruefen.
