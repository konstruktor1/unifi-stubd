# Security Notes

`unifi-stubd` ist fuer ein isoliertes Lab gedacht.

Lies vor dem Melden oder Veroeffentlichen sicherheitsrelevanter Details die
Repository-weite [Security Policy](../../SECURITY.md).
Private Meldungen koennen an `info@spinas.org` gesendet werden.

## Authkeys

Der UniFi `authkey` ist ein symmetrischer Schluessel fuer Inform-Payloads. Er darf nicht in Logs, Screenshots oder Git-Historie landen.

`adoption.env`, SSH-Hostkeys und Controller-API-Tokens duerfen nicht geteilt werden.

## Keine Host-Provisionierung

Der Controller darf nicht ungeprueft Host-Konfigurationen veraendern. Fuer den MVP gilt:

- `setparam` speichern.
- `noop` quittieren.
- Fuer Gateway-`system_cfg` nur bereinigte Metadaten wie Byte-Laenge und
  Top-Level-Keys erfassen.
- Restart-/Upgrade-/Provisioning-Kommandos als per Policy ignoriert markieren.
- Keine Shell-Kommandos vom Controller ausfuehren.

## Netzgrenzen

Discovery und Inform gehoeren nur ins Lab- oder Management-Netz. Das Projekt sollte nicht auf produktiven VLANs mit fremden Controllern laufen.

Die paketierte Linux-Lab-Konfiguration stellt den Adoption-SSH-Shim aus
Kompatibilitaetsgruenden auf `0.0.0.0:22` mit UniFi-Factory-Credentials bereit.
Das darf nur in einem isolierten Lab- oder Management-VLAN laufen; andernfalls
`ssh_listen` ueberschreiben.

## Personen- und Clientdaten

MAC-Tabellen, DHCP-Informationen, DPI-Daten und NetFlow koennen personenbeziehbare Metadaten enthalten. Beispiel-PCAPs gehoeren deshalb in `.gitignore` und sollten vor Weitergabe anonymisiert werden.
