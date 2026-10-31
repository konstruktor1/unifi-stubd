# Docker-Controller-Lab

Das Docker-Lab unter `lab/stub/` beantwortet eine Produktfrage: Kann das
aktuelle `unifi-stubd`-Binary eine Nicht-UniFi-Linux-Bridge als `US8`
darstellen, adoptiert werden, verbunden bleiben und Live-Metriken in UniFi
Network anzeigen?

Der Gate verwendet genau ein Produkt-Fixture: `stub-bridge-observe`. Die
uebrigen Services sind nur Bereitstellung: UniFi Network, MongoDB, ein
Controller-API-Helfer und ein Browser-Helfer. Payload-Dumps, lokale
Statusschemata, Gateway-Kommandos, synthetische Port-Maps und
Management-LAN-Varianten sind keine eigenen Controller-Gates.

## Controller-Gate

Der Einstieg fuehrt die aeussere Matrix ueber die deklarierte Build-Runner-
Tier-2-Instanz aus; siehe [CI/CD](ci-cd.md). Ihre Validierung verlangt
Docker-API-Zugriff auf den ausgewaehlten isolierten Lab-Daemon:

```sh
make integration-nonunifi-metrics
```

Die Matrix fragt beide Versionen zu Beginn jedes Laufs beim
Ubiquiti-Release-Dienst ab; keiner der beiden Kanaele wird im Repository
festgeschrieben. Das Produkt wird einmal gebaut und in beiden wegwerfbaren
Controller-Spuren bytegleich geprueft.

Der Gate ist nur erfolgreich, wenn beide Spuren Folgendes belegen:

- der tatsaechliche Controller meldet die aufgeloeste Version;
- der US8 ist vor der Adoption sichtbar und akzeptiert die Controller-Adoption;
- waehrend begrenzter bidirektionaler Bridge-Traffic laeuft, zeigt die
  Geraeteansicht den beobachteten Speed sowie positive `Down`- und
  `Up`-Trafficwerte.

Das sind Beobachtungen an der Produktgrenze. Der Gate bildet die eigenen
JSON-Strukturen des Produkts bewusst nicht in Assertion-Helfern nach.
Die vollstaendige Produktabnahme verlangt zusaetzlich die getrennten
Host-Safety- und No-Fabrication-Belege aus den [Requirements](requirements.md).

## Evidence und Cleanup

Die Stable-/RC-Matrix behaelt nicht committete Laufartefakte unter
`dist/controller-channel-matrix/`:

- aufgeloeste Stable-/RC-Versionen und Releasemetadaten;
- Product-Gate-Logs fuer beide Spuren.

Jede Spur startet mit leeren Controller- und Datenbank-Volumes. Bei Erfolg und
Fehler entfernt der Gate seine Container, Netzwerke, benannten Volumes,
Orphans und eigenen Image-Aliase. Es gibt keinen Keep-Running-Bypass und keine
festen Containernamen.

## Grenze

Core besitzt die Produktqualifikation auf der zugewiesenen Linux-ARM64-Test-
Runner-Maschine ueber `make qualify-candidate` und `make qualify-candidate-matrix`.
Der [Qualifikations- und Controller-Vertrag](../../lab/stub/TIER3-CONTROLLER.md)
definiert den erforderlichen nativen Kontext, den expliziten Datenpfad,
Kandidaten-/SSH-Eingaben, Release-Bindung, Handshake und Cleanup. Core verwendet
die oeffentliche Test-Runner-Lifecycle-Schnittstelle ueber seine bestehende native
Build-Runner-Bindung und fuehrt das begrenzte Gast-Fixture samt strengem SSH-
Inform-Tunnel aus. Labs besitzt Controller-Vorbereitung, frische Provenienz je
Spur und die verbindlichen Maschinendefinitionen. Die Pruefung belegt den exakten
Gast-Executable-Hash und das echte Controller-Verhalten; ein Stable-only-Ergebnis
laesst RC- und gemeinsame Abnahme offen.

Fuer die separate UniFi-OS-Controllerintegration akzeptiert der API-Helfer
`--controller-mode unifi-os`, `--controller-url` und `--auth-file`. Der Ursprung
muss aus dem Helfer erreichbar sein und darf weder Zugangsdaten noch Pfad,
Query oder Fragment enthalten. Die private regulaere JSON-Datei enthaelt nur
`username` und `password`; sie bleibt mit Modus `0600` ausserhalb von Git.
Dieser Modus verwendet UniFi-OS-Login, Sitzungscookie und rotierende
CSRF-Header; Network-Anfragen laufen anschliessend ueber `/proxy/network`.
Die Controllereinrichtung gehoert dem Companion; der Helfer verweigert seine
Aktion `bootstrap` in diesem Modus.

Der Make-Handshake unterstuetzt Schema 1 fuer den klassischen wegwerfbaren
Controller und Schema 2 fuer einen Labs-eigenen UniFi-OS-Controller mit
Core-eigenen API-/Browser-Helfern. Das Ergebnis vom 26.10.2026 fuer Stable 11.0.97 auf der
verbindlichen Testmaschine einschliesslich sichtbarer Portmetriken und Cleanup
steht in [STUB-02](backlog.md). RC bleibt unqualifiziert; die Schema-Unterstuetzung
allein belegt keine Abnahme einer anderen Controllerinstanz oder eines anderen
Kanals.

Das Lab belegt container-lokale Linux-Bridge-/FDB-Beobachtung und das daraus
entstehende controller-sichtbare Produktverhalten. Es belegt keine physische
Topologie, Proxmox-Host-Bridge, FreeBSD-/OPNsense-Runtime oder einen deployten
Controller. Diese bleiben separat besessene maschinengebundene
Kompatibilitaetstests.
