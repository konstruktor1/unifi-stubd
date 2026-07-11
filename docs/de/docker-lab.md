# Docker-Controller-Lab

Das Docker-Lab unter `lab/stub/` ist die projekt-eigene Integrationsumgebung
fuer den Go-Stub. Es verwendet drei laenger laufende Services:

- UniFi Network Application auf `https://127.0.0.1:8443/`
- MongoDB fuer den Controller
- Go-Inform-/Discovery-Proxy im internen Lab-Netz

Das Default-Controller-Image ist auf
`lscr.io/linuxserver/unifi-network-application:10.4.57-ls135` gepinnt.
`UNIFI_NETWORK_IMAGE` nur ueberschreiben, wenn bewusst eine andere
Controller-Version validiert wird. `UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION`
auf die `/status`-`server_version` dieser Version setzen, wenn der
Integrationstest sie erzwingen soll.

Das Integrations-Overlay `lab/stub/compose.tests.yaml` ergaenzt temporaere
Services `stub-bridge-observe`, `stub-port-map` und `stub-gateway-smoke`. Sie
werden aus dem aktuellen Repository-Checkout gebaut und vom Test-Harness wieder
entfernt.

Projekt-eigene Lab-Defaults liegen in
`lab/stub/configs/hosts/<hostname>/config.yaml`, mit einem Verzeichnis pro
gemeldetem Stub-Hostnamen, und werden read-only in die Stub-Container
gemountet. Das Test-Harness uebergibt Wegwerf-MACs/-IPs, Profile und Hostnamen
weiter als CLI-Overrides, damit diese Dateien stabil bleiben und keinen
Controller-State oder Secrets enthalten.

Leere Controller-Volumes werden automatisch ueber denselben lokalen Setup-API-
Ablauf wie in der Weboberflaeche initialisiert. Bereits konfigurierte Volumes
werden nach geprueftem Lab-Login wiederverwendet. Der Setup-Helfer gibt weder
Controller-Antworten noch generierte Device-Zugangsdaten, API-Tokens oder
Adoption-Keys aus.

## Controller-Gate

Aus dem Repository-Root starten:

```sh
make integration-docker
```

Das Target prueft:

- Compose-Konfiguration fuer Basis-Lab plus Test-Overlay.
- Runtime-Image-Build, inklusive `iproute2` fuer Bridge-/FDB-Observation.
- `bridge-observe`-Dry-Run-Payload aus einer container-lokalen Linux-Bridge.
- `management_lan.mode: preexisting-interface`-Dry-Run-Payload gegen die
  Container-`eth0`-Adresse fuer den neuen Switch-Management-LAN-Pfad.
- `port-map`-Dry-Run-Payload aus container-lokalen veth-Interfaces.
- Gateway-Dry-Run-Payload aus dem `uxg-lite`-Profil, inklusive `if_table`,
  `network_table`, read-only physischer `port_table`, `reported_networks`,
  `uplink_table` und `wan1` aus der gemeinsamen Port-View.
- Einen Inform-Request und ein weitergeleitetes Discovery-Paket pro Modus durch
  den Go-Proxy.
- Controller-API-Login gegen die Docker-UniFi-Network-Application.
- Controller-`/status`-Versionspruefung fuer das gepinnte Docker-Image.
- Pending-Adoption-Sichtbarkeit fuer Bridge-Observe- und Gateway-Smoke-Geraet.
- Controller-getriggerte Adoption ueber die Controller-API fuer Switch- und
  Gateway-artige Payloads.
- Eine mit dem Default-Key dekodierte `setparam`-Antwort mit Authkey,
  cfgversion und `use_aes_gcm=true`, nur als bereinigte Presence-Flags erfasst.
- Persistierten lokalen Stub-Adoption-State mit `STATE=connected`, Authkey,
  cfgversion und aktiviertem AES-GCM, ohne den Authkey auszugeben.
- Mindestens zwei Post-Adoption-AES-GCM-Informs pro adoptiertem Switch- und
  Gateway-Testgeraet.
- Drei aufeinanderfolgende Controller-Samples mit `adopted=true`, `state=1`
  und abgeschlossener Adoption.
- Kein Gateway-Kommando, Ergebnis, Duplikat oder ausstehender State vor und
  nach dem 75-Sekunden-Stabilitaetsfenster des aktuellen 10.4.57-Pins.

Die nur an Loopback gebundenen Default-Lab-Zugangsdaten sind `admin` / `admin`;
sie werden auch beim Bootstrap leerer Volumes verwendet. Fuer einen
abweichenden lokalen Lab-Controller koennen sie ueberschrieben werden:

```sh
UNIFI_STUB_LAB_ADMIN_USER=admin \
UNIFI_STUB_LAB_ADMIN_PASSWORD=... \
make integration-docker
```

## Docker-Gate Von Dev Nach Main

Das Docker-Lab ist der standardisierte Controller-Kompatibilitaets-Gate vor der
Promotion von `dev` nach `main`, sobald die gesammelten `dev`-Aenderungen einen
dieser Bereiche beruehren:

- Inform-Framing, Verschluesselung, Kompression oder Adoption-Responses;
- controller-sichtbare Payload-Form, Profildaten, Port-Rendering,
  Gateway-Tabellen oder WAN-Health;
- Observation-Modi, die gerenderten Controller-State beeinflussen;
- Docker-Lab-Fixtures, Controller-Image-Pinning oder Adoption-Testhelfer.

Der Gate haengt am Commit, nicht am Branch-Namen. Der von `dev` nach `main`
promotete Commit muss derselbe Commit sein, der den Docker-Gate bestanden hat,
oder ein Nachfolger, der nur nicht fachliche Dokumentation oder Release-
Metadaten enthaelt.

Standardlauf:

```sh
git switch dev
git pull --ff-only origin dev
make check
make integration-docker
git rev-parse HEAD
```

Pass-Kriterien:

- `make check` beendet mit `0`.
- `make integration-docker` beendet mit `0` und gibt `docker integration: ok`
  aus.
- Der Log zeigt die Versionspruefung der gepinnten UniFi Network Application.
- Der Log zeigt die erzeugten Bridge-, Port-Map- und Gateway-Wegwerf-
  Identitaeten.
- Switch- und Gateway-Gates erfassen jeweils Discovery, bereinigtes `setparam`,
  zwei GCM-Informs, vollstaendigen persistierten Adoption-State und stabile
  Connected-Samples.
- Der Gateway-Gate erfuellt die explizite Controller-Kommando-Erwartung vor und
  nach einem Stabilitaetsfenster von 75 Sekunden. Der aktuelle 10.4.57-Pin
  erwartet kein Kommando; Kompatibilitaetslaeufe koennen fuer Controller, die
  `wg-genkeypair` senden, mit
  `UNIFI_STUB_DOCKER_GATEWAY_COMMAND_EXPECTATION=required` den kompletten
  Antwortpfad erzwingen.
- Der abschliessende Gateway-Status meldet den positiven Controller-Hinweis,
  begrenzt auf die lokalen Werte, als effektiven und naechsten Takt, mit null
  aufeinanderfolgenden Fehlern und ohne aktiven Backoff.

Als Promotion-Evidenz wird festgehalten:

- Commit-SHA;
- ausgefuehrter Befehl;
- Controller-Image und erwartete Controller-Version;
- Start- und Endzeit;
- finaler Exit-Code;
- Link auf einen GitHub-Actions-Lauf oder einen aufbewahrten lokalen
  Terminal-Log.

`dev` wird nicht nach `main` promotet, solange der Docker-Gate fehlschlaegt,
bei controller-sichtbaren Aenderungen uebersprungen wurde oder gegen einen
anderen Commit lief. Fuer Fehleranalyse darf einmal mit
`UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` neu gestartet werden; danach muessen
temporaeres Geraet und State-Volume vor dem naechsten Standardlauf entfernt
werden.

Automatisierungsziel:

1. Manueller Gate bleibt die Basis, bis der Docker-Lab-Runner stabil ist.
2. Ein separater `Docker Integration`-GitHub-Actions-Workflow bekommt
   `workflow_dispatch` und `pull_request` fuer `dev` nach `main`.
3. Dieser Workflow fuehrt nach gruenem normalem `CI / check` nur
   `make integration-docker` aus.
4. Eine einzelne Concurrency-Gruppe fuer das Docker-Controller-Lab verhindert,
   dass zwei Adoption-Tests gleichzeitig Controller-State teilen.
5. Sobald stabil, wird `Docker Integration` als Required Status Check fuer
   `dev`-nach-`main`-Pull-Requests mit controller-sichtbaren Aenderungen
   verwendet.

Der normale `CI`-Workflow bleibt der schnelle Gate. Der Docker-Gate ist der
Controller-Kompatibilitaets-Gate, und der `main`-Package-Job bleibt der
Package-Install-Smoke-Gate.

## Cleanup-Semantik

Das Skript erzeugt pro Lauf Wegwerf-MACs/IPs, stoppt und entfernt temporaere
Stub-Container und Volumes und fordert den Controller auf, adoptierten State
fuer die Test-MACs zu loeschen. Controller-Volumes werden nicht zurueckgesetzt.
Die Controller-Loeschanfrage ist Best-Effort; verlaessliche Isolation kommt von
frischen Wegwerf-MACs pro Lauf.

Am Anfang eines Standardlaufs entfernt der Helfer ausserdem adoptierte Geraete
mit dem reservierten Docker-Fixture-MAC-Praefix `02:15:6d`. Dadurch erholt sich
der Gate selbst von einem vorherigen Lauf mit
`UNIFI_STUB_DOCKER_KEEP_RESOURCES=1`, ohne Geraete ausserhalb des
projekt-eigenen Fixture-Praefix anzufassen.

Liefert UniFi beim Loeschen eines solchen reservierten Fixtures
`api.err.DeviceBusy`, startet der Gate nur den wegwerfbaren Controller-Service
neu und wiederholt danach die normale Loeschung ueber die Controller-API. Die
Controller- und MongoDB-Volumes werden nicht zurueckgesetzt.

UniFi Network kann nicht adoptierte Pending-Zeilen bis zum Discovery-TTL im
Prozessspeicher behalten. Im beobachteten Docker-Lab sind diese Zeilen nicht in
MongoDB persistiert. Frische Wegwerf-MACs vermeiden Kollisionen zwischen
wiederholten Laeufen. Im Lab-Helfer bedeutet `wait-clean` "nicht adoptiert";
`wait-absent` nur verwenden, wenn ein Test wirklich das Verschwinden der Zeile
braucht.

`UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` nur setzen, wenn ein adoptiertes Testgeraet
oder das Stub-State-Volume nach einem fehlschlagenden Lauf absichtlich
inspiziert werden soll.

## Grenzen

Dieses Lab belegt container-lokale Linux-Bridge-/FDB-Observation, sysfs-Zaehler,
explizites Port-Mapping, Gateway-Tabellen-Rendering, Discovery-Weiterleitung,
Inform-Framing, Controller-Adoption, lokale Adoption-State-Persistenz und
AES-GCM. Es belegt nicht Proxmox-Host-Bridge-Verhalten,
FreeBSD-Runtime-Verhalten, LLDP-Import oder Event-Subscriptions.

Es belegt auch keine physische Topologie-Richtung. Container-Tests nutzen
Wegwerf-Identitaeten, decken also nicht den Fall ab, dass ein echter
Upstream-UniFi-Switch dieselbe physische Host-MAC bereits meldet. Reale
Proxmox- oder Bridge-Deployments sollten `uplink_neighbor`, `uplink_port` und
die Wahl zwischen synthetischer und physischer MAC gegen den Zielcontroller
validieren, bevor das Ergebnis als repraesentativ gilt.
