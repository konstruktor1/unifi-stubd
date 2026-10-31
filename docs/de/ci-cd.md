# Build, Pruefungen und Auslieferung

Das Root-`Makefile` ist die kanonische Schnittstelle fuer Build und Qualitaet.
Diese Seite ordnet seine Einstiege, Runtime-Voraussetzungen und Provider-Handoffs
zu.

Vor Host-Orchestrierung oder direkten Provider-Aufrufen den einmal definierten
Kontext ueber den vorhandenen nativen Projekt-/Sitzungsowner laden. Make prueft
und erbt ihn fuer diese Ziele:

```sh
: "${PROVIDER_CONSUMER_ROOT:?Nativen Projektkontext laden}"
: "${PROVIDER_CONSUMER_CONTEXT:?Nativen Consumer-Kontext laden}"
BUILD_RUNNER_ENTRY="$(jq -er '.command | select(endswith("/bin/provider")) | sub("/provider$"; "/build-runner")' .consumer-data/connections/build-runner/management.json)" || exit "$?"
export BUILD_RUNNER_ENTRY
```

Der CI-Owner liefert diesen Kontext einmal und vererbt ihn unveraendert. Aktuelle Quellen
stammen aus nativen Bindungen; installierte Snapshots, provider.json, Commits oder
Tags waehlen keinen Stand. CTs CLI ist sein aktuelles `bin/ct`.
Sein Service-MCP verwendet den geprueften externen
Management-Einstieg aus
`.consumer-data/connections/ct/management.json`.

Go-Payloadziele laufen in Build Runners `/workspace` mit `GOCACHE` und der
bestehenden Workspace-Pruefung. Sie benoetigen keine in den Container
uebertragene Host-Provider-Verbindung oder empfangende Kontextdefinition. Der
native Setup-Owner muss `tests/` und das temporaere `tests/.work/` fuer den
gemeinsamen Verifikationsspeicher als private Verzeichnisse mit Modus 0700
vorbereiten. Dies ist eine lokale Dateisystemvoraussetzung; Git speichert keine
Verzeichnismodi.

## Consumer-Einstiege

Alle von Agenten ausgefuehrten Go-Arbeiten verwenden den konfigurierten Build
Runner gemaess [`AGENTS.md`](../../AGENTS.md). Host-Go oder ein Go-basiertes
Make-Target ist kein Fallback.

| Ergebnis | Einstieg |
| --- | --- |
| Vollstaendiger Qualitaetsgate aus dem nativ konfigurierten Consumer auf dem Host | `make verify` |
| Aktuellen nativen BR-Integrationskontext lesen und pruefen | `make provider-check` |
| CT in seiner gespeicherten Pruefumgebung ausfuehren | `make quality-ct` |
| Lokales Setup inspizieren und nicht ausgefuehrte Produktpruefungen ausweisen | `make doctor` |
| Runner-lokaler Produkt-Subgate (kein vollstaendiger Pre-Merge-Gate) | `make check` |
| Controller-sichtbarer Gate (gebundener Tier-3-Lauf oder Stable-/RC-Matrix) | `make integration-nonunifi-metrics` |
| Linux-Pakete | `make package` |
| FreeBSD-Paketsatz (nur Kompatibilitaet; als Evidenz blockiert) | `make package-freebsd-pkg-repos` |
| Statische Paketquellen | `make package-repos` |

`make verify` fuehrt zuerst `make provider-check` aus, installiert danach die
festgelegten JavaScript-Projektabhaengigkeiten und fuehrt ESLint und Prettier im
deklarierten Dienst `node24` aus.
Im Dienst `go` laufen mit `make check vulncheck` gofmt, `go vet`, golangci-lint,
Repository-Policy, Konfigurations-/Profilvalidierung, Tests und der
Schwachstellenscan. Danach liest `make quality-ct` ueber `scripts/check-ct.py`
CTs aktuellen Vertrag im Management-MCP und beachtet die gespeicherte Auswahl
`build-runner`. Die entdeckte BR-Funktion `runner` fuehrt CTs aktuelle
oeffentliche CLI in der kurzlebigen Instanz `ct-quality` aus; dort laufen alle
ausgewaehlten Go-, JavaScript-, Bash- und Python-Pruefungen.
Der offizielle SDK-Client verwendet die enthaltene Projektumgebung unter
`.consumer-data/dependencies/unifi-stubd/mcp/`, deklariert durch
`requirements/unifi-stubd/mcp.txt`, und die aktuelle native Verbindung.
Die CT-Prueftoolchain stammt aus dem gepflegten Qualitaetsimage,
einschliesslich Go, Node, Python, Ruff, ShellCheck und shfmt. Host-Go oder
Host-Node sind nicht erforderlich.
Fehlende Tools oder Provider-Runtimes lassen den Gate fehlschlagen. Paket- und
Controller-Abnahme bleiben je nach Aenderung zusaetzlich erforderlich.

`build-runner/instances/ct-quality.yaml` bindet die aktuelle CT-Quelle ueber
`CT_SOURCE` schreibgeschuetzt unter `/opt/codetool` ein. Die native BR-Verbindung
leitet `BUILD_RUNNER_HOST_PATH_CT_SOURCE` aus der gespeicherten CT-Quellenbindung
ab; bei der Einrichtung eines weiteren Runners diesen generierten Eingang
beibehalten. Die gepflegte Imagequelle ist
`build-runner/project-setup/ct-quality/Dockerfile`. `make quality-ct-image`
verwendet nativen Docker-Build, weil BR keine Build-Operation fuer Consumer-
Dockerfiles anbietet; die Checks laufen weiter ueber den oeffentlichen BR-
Instanzrunner. Das Manifest validiert physischen `/workspace`-Root/cwd,
schreibgeschuetzte CT-Quelle und enthaltene HOME-, Abhaengigkeits-, Cache- und
Temporaerpfade. Fehlende Bindungen, Images oder eine fehlgeschlagene Validierung
stoppen den Check ohne nativen Host-Fallback.

`golangci-lint.mod` und `golangci-lint.sum` trennen die Linter-Abhaengigkeiten
vom Produktmodul. Beide verlangen Go 1.27 oder neuer. Das Makefile ruft
dieses Toolmodul innerhalb der deklarierten Build-Runner-Go-Runtime auf.
Die Go-Befehle verwenden den eigenen Paketscope aus `GO_PACKAGES`; installierte
Abhaengigkeiten unter `node_modules` liegen ausserhalb dieses Scopes.

Die aktuelle BR-Runtimeauswahl verlangt fuer jeden Docker-Aufruf frische
Beobachtungen der CPU des urspruenglichen Hosts und der tatsaechlich wirksamen
Image-Plattform. Die Architektur des Docker-Daemons ist eine getrennte
Beobachtung. Diese Identitaeten und die Validierung des deklarierten
Runtime-Owners beibehalten; keinen generischen Runner einsetzen oder den Host
aus Daemon oder Image ableiten.

## Offene Provider-Handoffs

Fruehere Gate- und Compilerbeobachtungen stehen im
[Verifikationsnachweis](../engineering/backlog.md). Die laufende Integration
aktueller Dateien, beobachtete Checks und offene Owneraktionen stehen im
[Engineering-Backlog](../engineering/backlog.md).
Die erforderlichen Sprachwerkzeuge laufen jeweils in ihrer
deklarierten Umgebung. Nativer MCP-Transport, Abhaengigkeiten und Operationsdaten
bleiben innerhalb des Consumer-Roots.

`make provider-check` vergleicht die gespeicherte BR-Bindung mit dem frischen
oeffentlichen Verbindungsexport, liest alle aktuell erforderlichen MCP-Ressourcen
unter einem bestaetigten aktiven Katalog-/Annotations-/Listenabonnement, ruft
Annotations ab und fragt die optionale Persistenzanleitung fuer
`runner` / `operator-targets` an. Die Anfrage legt weder einen Zielspeicher noch
eine gespeicherte Auswahl an. Der bestehende Owner `scripts/check-ct.py` verwendet
`scripts/provider_context.py` fuer gemeinsame native Bindungs- und
SDK-Kontextbehandlung von BR und CT. Sein CT-Gate gleicht Provider-ID,
Consumer-Root und Konfiguration mit dem aktuellen Vertrag ab.
Geschlossene logische Aufrufe werden auf die entdeckten
direkten name/arguments-Schemas projiziert.
Fehlgeschlagene Checks werden an Make weitergegeben. Bei Abschluss oder Abbruch
schliesst er die Verbindung und wartet auf das bestaetigte Provider-Ende.
CI verwendet denselben Einstieg `make verify`. Das
[Provider-Setup](../../.github/actions/provider-setup/action.yml) bereitet aktuelle
unabhaengige BR-/CT-Quellen, den nativen Client und die Management-Verbindungen
fuer das exakte Checkout vor; das ausgewaehlte CT-Image entsteht ueber den
bestehenden Make-Owner. Lokaler Erfolg belegt weder diese Voraussetzungen noch
einen erfolgreichen Lauf auf einem GitHub-hosted Worker.

Derselbe Checkowner liefert `annotations`, `doctor` und fuenf exakte eigene
Skillressourcen ueber `.consumer-data/connections/unifi-stubd/annotations.json`.
Der gespeicherte native Einstieg `--mcp-annotations` bleibt gueltig. Eigene
Funktions-, Annotations-, Update-, Doctor- und Delivery-Skills liegen unter
`.agents/skills/unifi-stubd*/`; ihre Ressourcen und Listenmeldungen unterstuetzen
SDK-Abonnements und Neuverbindung nach echten Byte-/Modusaenderungen. Diese
Terminal-Surface deklariert keinen zusaetzlichen PP-Katalog oder Provider-Vertrag.
AGENTS.md und die [lokale Integrationsanleitung](../engineering/provider-integration.md)
regeln Umsetzung, Speicher und verpflichtende Weitergabe.

Fuer `provider-update`, explizite BR- oder CT-Updates, anwendbare Annotations,
abonnierte Ressourcen-Aenderungshinweise und bestaetigte Aenderungen der aktuellen
Quellen den eigenen Skill `unifi-stubd-update` auswaehlen. Jedes Signal waehlt
denselben vollstaendigen autorisierten lokalen Ablauf fuer Uebernahme, Pruefung,
eigene Hinweise und Weitergabepflichten gemaess lokaler Anleitung. Das ist eine
Skillanweisung ohne neuen Updater-Befehl oder automatische Doctor-, Fremdprojekt-
oder Veroeffentlichungsaktion.

Die sieben aktuellen CT-Skills bleiben unter `.agents/skills/ct*/` exakte fremde
Ressourcen. Bei autorisierter CT-Integration fragt der oeffentliche Einstieg
`init` nur fehlende Erstauswahlen und speichert vollstaendig bestaetigte Antworten
im eigenen `ct.toml`; gespeicherte Auswahlen werden nur auf Wunsch neu geoeffnet.
Die aktuellen Entscheidungen behalten Formatter-/Linter-Konventionen und die
deaktivierte optionale i18n-Policy bei, waehlen `build-runner` fuer CT-Pruefungen
und aktivieren einen Quellenreview alle sieben Tage. Die Pruefumgebung ist von
Python-/MCP-Runtimeauswahl getrennt. Ihr Setup-Owner muss `owner_work` und
`first_check` mit echten Werkzeug-, Formatter- und Namespace-/Containment-Belegen
abschliessen. Eine gespeicherte BR-Auswahl ist kein bestandener Check und erlaubt
keinen nativen Fallback; beobachtete Setup-Ergebnisse stehen im Engineering-Backlog.

Der Quellenreview wird bei CT-Aufrufen als faellige Agentenarbeit gemeldet; es
gibt keinen Hintergrundjob. CLI-Stderr `ct-source-review:` und Diagnosen auch
erfolgreicher MCP-Aufrufe an den aufrufenden Agenten weiterreichen. Bei Faelligkeit
den aktuellen `source-review`-Auftrag verwenden, echte Primaerquellen lesen und
autorisierte Korrekturen/Checks abschliessen, bevor sein deklariertes Ergebnis
gespeichert und der Faelligkeitszustand frisch nachgelesen wird. Die
[lokale Integrationsanleitung](../engineering/provider-integration.md) besitzt
die konkrete Persistenz- und Weitergabemethode. Eine fehlgeschlagene native
Kontextinitialisierung muss vor Exportauswertung oder abhaengiger Arbeit stoppen.

Doctor nur fuer angeforderte lokale Diagnose auswaehlen. `make doctor`
inspiziert. Sein Diagnosebefehl endet bei unvollstaendiger Abdeckung bewusst mit
Status 1, daher meldet Make ein fehlgeschlagenes Ziel: Die erforderliche
Produkt-Runtimeabnahme wurde nicht ausgefuehrt. Das konfigurierte
eigene Python unterstuetzt auch `scripts/check-ct.py --doctor`, optional mit
`--mode inspect|verify`, `--scope setup`, `--consumer-role` und `--phase` gemaess
Hilfe/Schema. Explizite Verifikation fuehrt die bestehende Repository-Policy,
BR-Identitaetspruefung und CT-Vertragsaufloesung aus. Sie startet weder Workload
noch Controller und weist semantische YAML-Validierung, Go-/JavaScript-Checks
und Controller-Abnahme weiterhin als nicht ausgefuehrt aus. Dieses
Diagnoseergebnis ist vom Qualitaetsgate `make verify` getrennt.

`make integration-nonunifi-metrics` verwendet das Consumer-eigene Manifest
`build-runner/instances/controller-gate.yaml` ueber Build Runner. Seine
Tier-2-Basis ist ein unveraenderliches CircleCI-Image `cimg/base` mit Bash,
curl, jq, Make, Docker CLI, Compose und Buildx. Die Validierung prueft diese
Werkzeuge und den Docker-API-Zugriff vor dem Matrixlauf; pro Job werden keine
Tools installiert. Go-Builds gehoeren in die Produkt- und Helfer-Dockerfiles;
dieses Orchestrierungsimage benoetigt kein Go.

Fuer den zugewiesenen Tier-3-Kandidatenlauf von Core setzt der Qualifier am
gleichen Make-Einstieg
`UNIFI_STUB_TIER3_GATE_MANIFEST=dist/tier3-controller-input/<run-id>.env`.
Make reicht die Bindung weiter; der Matrix-Einstieg delegiert vor jeder
Standalone-ZIP-Aufloesung an den
[Tier-3-Controller-Vertrag](../../lab/stub/TIER3-CONTROLLER.md). Ohne diese
Eingabe bleibt die bisherige Stable-/RC-Matrix unveraendert.
`make qualify-candidate` und `make qualify-candidate-matrix` fuehren den
Produktablauf im bestehenden nativen Core-Test-Runner-Kontext aus. Labs liefert
Controller-Vorbereitung und frische Herkunftsnachweise.
`make integration-tier3-candidate-controller-cleanup` bleibt das explizite
Recovery-Ziel nach einem abgebrochenen Gate. Ein gebundenes Stable-Ergebnis trifft keine
Aussage ueber RC.

Die lokale Netzwerkberechtigung des Browsers gilt nur fuer seine Controller-
Origin. Controller-CSRF-Header bleiben auf dieser Origin. Die oeffentliche
Geraetebild-Datenbank unter `https://static.ui.com/fingerprint/ui/public.json`
wird separat ohne Controller-Zugangsdaten abgerufen; andere externe
Browseranfragen bleiben blockiert. Antwortgrenzen und bereinigte Fehlerdiagnosen
sind im Vertrag beschrieben.
Die UniFi-OS-Abnahme liest Linkgeschwindigkeit sowie RX/TX-Pakete und -Bytes
aus derselben eindeutigen sichtbaren Zeile des nativen Port Managers mit
Zielgeraetefilter.

Das Manifest gewaehrt ausdruecklich Docker-API-Zugriff und die fuer den Socket
erforderliche Root-Identitaet im Orchestrierungscontainer. Nur mit einem
isolierten Lab-Daemon verwenden. Der Make-Einstieg uebergibt den physischen
Lab-Pfad fuer Bind-Mounts; dieser muss fuer den ausgewaehlten Docker-Endpunkt
sichtbar sein. Ein Remote-Pfad wird nicht geraten.

Ohne gebundenes Tier-3-Manifest loest der erhaltene Matrix-Payload das neueste
veroeffentlichte Stable und den neuesten Release Candidate
zu Laufbeginn auf und verwendet in beiden wegwerfbaren Spuren denselben
Kandidaten. Details und Bedingungen stehen im
[Docker-Controller-Lab](docker-lab.md). Die vollstaendige Produktabnahme
verlangt zusaetzlich die fokussierten Host-Safety- und No-Fabrication-Belege
aus den [Requirements](requirements.md).
Dies beschreibt erhaltenes Fixture-Verhalten. Controllergestuetzte Abnahme muss
die zugewiesene Tier-3-Maschine verwenden; eine fehlende Bindung laesst Ownerarbeit
offen und erlaubt weder ein Core-Fixture noch einen GitHub-Worker als Ersatz.

## Verantwortung der Automatisierung

- [CI](../../.github/workflows/ci.yml) besitzt `make verify`, Coverage und den
  Linux-Paketbau auf `main` auf GitHub-hosted Ubuntu-Workern. Build Runner
  liefert die kurzlebigen Toolchain-Container; der Install-Smoke verwendet
  ebenfalls einen wegwerfbaren GitHub-hosted Worker. Checks laufen bei Pushes
  in Repository-Branches und manuellem Dispatch. Fork-Pull-Requests erhalten
  einen blockierenden Hinweis ohne Codeausfuehrung; nach dem Review uebernimmt
  ein Maintainer den Beitrag in einen Repository-Branch.
- [Controller Integration](../../.github/workflows/tier2-docker.yml) besitzt den
  manuell gestarteten Controller-Gate. GitHub betreibt den Orchestrierungsjob und
  bereitet BR ueber dasselbe Provider-Setup vor. Ein gehosteter Job belegt weder
  Zugriff auf eine bindende Tier-3-Testmaschine noch Controller-Bereitschaft oder
  Host-Safety-Abnahme. Controllergestuetzte Arbeit verlangt weiterhin die
  zugewiesene Maschine, das aktuelle Companion-Inventar und den deklarierten
  Lab-Zugriffspfad.
- [Package Repositories](../../.github/workflows/package-pages.yml) verwendet
  GitHub-hosted Ubuntu-Worker fuer Linux-Releaseartefakte und die Paketseite.
- Nur [FreeBSD Packaging](../../.github/workflows/freebsd-packaging.yml) verwendet
  den vorhandenen Self-hosted Runner `docker-dmz-unifi-stubd` fuer seinen
  manuellen Kompatibilitaetsbuild und temporaere Laufartefakte. Diese Rolle
  behaelt ihre Docker-Daemon-Beschraenkung und darf nicht publizieren.

### Providerquellen auf GitHub-hosted Workern

Das [Provider-Setup](../../.github/actions/provider-setup/action.yml) bezieht BR
und CT unabhaengig aus ihren konfigurierten privaten Forgejo-Quellen in die
bestehende Consumer-Rolle `.consumer-data/dependencies/`. Es liest den aktuellen
Default-Branch und waehlt keinen Provider-Release, Tag oder historischen
Ausfuehrungsbaum aus. Der Quellentransport ersetzt nicht die aktuellen
Artefaktpruefungen der oeffentlichen Launcher. Die Quelle wird weder aus dem cwd
des Checkouts noch aus einem Mac-Pfad geraten.

Privater Lesezugriff verwendet das GitHub-Actions-Secret `PROVIDER_SOURCE_TOKEN`
und die Repository-Variablen `PROVIDER_SOURCE_USERNAME`, `BUILD_RUNNER_SOURCE_URL` und
`CT_SOURCE_URL`. Die Quellorte bleiben in den nativen GitHub-Einstellungen. Der Zugriff muss beide
Provider-Repositories nur lesend erreichen; Zugangsdaten gehoeren nicht in URLs,
gespeicherte Git-Remotes, Verbindungen, Logs oder eingecheckte Dateien. Die
konfigurierten Quell-URLs sind native CI-Eingaenge. Das GitHub-Repository-Token
gewaehrt fuer sich keinen Forgejo-Zugriff.

Der Job definiert das exakte physische Checkout einmal ueber
`PROVIDER_CONSUMER_ROOT` und seinen nativen `PROVIDER_CONSUMER_CONTEXT`;
Kinder erben beide unveraendert. Der Job stellt kompatibles natives Python und
den Projekt-MCP-Client bereit. Danach erzeugt der bestehende Preflight jede
gespeicherte Management-Verbindung mit dem oeffentlichen
`bin/provider mcp --connection` der jeweiligen Quelle, prueft `help --json`
und fuehrt `check --json` aus. BR behaelt den abgeleiteten aktuellen
CT-Quelleneingang fuer die ausgewaehlte CT-Umgebung dieses Consumers.
Go, Node und die ausgewaehlten CT-Werkzeuge laufen in BR-Containern. BR und CT
bleiben unabhaengige Funktionen.

Auf wegwerfbaren GitHub-Ubuntu-Workern verwenden Provider-Setup und
Host-Orchestrierung dieselbe Root-UID wie der BR-Controller. Das Setup validiert
zuerst den nativen physischen Root und uebertraegt dann die Eigentuemerschaft
dieses Job-Checkouts, bevor privater Provider-Zustand entsteht. Die privaten
`0700`-Pruefungen bleiben erhalten. Dieses Setup lehnt selbst gehostete Worker ab
und gilt nicht fuer den FreeBSD-Runner.

Beide Repositories lehnen anonyme Zugriffe ab. Das ausgewaehlte GitHub-Secret und
die Benutzer-Variable sind jetzt eingerichtet. Der Lesetoken ist auf genau die
beiden Provider-Repositories beschraenkt. Ein frischer wegwerfbarer Consumer hat
beide Quellen geklont, den SDK-Client installiert und ueber den echten Preflight
native Bindungen erzeugt. Der Quellen-Owner hat die geprueften Korrekturen aus
CT-004/CT-006 im aktuellen Standardbranch veroeffentlicht. Repository-Lieferung
und ein beobachteter GitHub-Lauf schliessen CI-001 ab; lokale Tests und ein
vorbereiteter Workflow allein belegen keine erfolgreiche Remote-Ausfuehrung.
Der native Client vergleicht erhaltene CT-Anweisungen mit den aktuellen
Quellressourcen byte- und modusgenau; eine abweichende Lieferung stoppt mit
konkreter Owneraktion. `make verify` fuehrt ausserdem `make workflow-lint` im
Go-Dienst zur Workflow- und Action-Metadatenvalidierung aus.

### Historische Runner-Beobachtung

Die reine GitHub-Lesung vom 2026-10-28 fand keine Provider-Launchervariablen und
einen Online-FreeBSD-Runner ohne `docker-engine`. Der fruehere Vorschlag, fuer
allgemeine CI einen Docker-faehigen Self-hosted Worker zu verlangen, war zu weit
gefasst und ist durch die Rollenaufteilung oben ersetzt. Der FreeBSD-Runner wird
fuer diese Jobs weder umgelabelt noch mit Docker-Zugriff versehen. Fruehere gruene
Workflow-Laeufe verifizieren den neuen Provider-Bootstrap nicht.

Das Administrationsinventar zu `unifi-stubd-gh-runner` auf dem Gast `docker`
beschreibt diese vorhandene FreeBSD-Rolle; sie ist keine Voraussetzung fuer
allgemeine Checks auf GitHub-hosted Workern. Der direkte FreeBSD-Builder bleibt
ein External-Machine-Kompatibilitaetspfad ohne Build-Runner-Evidenz oder
Publikationspfad.

Releaseentscheidungen und Betreiberschritte stehen im
[Entwicklungsworkflow](development.md) und in der
[Release-Checkliste](release.md).
