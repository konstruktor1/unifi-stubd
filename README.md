# unifi-stubd

`unifi-stubd` ist ein Lab-Projekt fuer einen minimalen UniFi-Device-Stub. Ziel ist, eine Proxmox-Bridge, OPNsense/pfSense-VM oder ein anderes Nicht-UniFi-System im UniFi Network Controller als sichtbares Geraet erscheinen zu lassen, ohne dass der Controller echte Hardware provisioniert.

Der erste sinnvolle Zieltyp ist ein Fake-UniFi-Switch, nicht ein Gateway. Switch-Emulation braucht deutlich weniger bewegliche Teile und kann trotzdem Topologie, Ports, MAC-Tabellen und grobe Traffic-Zaehler sichtbar machen.

## Zielbild

```text
Proxmox / OPNsense / Linux host
  unifi-stubd
    discovery: UDP 10001 TLV announcements
    inform:    TNBU packet encode/decode
    adoption:  authkey/cfgversion speichern
    payload:   fake 16-port switch status
    adapters:  Linux bridge FDB -> UniFi mac_table
```

## Nicht-Ziele

- Kein Ersatz fuer einen UniFi Gateway.
- Keine echte Controller-Provisionierung auf dem Host ausfuehren.
- Keine Port- oder Firewall-Konfiguration blind aus Controller-Kommandos anwenden.
- Keine Annahme, dass UniFi-DPI ohne UniFi Gateway vollstaendig reproduzierbar ist.

## Quick Start

Der aktuelle Code ist ein erster Starter. Er kann Discovery-Pakete bauen/senden und optional ein minimales `/inform` an den Controller schicken.

```sh
cd /Users/corspi/Nextcloud/Projekte/codes/unifi-stubd
go test ./...
go run ./cmd/unifi-stubd -dry-run
```

Discovery im Lab senden:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac auto \
  -ip 192.168.1.50 \
  -hostname auto
```

Bekannte Profile anzeigen:

```sh
go run ./cmd/unifi-stubd -list-profiles
```

Discovery plus minimalen Inform-Heartbeat direkt an den Controller senden:

```sh
go run ./cmd/unifi-stubd \
  -mac auto \
  -ip 192.168.1.50 \
  -hostname auto \
  -controller http://192.168.1.10:8080/inform \
  -once
```

Nur L3-Inform ohne UDP-Discovery:

```sh
go run ./cmd/unifi-stubd \
  -controller http://192.168.1.10:8080/inform \
  -no-discovery \
  -once
```

Built-in SSH fuer Advanced Adoption aktivieren:

```sh
sudo ./unifi-stubd \
  -profile us16p150 \
  -mac auto \
  -ip 10.0.0.151 \
  -hostname auto \
  -controller http://192.168.1.10:8080/inform \
  -uplink-speed auto \
  -ssh-listen 0.0.0.0:22 \
  -ssh-user ubnt \
  -ssh-password ubnt
```

Der SSH-Server laeuft dann im `unifi-stubd`-Prozess selbst und beantwortet Kommandos wie `syswrapper.sh set-adopt ...` oder `mca-cli-op set-inform ...`.

Fuer ein 10G-Testprofil:

```sh
sudo ./unifi-stubd \
  -profile us16xg \
  -mac auto \
  -ip 10.0.0.151 \
  -hostname auto \
  -controller http://192.168.1.10:8080/inform \
  -uplink-speed auto \
  -ssh-listen 0.0.0.0:22
```

`-hostname auto` uebernimmt den Hostnamen des Systems. `-mac auto` erzeugt eine stabile lokal administrierte MAC aus Hostname und Profil; dadurch bekommt ein neues Profil automatisch eine neue Geraeteidentitaet im Controller. `-uplink-speed auto` versucht den Egress-Link zum Controller zu finden und dessen Speed fuer Port 1 zu melden; wenn der Treiber keine Geschwindigkeit liefert, bleibt der Profilwert aktiv.

## Projektstand

- [x] Recherche und Projektzuschnitt
- [x] Discovery-Paket-Builder
- [x] Inform-Packet-Encoder/Decoder-Grundlage
- [x] Minimaler `/inform` Client mit Default-Key
- [x] Minimaler Fake-Switch-Payload
- [x] Built-in SSH fuer Advanced Adoption
- [ ] Controller-Lab gegen eine gepinnte UniFi Network Version
- [ ] Adoption-State-Machine
- [ ] Linux-Bridge-FDB in `port_table[].mac_table` einspeisen
- [ ] Systemd-Service fuer Proxmox/OPNsense-Lab

## Dokumentation

- [Research](docs/research.md)
- [Protocol Notes](docs/protocol-notes.md)
- [Architecture](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
- [Lab Plan](docs/lab-plan.md)
- [Security Notes](docs/security.md)
