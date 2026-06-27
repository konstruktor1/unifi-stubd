# OPNsense-API-Generator Referenz

Diese Seite dokumentiert die `unifi-stubd-opnsense`-Source-Datei und das
Generator-Verhalten. Fuer die Schritt-fuer-Schritt-Einrichtung auf einem
OPNsense-Host siehe das
[OPNsense-API-Generator How-to](opnsense-generator.md).

## Umfang

`unifi-stubd-opnsense` ist bewusst vom Daemon getrennt:

- Es liest eine bestehende `unifi-stubd`-YAML-Config.
- Es liest eine OPNsense-Source-YAML-Datei.
- Es fuehrt HTTP-`GET`-Requests gegen OPNsense aus.
- Es schreibt generiertes `unifi-stubd`-YAML nach stdout oder nach `-out`.
- Es laeuft nicht als Live-Sync-Dienst.
- Es aendert keine OPNsense-Interfaces, Routen, Firewall-Regeln, VLANs oder den
  laufenden `unifi-stubd`-Daemon.

## Kommando-Referenz

```sh
unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml
```

Flags:

| Flag | Erforderlich | Bedeutung |
| --- | --- | --- |
| `-config` | nein | Basis-`unifi-stubd`-YAML-Config. Default ist der Paket-Config-Pfad. |
| `-source` | ja | OPNsense-Source-YAML-Datei. |
| `-out` | nein | Generiertes YAML in diesen Pfad schreiben. Ohne `-out` wird nach stdout geschrieben. |
| `-validate` | nein | Basisconfig, Source-YAML und Credential-Laden pruefen, ohne API-Aufruf und ohne Output. |

## OPNsense-API-Aufrufe

Der Client baut URLs als `<base_url>/api/<path>` und sendet Basic Auth mit dem
konfigurierten API-Key als Username und dem API-Secret als Passwort.

Implementierte Lesezugriffe:

| Source-Einstellung | Endpoint |
| --- | --- |
| immer | `GET /api/interfaces/overview/interfaces_info` |
| Fallback fuer ein fehlendes gemapptes Interface | `GET /api/interfaces/overview/get_interface/<interface>` |
| `gateway_status: true` | `GET /api/routes/gateway/status` |

Responses sind auf 8 MiB begrenzt. Fehlermeldungen enthalten Endpoint-Pfade und
HTTP-Statuscodes, aber keine API-Key- oder Secret-Werte.

## Source-YAML-Felder

Top-Level der Source-Datei:

| Feld | Erforderlich | Default | Beschreibung |
| --- | --- | --- | --- |
| `base_url` | ja | keiner | OPNsense-WebGUI-/API-Basis-URL, zum Beispiel `https://127.0.0.1`. Muss `http` oder `https` verwenden und einen Host enthalten. |
| `api_key_file` | bedingt | leer | Datei mit dem rohen API-Key. Wird genutzt, wenn `api_key_env` nicht gesetzt oder leer ist. |
| `api_secret_file` | bedingt | leer | Datei mit dem rohen API-Secret. Wird genutzt, wenn `api_secret_env` nicht gesetzt oder leer ist. |
| `api_key_env` | bedingt | leer | Environment-Variable mit dem rohen API-Key. Hat Vorrang vor `api_key_file`, wenn nicht leer. |
| `api_secret_env` | bedingt | leer | Environment-Variable mit dem rohen API-Secret. Hat Vorrang vor `api_secret_file`, wenn nicht leer. |
| `ca_file` | nein | leer | PEM-CA-Bundle zur TLS-Pruefung von OPNsense. |
| `insecure_skip_verify` | nein | `false` | Erlaubt self-signed Lab-Endpunkte ohne Zertifikatspruefung. Nur bewusst verwenden. |
| `timeout_ms` | nein | `2000` | Request-Timeout in Millisekunden. Muss positiv sein. |
| `uplink_port` | nein | `0` | Generierter `uplink_port`-Hinweis. Wird nur angewendet, wenn die Basisconfig keinen `uplink_port` hat. |
| `gateway_status` | nein | `false` | Aktiviert Gateway-Status-API-Reads und WAN-Health-Hinweise. |
| `interfaces` | ja | keiner | Port-zu-Interface-Mappings. Muss mindestens einen Eintrag enthalten. |
| `wan_health` | nein | leer | Optionaler generierter `wan_health`-Block. Wird nur angewendet, wenn `wan_health.source` nicht leer ist. |

Credential-Regeln:

- Mindestens eine Key-Quelle und eine Secret-Quelle muessen konfiguriert sein.
- Environment-Variablen gewinnen gegen Dateien, wenn sie konfiguriert und nicht
  leer sind.
- Leere Credential-Werte werden abgelehnt.
- Credential-Werte werden nie in die generierte `unifi-stubd`-Config gerendert.

## Interface-Mapping-Felder

Jeder `interfaces[]`-Eintrag mappt einen dargestellten UniFi-Port auf ein
OPNsense-Interface:

| Feld | Erforderlich | Beschreibung |
| --- | --- | --- |
| `port` | ja | Eins-basierter dargestellter UniFi-Portindex. Muss positiv und in der Source-Datei eindeutig sein. |
| `interface` | ja | OPNsense-/FreeBSD-Interface wie `ixl0`, `igb0` oder `vtnet0`. Slashes werden abgelehnt. |
| `name` | nein | Generierter Port-Label-Override. |
| `role` | nein | Effektive Rolle wie `wan`, `wan2`, `lan`, `lan2` oder `unassigned`. Wird auf lowercase normalisiert. |
| `network_group` | nein | UniFi-Network-Group-Label, zum Beispiel `WAN`, `WAN2` oder `LAN`. |
| `portconf_id` | nein | Controller-Port-Profil-Zuweisungs-ID zum Spiegeln. |
| `networkconf_id` | nein | Controller-Network-Zuweisungs-ID zum Spiegeln. |
| `native_networkconf_id` | nein | Controller-Native-Network-Zuweisungs-ID zum Spiegeln. |
| `network_name` | nein | Controller-Network-Anzeigename zum Spiegeln. |
| `vlan` | nein | Controller-Anzeige-VLAN-ID zum Spiegeln. |
| `speed` | nein | Dargestellter Link-Speed-Override in Mbps. |
| `media` | nein | Dargestelltes Medium wie `GE`, `SFP+` oder `SFP28`. |

Der OPNsense-Interface-Name ist Source-Metadaten. UniFi-`ifname` bleibt
profilbasiert. UXG-Pro-Port 3 bleibt zum Beispiel `eth2` in controllerseitigen
Payloads, auch wenn das Source-Interface `ixl0` ist.

## Merge-Verhalten

Die Generierung startet mit der geladenen Basisconfig.

Port-Overrides:

- OPNsense-basierte Overrides werden aus `interfaces[]` erzeugt.
- Bestehende `port_overrides` aus der Basisconfig werden nach `port` gemerged.
- Basisconfig-Werte gewinnen feldweise.
- Nur generierte Ports bleiben erhalten.
- Die Ausgabe wird nach Port sortiert.

Top-Level-Felder:

- `uplink_port` aus der Source wird nur angewendet, wenn die Basisconfig
  `uplink_port: 0` oder keinen Wert hat.
- `wan_health` aus der Source wird nur angewendet, wenn `wan_health.source`
  nicht leer ist.

Gateway-Status:

- Gateway-Status-Werte werden nur auf Overrides angewendet, deren generierte
  oder basisconfig-basierte Rolle `wan` oder `wan2` ist.
- `wan_connected`, `wan_latency_ms` und `wan_uptime_percent` koennen generiert
  werden.
- LAN- und unassigned-Ports erhalten keine WAN-Health-Hinweise.

## Beispiel-Source-Datei

Siehe
[`lab/stub/configs/hosts/opnsense-api-source.example.yaml`](../../lab/stub/configs/hosts/opnsense-api-source.example.yaml)
fuer ein vollstaendiges bereinigtes Source-Beispiel.

## Troubleshooting

`opnsense base_url is required`:

- `base_url` in der Source-YAML setzen.

`api_key requires an env var or file` oder `api_secret requires an env var or file`:

- `api_key_file` und `api_secret_file` konfigurieren, oder `api_key_env` und
  `api_secret_env` setzen.

`returned HTTP 401`:

- Pruefen, ob API-Key und Secret korrekt sind und zu einem OPNsense-User
  gehoeren, der die abgefragten API-Seiten lesen kann.

`certificate signed by unknown authority`:

- Bevorzugt `ca_file` auf ein PEM-CA-Bundle setzen.
- Nur fuer isolierte Labs `insecure_skip_verify: true` setzen.

Generierte Config verwendet `eth2` statt `ixl0` als `ifname`:

- Das ist erwartet. `eth2` ist das UniFi-Profilinterface. `ixl0` wird als
  Source-Metadaten im generierten Port-Override gerendert.

Manuelle Basisconfig-Werte werden nicht ueberschrieben:

- Das ist erwartet. Bestehende `port_overrides` aus der Basisconfig gewinnen
  feldweise gegen generierte Werte.
