# Priorisierter Backlog

Zuletzt aktualisiert: 2026-07-11.

Dieser Backlog enthaelt die naechsten konkreten Produktarbeiten fuer
`unifi-stubd`. Die Roadmap bleibt die langfristige Feature-Liste. Die Eintraege
sind nach Umsetzungsprioritaet geordnet und muessen die zentrale
Sicherheitsgrenze erhalten: Controller-Daten duerfen Host-Netzwerk, Dienste,
Benutzer, Pakete, Routen oder Firewall nicht veraendern.

Der Aufwand ist relativ angegeben: `S` fuer eine fokussierte Aenderung, `M`
fuer eine moduluebergreifende Aenderung und `L` fuer Arbeiten mit
Controller-Integration und Fixtures.

| ID | Prioritaet | Massnahme | Abnahmekriterium | Aufwand |
| --- | --- | --- | --- | --- |
| B01 | P0 | **Abgeschlossen 2026-07-10:** Adoption-SSH-Logs bereinigen | Adoption-SSH-Logs verwenden feste Kommandobezeichnungen. Ein echter SSH-Regressionstest belegt, dass Keys und Passwoerter nicht in Logs, Antworten oder Authentifizierungsfehler gelangen, waehrend der Key persistiert bleibt; der bestehende Status-Regressionstest bleibt gruen. | S |
| B02 | P0 | **Abgeschlossen 2026-07-10:** Go-Build-Linie aktualisieren | GitHub- und Docker-Builds verwenden den aktuellen Go-1.25-Patch ohne festgeschriebene Patchversion oder `toolchain`-Direktive. `make check`, Race Detector, `make vulncheck` und ein Paket-Build laufen mit dem aktuellen Go-1.25-Release erfolgreich. | S |
| B03 | P1 | **Abgeschlossen 2026-07-10:** Interface-Observation korrekt aktualisieren | Jeder Runtime-Port-Build startet mit unveraenderter Betreiberkonfiguration und liest einen frischen Snapshot fuer MAC, IPv4, IPv6, Link-State, Speed, Medium und Counter. Explizite Werte behalten Vorrang; vollstaendige Null-Counter-Resets erreichen den Payload. | M |
| B04 | P1 | **Abgeschlossen 2026-07-10:** Persistenz und Prozessende haerten | `SIGTERM` und `SIGINT` stoppen den Daemon sauber und geben Runtime-Ressourcen frei. Adoption und Status verwenden temporaere Dateien im selben Verzeichnis, Datei- und Verzeichnissynchronisation sowie atomisches Umbenennen. Regressionstests fuer unterbrochene Schreibvorgaenge bestehen. | M |
| B05 | P1 | **Abgeschlossen 2026-07-10:** Echten Controller-Integrationstest wiederherstellen | Der gepinnte Docker-Controller-Gate bootstrapt leere Volumes und prueft Discovery, Pending Adoption, bereinigtes `setparam`, persistierten Auth-State, mindestens zwei AES-GCM-Informs und drei stabile Connected-Samples fuer Switch- und Gateway-Fixtures. | L |
| B06 | P1 | **Abgeschlossen 2026-07-10:** Sicheren Controller-Command-Lifecycle vervollstaendigen | Nur ein argumentfreies `wg-genkeypair` mit Request-ID wird verarbeitet. X25519-Keypaar und Request-ID bleiben im Speicher, das korrelierte Ergebnis wird bis zu einem erfolgreichen Inform erneut gesendet und Logs/Status enthalten nur bereinigte Zaehler. Der gepinnte Controller-Gate belegt genau ein verarbeitetes Kommando, ein ausgeliefertes Ergebnis, kein Duplikat, kein ausstehendes Ergebnis und nach einem Stabilitaetsfenster keine wiederkehrende Task-Schleife. | M/L |
| B07 | P2 | **Abgeschlossen 2026-07-10:** Heartbeat- und Fehlerstrategie einfuehren | Positive Controller-Intervallhinweise steuern nach Begrenzung auf betreiberseitig konfigurierte Werte den dynamischen Timer. Transportfehler verdoppeln den effektiven Takt mit begrenztem symmetrischem Jitter; ein abgeschlossener HTTP-Austausch setzt die Fehlerfolge zurueck. Unit- und Live-Daemon-Tests decken Begrenzung, Cap, Jitter, Reset, Status und ungueltige Konfiguration ab. Der gepinnte Controller-Gate belegt, dass sein positiver Rueckgabewert zum begrenzten effektiven und naechsten Takt wird, ohne ausstehende Fehler oder Backoff. | M |
| B08 | P2 | **Abgeschlossen 2026-07-10:** Coverage-Target reparieren | `make coverage` instrumentiert alle produktiven `internal/...`-Pakete ueber die externen `tests/...`-Suites, meldet den aggregierten Wert und erzwingt lokal sowie in CI anfangs mindestens 60 Prozent. | S |
| B09 | P2 | **Abgeschlossen 2026-07-11:** Profilreife evidenzbasiert ausweisen | Experimentelle Gateway-Profile melden `experimental`. `tested` erfordert mindestens einen konkreten Eintrag in `validated_controller_versions`; `us8` und `uxg-lite` dokumentieren ihre Controller-Evidenz fuer 10.3.58 und 10.4.57, ohne das Gateway-Profil hochzustufen. | S |
| B10 | P2 | Release-Hygiene wiederherstellen | Changelog und Installationsdokumentation entsprechen dem aktuellen Tag. Release-Artefakte enthalten ein dauerhaftes SBOM. Die Release-Checkliste deckt Signierung und Repository-Metadaten ab. | M |

## Umsetzungsreihenfolge

1. Die Sicherheitsblocker B01 und B02 wurden am 2026-07-10 abgeschlossen.
2. Laufzeitkorrektheit: B03 und B04 wurden am 2026-07-10 abgeschlossen.
3. Controller-Kompatibilitaet: B05, B06 und B07 wurden am 2026-07-10 abgeschlossen.
4. Qualitaets- und Release-Gates: B08 und B09 sind abgeschlossen; als Naechstes folgt B10.

## Alpha-Release-Gate

Keinen weiteren Alpha-Tag erstellen, bevor B01 bis B06 abgeschlossen sind und
diese Befehle auf dem Release-Kandidaten erfolgreich laufen:

```sh
make check
go test -race ./...
make vulncheck
make integration-docker
make package
```

## Beta-Einstiegskriterien

- Das Controller-Gate besteht gegen mindestens zwei gepinnte UniFi-Network-Versionen.
- Linux- und FreeBSD-Paket-Smoke-Tests einschliesslich Upgrade und Rollback bestehen.
- Kein Controller-Kommando verbleibt in einer wiederkehrenden Task-Schleife.
- Dynamische Interface-Aenderungen werden ohne Daemon-Neustart dargestellt.
