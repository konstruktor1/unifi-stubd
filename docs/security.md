# Security Notes

`unifi-stubd` ist fuer ein isoliertes Lab gedacht.

## Authkeys

Der UniFi `authkey` ist ein symmetrischer Schluessel fuer Inform-Payloads. Er darf nicht in Logs, Screenshots oder Git-Historie landen.

## Keine Host-Provisionierung

Der Controller darf nicht ungeprueft Host-Konfigurationen veraendern. Fuer den MVP gilt:

- `setparam` speichern.
- `noop` quittieren.
- Restart/upgrade/provisioning nur loggen.
- Keine Shell-Kommandos vom Controller ausfuehren.

## Netzgrenzen

Discovery und Inform gehoeren nur ins Lab- oder Management-Netz. Das Projekt sollte nicht auf produktiven VLANs mit fremden Controllern laufen.

## Personen- und Clientdaten

MAC-Tabellen, DHCP-Informationen, DPI-Daten und NetFlow koennen personenbeziehbare Metadaten enthalten. Beispiel-PCAPs gehoeren deshalb in `.gitignore` und sollten vor Weitergabe anonymisiert werden.

