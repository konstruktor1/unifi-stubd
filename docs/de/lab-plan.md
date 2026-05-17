# Lab Plan

## Ziel

Der MVP ist erreicht, wenn der Controller ein Fake-Switch-Geraet als adoptierbar sieht und nach Adoption dauerhaft als connected fuehrt.

## Empfohlener Aufbau

```text
UniFi Network Controller
  192.0.2.10

Linux host / Proxmox lab
  192.0.2.50
  unifi-stubd

Optional:
  echter UniFi Switch fuer Vergleichs-PCAPs
```

## Controller-Ports

Fuer dieses Projekt besonders relevant:

- UDP `10001`: Device discovery.
- TCP `8080`: Device inform.
- UDP `3478`: STUN, spaeter relevant.
- TCP `8443`: Controller UI/API.
- TCP `5671`: Traffic Flow Logging bei UXG, spaeter relevant.
- UDP `10101`: Client Fingerprinting, spaeter relevant.

Quelle: [UniFi Required Ports Reference](https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference)

## Mitschnitt

Auf dem Controller oder Mirror-Port:

```sh
sudo tcpdump -i any -nn -s0 -w unifi-inform.pcap 'udp port 10001 or tcp port 8080'
```

Auf dem Stub-Host:

```sh
sudo tcpdump -i any -nn -s0 'udp port 10001 or tcp port 8080'
```

## Debug-Logs

Im UniFi Controller sind diese Logbereiche besonders interessant:

- `discover`
- `inform`
- `devmgr`
- `ssh`

Typische Fehler, nach denen gesucht werden sollte:

- `invalid inform_ip`
- `inform decrypt error`
- `Inform Invalid`
- `ADOPTING -> UNKNOWN`
- `INFORM_ERROR`

## Lab-Reihenfolge

1. `make lint`
2. `make test`
3. `go run ./cmd/unifi-stubd -dry-run`
4. `go run ./cmd/unifi-stubd -once`
5. Controller pruefen: erscheint ein Geraet?
6. PCAP oeffnen und TLVs pruefen.
7. Inform-POST mit Default-Key implementieren.
8. Controller-Antwort dekodieren.
9. Adoption ausloesen und `setparam` speichern.

## Beispiel-Lab-Befehl

Beispiel-Lab-Stand:

- UniFi Controller: `192.0.2.10`
- Host-IP fuer den Weg zum Controller: `192.0.2.50`
- Fake-MAC: `02:11:22:33:44:55`
- Fake-Modell: `US16P150`
- Fake-Ports: `16`

Die Adressen aus `192.0.2.0/24` sind reine Dokumentationsbeispiele. Ersetze
sie durch Adressen aus einem isolierten Lab-Netz.

Discovery plus minimalen Inform-Heartbeat senden:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 192.0.2.50 \
  -hostname proxmox-vmbr0 \
  -controller http://192.0.2.10:8080/inform \
  -once
```

Nur direkten L3-Inform-Lauf ohne UDP-Discovery senden:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 192.0.2.50 \
  -hostname proxmox-vmbr0 \
  -controller http://192.0.2.10:8080/inform \
  -no-discovery \
  -once
```

Built-in SSH fuer Advanced Adoption:

```sh
sudo install -m 0755 unifi-stubd /usr/local/bin/unifi-stubd
sudo install -d -m 0755 /etc/unifi-stubd /var/lib/unifi-stubd
sudo install -m 0600 packaging/linux/etc/unifi-stubd/config.yaml /etc/unifi-stubd/config.yaml
sudo /usr/local/bin/unifi-stubd
```

Der Controller kann dann fuer Advanced Adoption `ubnt` / `ubnt` gegen Port `22` nutzen. Management-SSH des Lab-Systems sollte in diesem Aufbau auf einen anderen Port gelegt werden.

## Docker-Controller-Lab

Fuer lokale Gateway-Profil-Simulation ohne Vendor-Firmware gibt es das Docker
Compose Lab in `lab/controller-gateway-stubs.compose.yaml`. Es startet eine
UniFi Network Application, MongoDB, einen Inform-MITM und ein ausgewaehltes
`unifi-stubd` Stub-Profil.

Der generische Switch-Stub-Service `stub-us8` baut das Root-`Dockerfile` und
uebergibt `-profile us8` zur Laufzeit. Jedes Gateway-Profil hat ein eigenes
Dockerfile unter `lab/gateway-profiles/`; die Image-Entrypoints enthalten damit
das jeweilige Gateway-Profil. Compose liefert nur lab-spezifische
Laufzeitwerte wie MAC-Adresse, IP-Adresse, Hostname und Controller-URL.

Generischen US-8-Switch-Stub starten:

```sh
mkdir -p lab/captures
docker compose -f lab/controller-gateway-stubs.compose.yaml \
  --profile stub \
  up -d --build
```

Gateway-Lite-Profil starten:

```sh
mkdir -p lab/captures
docker compose -f lab/controller-gateway-stubs.compose.yaml \
  --profile uxg-lite \
  up -d --build
```

Verfuegbare Gateway-Stub-Profile sind `ugw3`, `uxg-lite`, `uxgpro` und
`ucg-fiber`. Das Compose-Profil `gateways` startet alle vier Gateway-Profile
fuer Payload-Vergleiche, aber Gateway-Adoption-Tests sollten normalerweise ein
Gateway pro frischer Controller-Site nutzen.

Die UI ist unter `https://localhost:8443` erreichbar. Im UniFi-Setup bleibt
TCP `8080` fuer Device-Kommunikation aktiv; als Inform-Host-Override wird
`unifi` gesetzt. Der Stub sendet nach `http://unifi:8080/inform`, der Traffic
geht durch den MITM-Container, und lokale Rohmitschnitte landen im ignorierten
Verzeichnis `lab/captures/`.

Die echten Firmware-Simulationsprofile werden separat in
`research/firmware/profiles.yaml` gefuehrt. Das sind nicht nur
`unifi-stubd`-Profile: jedes echte Firmware-Profil braucht ein lokales
Vendor-Firmware-Image, ein extrahiertes Rootfs, Architektur-Notizen und einen
passenden Prozess-Wrapper. Aktuell ist nur UXG-Pro `5.0.16` im Controller-Lab
adoptiert lauffaehig.

Das Compose-Lab nutzt das UniFi-Network-Application-Image von LinuxServer.io
mit separatem MongoDB-Container. Ubiquitis aktuelle Self-Hosting-Richtung ist
UniFi OS Server, aber Ubiquiti dokumentiert, dass dieser nicht als
eigenstaendiger Docker-/Podman-Container bereitgestellt wird.

OpenRC-Service:

```sh
sudo install -m 0755 packaging/linux/etc/init.d/unifi-stubd /etc/init.d/unifi-stubd
sudo rc-update add unifi-stubd default
sudo rc-service unifi-stubd restart
```

Systemd-Service:

```sh
sudo install -m 0644 packaging/linux/usr/lib/systemd/system/unifi-stubd.service /etc/systemd/system/unifi-stubd.service
sudo systemctl daemon-reload
sudo systemctl enable --now unifi-stubd.service
```

## Pakete bauen

Alle unterstuetzten Paketformate bauen:

```sh
make package
```

Ein einzelnes Format bauen:

```sh
make package-deb
make package-rpm
make package-arch
make package-tgz
make package-freebsd-tgz
```

Version, Release oder Zielarchitektur ueberschreiben:

```sh
PKG_VERSION=0.1.0 PKG_RELEASE=1 PKG_GOARCH=amd64 \
  PKG_MAINTAINER='Name <email@example.com>' make package
```

Die Ausgaben landen unter `dist/packages/`. Native Debian-, RPM- und Arch-Linux-Pakete werden mit nFPM aus `packaging/nfpm.yaml` gebaut; die Linux- und FreeBSD-`.tar.gz`-Pakete entstehen aus ihren OS-spezifischen Staging-Baeumen. FreeBSD/OPNsense ist aktuell Stub-only.

Layout:

- Code: `/usr/local/bin/unifi-stubd`
- Config: `/etc/unifi-stubd/config.yaml`
- Adoption-SSH-Key: `/etc/unifi-stubd/ssh_host_rsa_key`
- Runtime-State: `/var/lib/unifi-stubd/adoption.env`
- Logs: `/var/log/unifi-stubd.log`, `/var/log/unifi-stubd.err`

Lab-Switch-Identitaeten liegen unter `lab/`; der paketierte Linux-Dateibaum
und FreeBSD-Dateibaum ist in `packaging/installed-files.md` beschrieben.
