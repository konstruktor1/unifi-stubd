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

## Registry fuer exakte Build-Policies

Das Produkt enthaelt zwei typisierte Registries. `ProtocolProfile` bindet
vollstaendige statische Evidenz, nicht ein Modelllabel oder Firmware-
Versionspraefix, an eine versionsunabhaengige Verhaltens-ID.
`ProtocolBehaviorProfile` komponiert getrennt Wire-, Discovery-, Feldgruppen-,
Command-, positive Capability- und Presence-Policies:

| Evidenz | Exakte Buildklasse | Wire | Discovery | Commands | Feldgruppen / Presence | Positive Capabilities |
| --- | --- | --- | --- | --- | --- | --- |
| `DT-001` | UXG Pro 5.0.16 | gemeinsamer sicherer zlib-TNBU-Subset | unsupported bis zu exakten Gatewayframes | Top-Level 7 / direkt 70 | UXG Producer-/Request-Gates | Gateway-Management plus Producer-Telemetrie |
| `DT-002` | UDM Pro SE 5.1.19 | gemeinsamer sicherer zlib-TNBU-Subset | unsupported bis zu exakten Gatewayframes | Top-Level 7 / direkt 65 | UDM Saved-Config-/Probe-Gates | Gateway-Management plus Saved-Config-/Probe-Vokabular |
| `DT-008` | Mainline BCM 7.4.1 | gemeinsamer sicherer zlib-Subset; Snappy deaktiviert | gemeinsamer Mainline-v2-Builder | Top-Level 7 / direkt 40, Hostmutation abgelehnt | gemeinsame Mainline-Gruppen plus BCM-Presence | Switch/Mainline plus BCM-Tabellenvokabular |
| `DT-009` | Mainline MSCC 7.4.1 | gemeinsamer sicherer zlib-Subset | gemeinsamer Mainline-v2-Builder | gleiche Direkt-40-Policy | gemeinsame Mainline-Gruppen plus MSCC-Presence | Switch/Mainline plus MSCC-Plattformvertrag |
| `DT-010` | Switch-WAN 7.3.109 | gemeinsamer sicherer zlib-Subset | getrennter WAN-v2-Alternate-Builder-Vertrag | Top-Level 7 / direkt 41, Hostmutation abgelehnt | WAN-/Modell-/Source-Gates | source-gated WAN-/RPS-Vokabular; keine Aktionsautoritaet |
| `DT-012` | aktuelles `usmini2` 2.1.6 | Raw-JSON-TNBU | gemeinsamer aktueller ESP-Directed-Broadcast | belegte Top-Level 5; direktes `cmd` unsupported | gemeinsame Low-Resource-Gruppen plus usmini2-Presence | aktuelle ESP-Basis; keine positive PSE-Aussage |
| `DT-013..015` | aktuelle Ultra/M25G 2.1.8 | gemeinsam zlib mit Raw-Fallback | gemeinsamer aktueller ESP-Directed-Broadcast | gleiche Top-Level-5-Policy; direktes `cmd` unsupported | gemeinsame Low-Resource-Gruppen plus getrennte Presence | gemeinsame aktuelle ESP-Basis; positives PSE-Vokabular nur fuer Ultra und M25G8 |

Die IDs beschreiben semantische Vertraege und enthalten keine Artefaktversion.
Gleiche vollstaendige Kompositionen koennen keine doppelten Verhaltens-IDs
erhalten; einzelne Policybausteine werden ueberall dort wiederverwendet, wo die
Evidenz dieselbe Semantik belegt. Fehlende Capability-Flags deaktivieren nur
unbelegtes Produktverhalten; sie behaupten nicht, dass Hardware fehlt. Fuer
M25G5 liefert diese App-Analyse insbesondere keine positive PSE-Policy, die
externe Hardware bleibt jedoch unbekannt.

Jeder ausgewaehlte Baustein legt Achse/Selektor, die belegenden `DT-*`-Records
und das gewonnene Prinzip offen. Ausfuehrbaren Produktnachweis braucht er nur,
wenn ein benanntes Produktrequirement sein Verhalten auswaehlt.
Plattformselektoren sind auf die typisierten Architekturklassen aus
`device-types.tsv` begrenzt; jeder andere Wert faellt geschlossen aus.

`DT-003..007` und die M487-Komponente `DT-011` sind bekannte
Evidenzidentitaeten, deren Protokollsemantik unsupported oder `UNKNOWN` bleibt;
ihre sechs Policyachsen stehen deshalb alle auf `unsupported`. M487 und
aktuelles usmini2 teilen einen Firmwarecontainer, aber weder Agent noch
Wire-Policy; ebenso teilen BCM und MSCC ein aeusseres Firmwarebild, behalten
aber exakte Komponentenidentitaeten und getrennte vollstaendige
Verhaltenskompositionen. Leere Hashes, Ein-Byte-Hashaenderungen, benachbarte
Versionen und gemischte Komponenten fallen deshalb geschlossen aus, statt ein
Nachbarprofil zu waehlen. Buildmarker sind exakt beobachtete Firmware-, Rootfs-
oder Komponentenmarker und keine normalisierten Produktlabels.

Die Registries sind Metadaten und eine strikte Lookup-Grenze. Die Anwendung jedes
Selektors in Inform, Discovery, Command-Dispatch und Payload-Rendering gehoert
zu den jeweiligen Implementierungsrequirements; die Registries allein veraendern
weder die bestehenden benutzersichtbaren Formprofile noch den Runtime-Traffic.

### Exakte Controller-Intervall-Policy

`REQ-LIFE-010` ist ein getrennter, expliziter Consumer der Evidence-Registry.
Positive Controller-Intervallhinweise sind nur aktiv, wenn Built-in-Quelle,
Profilname, Modell, Version und Boardidentitaet ein exaktes Tupel bilden:

| Built-in-Profil | Exaktes Modell | Exakte Version | Intervall-Evidence-ID | Policy fuer positives `noop.interval` |
| --- | --- | --- | --- | --- |
| `us8` | `US8` | `7.4.1.16850` | `DT-008` | aktiv; 30 Sekunden Firmwareminimum |
| `us8p60` | `US8P60` | `7.4.1.16850` | `DT-008` | aktiv; 30 Sekunden Firmwareminimum |
| `us16p150` | `US16P150` | `7.4.1.16850` | `DT-008` | aktiv; 30 Sekunden Firmwareminimum |
| `us24p250` | `US24P250` | `7.4.1.16850` | `DT-008` | aktiv; 30 Sekunden Firmwareminimum |
| `us48p500` | `US48P500` | `7.4.1.16850` | `DT-008` | aktiv; 30 Sekunden Firmwareminimum |
| `uxgpro` | `UXGPRO` | `5.0.16.30689` | `DT-001` | aktiv; 1 Sekunde Firmwareminimum |
| `ucg-fiber` | `UCGF` | `5.0.16` | `DT-004` | fail-closed; exakte Evidenz ist bekannt, aber kein Firmwareminimum katalogisiert |
| `ugw3` | `UGW3` | `4.4.57.5578372` | `DT-007` | fail-closed; exakte Evidenz ist bekannt, aber kein Firmwareminimum katalogisiert |
| `uxg-lite` | `UXG` | `5.0.16.30689` | `DT-006` | fail-closed; exakte Evidenz ist bekannt, aber kein Firmwareminimum katalogisiert |
| `us16xg` | `US16XG` | `7.4.1.16850` | nicht gebunden | fail-closed; keine exakte Intervall-Policy-Bindung |
| `usaggpro` | `USAGGPRO` | `7.4.1.16850` | nicht gebunden | fail-closed; keine exakte Intervall-Policy-Bindung |
| `usw-pro-xg-48` | `USWProXG48` | `7.4.1.16850` | nicht gebunden | fail-closed; keine exakte Intervall-Policy-Bindung |

Diese Entscheidungen betreffen nur Intervallhinweise. Sie stufen weder die
Profilreife hoch noch aktivieren sie saemtliches Protokollverhalten einer
Evidence-ID. Insbesondere gehoert das 30-Sekunden-Minimum zu jedem aufgefuehrten
Board-/Profiltupel; es ist kein buildweiter `DT-008`-Default. Externe Profile,
benachbarte Versionen, Modell- oder Versions-Overrides und aehnliche Identitaeten
fallen geschlossen aus, statt eine Nachbarentscheidung zu uebernehmen.

Fuer ein aktiviertes Tupel ist die effektive Untergrenze
`max(Firmwareminimum, Betreiberminimum)` und die effektive Obergrenze
`min(Betreibermaximum, 300 Sekunden)`. Bei jedem deaktivierten Tupel wird ein
positiver Hinweis als ignoriert erfasst und der vom Betreiber gewaehlte
Basistakt bleibt aktiv. Die Konfiguration scheitert an der Validierung, wenn
diese effektive Obergrenze unter der effektiven Untergrenze liegt. Dry-Run-Plan
und bereinigter Status zeigen, ob Hinweise
unterstuetzt sind, die Evidence-ID, Firmware- und effektives Minimum oder den
festen Fail-closed-Grund.

Diese LIFE-010-Implementierung existiert nur im aktuellen Dirty-Worktree-
Produktkandidaten. Reviewte Source-Hash-Bindung, T4-Abnahme derselben Bytes auf
der vorgesehenen Controller-Testmaschine, separate G3-Erreichbarkeit aus
`main` und Release-Evidence bleiben offen; dies ist keine T4-, G3- oder
Release-Aussage.

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

Evidence-abgeglichene UXG-Pro-5.0.16-Behandlung:

| Evidence-Klasse | Phase | Request/Response-Form | Stub-Verhalten |
| --- | --- | --- | --- |
| Dynamisch DT-001 | Vor Adoption | Firmware-Informs nutzen den Default-Key und bekommen HTTP `404`, bis ein Admin `Adopt` klickt. | `default=true`, `state=1` behalten. |
| Dynamisch DT-001 | Erster akzeptierter Inform | Controller antwortet HTTP `200` mit `_type: "setparam"` und `mgmt_cfg`. | `authkey`, `cfgversion` und `use_aes_gcm=true` parsen und persistieren; Wire-State `0` betreten. |
| Dynamisch DT-001 | Key-Wechsel | Firmware sendet den naechsten Inform sofort mit Adopted-Key und AES-GCM. | `default=false`, `state=0` senden. In diesem Capture gab es keinen Restart. |
| Dynamisch DT-001 | Akzeptierter Adopted-Inform | Controller antwortet mit `_type: "noop"`. | Zu `default=false`, `state=2` (`Managed`) wechseln. Die statische Analyse erlaubt auch erfolgreiches `setparam`; ein Authkey-Feld oder eine tatsaechliche Inform-URL-Aenderung fuehrt danach wieder zu `0`. |
| Dynamische DT-001-CAP-05-Form plus Produktsicherheit | System-Konfiguration | Die erhaltene Zusammenfassung zeigt ein weiteres `_type: "setparam"` mit Top-Level-`cfgversion`, vorhandenem `mgmt_cfg` ohne offengelegten Inhalt und `system_cfg`. | Top-Level-`cfgversion` nicht als Adoption-Store-Eingabe uebernehmen und nur begrenzte `system_cfg`-Strukturmetadaten behalten; keine Host-User, Firewall, Routen, Zertifikate, Tokens oder Secrets anwenden. |
| Statisches Vokabular plus Produktsicherheit | Upgrade-Status | Firmware kennt `_type: "upgrade"`; diese Zeile gehoert nicht zu DT-001. | Die Antwort nur als ignorierten Status erfassen. Advertisierte Betreiber-Version, persistierten Lifecycle-State, Services, Firmware und Host nicht veraendern. |
| Statische Command-Evidence plus Produktpolicy | WireGuard-Key-Anfrage | Firmware kennt `_type: "cmd"`, `cmd: "wg-genkeypair"` und eine `request_id`; diese Zeile gehoert nicht zu DT-001. | Ein X25519-Keypaar im Speicher erzeugen und als firmwarekompatible Notification an den naechsten regulaeren Inform anhaengen: `inform_as_notif=true`, `notif_reason="cmd"` und ein `notif_payload` mit `event_string`, korrelierter `request_id` sowie `params.private_key`/`params.public_key`. Der aktuelle Core leert den Slot nur, wenn ein HTTP `200` auch seinen Delivery-Punkt erreicht; die evidenzgebundene Every-HTTP-200-Regel bleibt offen. Keys und Request-ID weder persistieren noch loggen. |
| Statische Firmwareanalyse | Steady State | Firmware akzeptiert `_type: "noop"` mit `interval`, `immediate`, Blocklisten und optional einem String `server_time_in_utc`; diese Zeile gehoert nicht zu DT-001. | Bei einem exakt unterstuetzten Tupel ein positives Intervall von der strengeren Firmware-/Betreiberuntergrenze bis zur kleineren Betreiber-/300-Sekunden-Obergrenze begrenzen. Andernfalls den Hinweis ignorieren und die Betreiberbasis behalten. Den Heartbeat-Fehler-Backoff zuruecksetzen. `immediate=true` plant genau einen begrenzten, von null verschiedenen Folge-Inform, ohne den beibehaltenen Normaltakt zu ersetzen. Jedes `noop` ersetzt die vollstaendige prozesslokale Blockauswahl. Ein Controller-Zeitstring verschiebt nur die in spaeteren Stub-Payloads gemeldete Zeit. |

Die Blockauswahl ist eine producer-lokale Policy und keine Command-Oberflaeche.
Jedes akzeptierte `noop` ersetzt die vorherige Include-/Exclude-Auswahl,
einschliesslich leerer neuer Listen. Bekannte Producer pruefen nur ihre exakten
Blockbezeichner; UDAPI-Deskriptoren bleiben begrenzte Statusmetadaten und werden
niemals ausgefuehrt.

`server_time_in_utc` wird nur als Stringmetadatum eines `noop` akzeptiert. Der
Stub behaelt einen prozesslokalen Controller-minus-Device-Offset und wendet ihn
auf die gemeldete Payload-Zeit an. Er ruft keinen Host-Zeitdienst auf, setzt
weder Systemuhr noch Host-Zeitzone.

Das erste im Lab beobachtete `mgmt_cfg` enthielt `capability`,
`selfrun_guest_mode`, `cfgversion`, `led_enabled`, `stun_url`, `mgmt_url`,
`authkey`, `use_aes_gcm=true` und `report_crash`. Es enthielt kein neues
`inform_url`; die Firmware nutzte die bestehende Inform-URL weiter und wechselte
nur Key/Cipher-Kontext.

GW-14 dokumentiert einen statischen Parse-/Install-/Persist-/Next-Use-Pfad fuer
`mgmt_cfg.inform_url`; kein erhaltener dynamischer Capture loest eine
Inform-URL-Aenderung aus. Diese Researchbeobachtung erzeugt keine eigene
Produkttestpflicht. Der aktuelle Produktvertrag verlangt nur begrenzte
Adoption-Identity und die Host-Sicherheitsgrenze.

Belegt ist nur der aktivierende Wert `use_aes_gcm=true`. Ein vom Controller
geliefertes `false` stuft eine bestehende Stub-Session nicht zurueck;
firmwaregleiche Downgrade-Semantik bleibt wegen fehlender Evidenz ausdruecklich
zurueckgestellt.

Das adoptierte `system_cfg` war ein JSON-String mit den Top-Level-Keys
`ubntconf` und `udapi`. Das `udapi`-Objekt enthielt Interfaces, Services,
System-User, Firewall-Sets/-Filterketten/-Settings, statische Routen und
Radius-Profile. Das ist Provisioning-Datenmaterial fuer ein echtes Gateway,
aber kein sicherer Satz Host-Anweisungen fuer `unifi-stubd`.

`wg-genkeypair` ist das einzige Controller-Kommando mit einem ausfuehrbaren
Ergebnis. Es nutzt die prozessinterne X25519-Implementierung von Go und fuehrt
weder Shell-Kommandos, Dateischreibvorgaenge, Service-Aktionen noch Host-
Netzwerkmutationen aus. Fehlende Request-IDs, Argumente, Parameter, ungueltige
Requests und alle anderen Kommandos werden ignoriert.

Die Heartbeat-Policy bleibt lokal. Paketierte Defaults starten mit 10 Sekunden
und konfigurieren Betreibergrenzen von 5 bis 60 Sekunden; diese Grenzen
aktivieren fuer sich allein keine Controller-Hinweise. Zuerst muss ein exaktes
LIFE-010-Tupel sein Firmwareminimum liefern. Der Fehler-Backoff ist auf 300
Sekunden begrenzt und verwendet 20 Prozent symmetrischen Jitter. Transport-,
Response-Decode-, JSON-, Dispatcher- und Nicht-200-Fehler verdoppeln die
effektive Wartezeit bis zur lokalen Obergrenze. Nur ein leeres HTTP 200 oder ein
200 mit akzeptiertem Controllerobjekt setzt den Fehlerzaehler zurueck. Liegt der
konfigurierte Normaltakt ueber dem Backoff-Maximum, bleibt er die Untergrenze,
damit ein Fehler Inform-Traffic niemals beschleunigt.

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
dessen Root-Felder bleiben null oder fehlen. Der lokale Runtime-Status erfasst
Root, Management-`if_table`, logischen `uplink` und ausgewaehlten
`port_table`-Eintrag. `source_interface` bleibt als lokale Diagnosemetadaten in
der logischen und der Port-Ansicht.

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
