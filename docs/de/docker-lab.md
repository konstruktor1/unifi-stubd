# Docker-Controller-Lab

Das Docker-Lab unter `lab/stub/` ist die projekt-eigene Integrationsumgebung
fuer den Go-Stub. Es verwendet drei laenger laufende Services:

- UniFi Network Application auf `https://127.0.0.1:8443/`
- MongoDB fuer den Controller
- Inform-MITM im internen Lab-Netz

Das Default-Controller-Image ist auf
`lscr.io/linuxserver/unifi-network-application:10.3.58-ls129` gepinnt.
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

## Smoke-Test

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
  `network_table`, `config_port_table`, `ethernet_overrides` und
  `reported_networks` aus der gemeinsamen Port-View.
- Einen Inform-Request pro Modus durch den MITM.
- Controller-API-Login gegen die Docker-UniFi-Network-Application.
- Controller-`/status`-Versionspruefung fuer das gepinnte Docker-Image.
- Pending-Adoption-Sichtbarkeit fuer Bridge-Observe- und Gateway-Smoke-Geraet.
- Controller-getriggerte Adoption ueber die Controller-API fuer Switch- und
  Gateway-artige Payloads.
- Persistierten lokalen Stub-Adoption-State mit `STATE=connected` und gesetztem
  Authkey, ohne den Authkey auszugeben.
- Mindestens einen Post-Adoption-Inform-Heartbeat pro adoptiertem Switch- und
  Gateway-Testgeraet durch den MITM.

Die Default-Lab-Zugangsdaten sind `admin` / `admin`. Fuer einen abweichenden
lokalen Lab-Controller koennen sie ueberschrieben werden:

```sh
UNIFI_STUB_LAB_ADMIN_USER=admin \
UNIFI_STUB_LAB_ADMIN_PASSWORD=... \
make integration-docker
```

## Cleanup-Semantik

Das Skript erzeugt pro Lauf Wegwerf-MACs/IPs, stoppt und entfernt temporaere
Stub-Container und Volumes und fordert den Controller auf, adoptierten State
fuer die Test-MACs zu loeschen. Controller-Volumes werden nicht zurueckgesetzt.

UniFi Network kann nicht adoptierte Pending-Zeilen bis zum Discovery-TTL im
Prozessspeicher behalten. Im beobachteten Docker-Lab sind diese Zeilen nicht in
MongoDB persistiert. Frische Wegwerf-MACs vermeiden Kollisionen zwischen
wiederholten Laeufen.

`UNIFI_STUB_DOCKER_KEEP_RESOURCES=1` nur setzen, wenn ein adoptiertes Testgeraet
oder das Stub-State-Volume nach einem fehlschlagenden Lauf absichtlich
inspiziert werden soll.

## Grenzen

Dieses Lab belegt container-lokale Linux-Bridge-/FDB-Observation, sysfs-Zaehler,
explizites Port-Mapping, Gateway-Tabellen-Rendering, Inform-Framing,
Controller-Adoption und lokale Adoption-State-Persistenz. Es belegt nicht
Proxmox-Host-Bridge-Verhalten, FreeBSD-Runtime-Verhalten, LLDP-Import oder
Event-Subscriptions.

Es belegt auch keine physische Topologie-Richtung. Container-Tests nutzen
Wegwerf-Identitaeten, decken also nicht den Fall ab, dass ein echter
Upstream-UniFi-Switch dieselbe physische Host-MAC bereits meldet. Reale
Proxmox- oder Bridge-Deployments sollten `uplink_neighbor`, `uplink_port` und
die Wahl zwischen synthetischer und physischer MAC gegen den Zielcontroller
validieren, bevor das Ergebnis als repraesentativ gilt.
