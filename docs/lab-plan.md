# Lab Plan

## Ziel

Der MVP ist erreicht, wenn der Controller ein Fake-Switch-Geraet als adoptierbar sieht und nach Adoption dauerhaft als connected fuehrt.

## Empfohlener Aufbau

```text
UniFi Network Controller
  192.168.1.10

Linux host / Proxmox lab
  192.168.1.50
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

## Testreihenfolge

1. `go run ./cmd/unifi-stubd -dry-run`
2. `go run ./cmd/unifi-stubd -once`
3. Controller pruefen: erscheint ein Geraet?
4. PCAP oeffnen und TLVs pruefen.
5. Inform-POST mit Default-Key implementieren.
6. Controller-Antwort dekodieren.
7. Adoption ausloesen und `setparam` speichern.

## Lokaler Testbefehl

Aktueller Lab-Stand:

- UniFi Controller: `10.0.0.194`
- Host-IP fuer den Weg zum Controller: `10.0.0.150`
- Fake-MAC: `02:11:22:33:44:55`
- Fake-Modell: `US16P150`
- Fake-Ports: `16`

Discovery plus minimalen Inform-Heartbeat senden:

```sh
cd /Users/corspi/Nextcloud/Projekte/codes/unifi-stubd && go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 10.0.0.150 \
  -hostname proxmox-vmbr0 \
  -controller http://10.0.0.194:8080/inform \
  -once
```

Nur direkten L3-Inform-Test ohne UDP-Discovery senden:

```sh
cd /Users/corspi/Nextcloud/Projekte/codes/unifi-stubd && go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 10.0.0.150 \
  -hostname proxmox-vmbr0 \
  -controller http://10.0.0.194:8080/inform \
  -no-discovery \
  -once
```

Built-in SSH fuer Advanced Adoption:

```sh
sudo unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 10.0.0.151 \
  -hostname unifi-stubd-lab \
  -controller http://10.0.0.194:8080/inform \
  -ssh-listen 0.0.0.0:22 \
  -ssh-user ubnt \
  -ssh-password ubnt
```

Der Controller kann dann fuer Advanced Adoption `ubnt` / `ubnt` gegen Port `22` nutzen. Management-SSH des Lab-Systems sollte in diesem Aufbau auf einen anderen Port gelegt werden.
