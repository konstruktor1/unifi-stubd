# Lab-Projektstand

Zuletzt aktualisiert: 2026-05-20.

Diese Seite verfolgt Lab- und Firmware-Referenzarbeit. Sie ist absichtlich vom
[Projektstand](project-status.md) getrennt, der den `unifi-stubd`-Daemon selbst
beschreibt.

Die Lab-Arbeit dient dazu, echtes UniFi-Firmware-Verhalten und Controller-
Erwartungen zu verstehen. Sie ist kein Produktcode und macht `unifi-stubd`
nicht zu einem vollstaendigen Gateway oder UniFi-OS-Ersatz.

## UDM Pro SE VM-Referenz

Die UDM-Pro-SE-QEMU/UTM-Arbeit ist ein Firmware-Referenzpfad, kein Ersatz fuer
den Go-Stub.

Was funktioniert:

- Der direkte Vendor-Kernel und der Vendor-U-Boot-Payload wurden auf QEMU
  `virt` getestet; dort kommt keine brauchbare serielle Ausgabe.
- Ein fremder, QEMU-virt-faehiger ARM64-Kernel kann den UDM-Initramfs/Rootfs-
  Pfad so weit booten, dass das UDM-Firmware-`systemd` startet.
- Das Lab-Initramfs behaelt UDM-Initramfs, UDM-Rootfs, Overlay-Layout und den
  finalen `/sbin/init`-Pfad bei, patcht aber die VM-spezifische Hardwaregrenze.
- Der gemockte Boot erreicht UDM-Firmware-`systemd`, installiert Userspace-
  Hardware-Mocks, startet UDM-nahe Dienste wie `uhwd`, `usd`, `rpsd`, nginx,
  SSH, UniFi OS Agent und `ubios-udapi-server` und erreicht den seriellen
  Login-Prompt.
- `network-init.service` laeuft im gemockten Pfad durch und erzeugt die
  erwarteten UDM-artigen Switch-/LAN-Devices.
- Die Web-Setup-Oberflaeche ist mit passender Lab-Netzkonfiguration erreichbar,
  wenn die VM laeuft.
- Der letzte UTM-Full-Test hat das Profil auf `UDM-Pro-SE-QEMU` angewendet, die
  serielle 4-GiB-VM gebootet, die Web-Setup-Oberflaeche auf der UTM-Shared/NAT-
  Gastadresse erreicht und `/api/system` mit `hasInternet=true` erhalten.
- In diesem UTM-Lauf zeigte der Network-nahe API-Status `eth9` als SFP+-WAN-
  Rolle, hielt `eth8` auf `br0` und fuehrte `br0` mit `192.168.1.1/24` plus
  Host-only-Alias `192.168.128.2/24`.
- Browser-Login und Refresh waren stabil, wenn derselbe Hostname durchgehend
  genutzt wurde. Ein Wechsel zwischen `localhost` und `127.0.0.1` erzeugt einen
  getrennten Cookie-Scope und kann wie ein Login-Loop wirken.

Aktuelles UTM-Mapping:

- `eth9`: erste 10G-SFP+-WAN-Rolle, hinter UTM `Shared` / NAT.
- `eth8`: 2.5G-RJ45-LAN-Rolle, hinter UTM `Host` Networking und auf `br0`.
- Die VM nutzt 4 GiB RAM und einen rein seriellen Anzeigeweg.
- Der Installer schreibt einen UTM-`Network:0:PortForward`-Eintrag von Host
  `127.0.0.1:10443` auf Gast-`443`, aber der letzte beobachtete UTM-CLI-Lauf
  hat diesen Host-Port nicht selbst gebunden. Bis das native UTM-Forwarding
  belegt ist, die direkte UTM-Shared/NAT-Gastadresse oder einen expliziten
  lokalen TCP-Forwarder verwenden.

So sind die Testergebnisse zu lesen:

- Der Docker-Test ist der schnellste Weg, UniFi-OS-Setup-UI, Core/API-
  Erwartungen, Support-Bundle-Erzeugung und die projekt-eigenen Fassaden zu
  pruefen. Fuer Boot-Validierung ist er schwaecher, weil er Host-Kernel,
  Docker-Netzwerk, Docker-Prozessaufsicht und Lab-Wrapper nutzt statt eines
  echten VM-Bootpfads.
- Der UTM-Test ist die bessere Referenz fuer die Frage "bootet das wie eine
  UDM?", weil er eine volle ARM64-VM startet, durch UDM-Initramfs/Rootfs-
  Uebergabe geht, Firmware-`systemd` startet, die Firmware-Netzinitialisierung
  ausfuehrt und den Gast-NICs stabile UDM-artige Rollen gibt.
- Die direkten QEMU-only-Tests scheiterten an der nativen Vendor-Hardwaregrenze:
  Vendor-Kernel und Vendor-U-Boot-Payload erzeugen auf QEMU `virt` keine
  brauchbare serielle Ausgabe, weil diese Maschine kein Annapurna-Labs-AL324-
  Board ist. Der Foreign-Kernel-QEMU-Pfad funktioniert als Lab-Bootpfad, ist
  fuer die Mac-seitige Browser-/Netzvalidierung aber weniger aussagekraeftig
  als UTM, weil das UTM-Profil die Zwei-NIC-Form aus Shared/NAT plus Host
  Networking besitzt, die im Full-Test genutzt wurde.

Wichtige Grenze:

- Der native Network-Self-Provisioning-/Self-Inform-Pfad ist noch in Arbeit.
  Ein hostseitiger `unifi-stubd`-Inform wurde nur als Diagnosebeweis genutzt,
  dass die Network-Anwendung den Device-Payload adoptieren kann. Das ist nicht
  die Zielarchitektur fuer die VM-Referenz.

Bekannte Restluecken:

- QEMU `virt` ist kein Annapurna-Labs-AL324-Boardmodell.
- Vendor-spezifische Kernelmodule wie `xt_dpi` und `xt_dyn_random` fehlen im
  fremden Kernel.
- Einige spaete Dienste scheitern noch oder starten neu, darunter Bluetooth,
  Logging, Directory und `mcagent`.
- Dummy-Devices in der VM erzeugen weiterhin einige `ethtool`-Warnungen.
- `/api/setup/support/generate` ist im Docker-Webportal-Pfad verlaesslich, lief
  im letzten UTM-VM-Lauf aber in einen Timeout und braucht noch VM-seitiges
  Debugging.
- Natives UTM-Localhost-Forwarding fuer `127.0.0.1:10443` ist im Profil
  konfiguriert, aber durch den letzten Test nicht bewiesen. Einen Hilfs-TCP-
  Forwarder nicht mit UTM-eigenem Port-Forwarding verwechseln.

## UDM Pro SE Docker-Webportal

Das Docker-UDM-Pro-SE-Profil ist ein Setup-/API-Inspektionspfad, nicht die
native VM-Boot-Referenz.

Was funktioniert:

- Der networkless Docker-Pfad startet `ubios-udapi-server`, `udapi-bridge` und
  `mcad` gegen deterministische `/mock`-Hardwaredaten.
- Der modulare C-`LD_PRELOAD`-Shim unter `mock/ldpreload/` leitet ausgewaehlte
  `/proc`-, `/sys`-, MTD- und Persistenzpfade um, stellt eine RTL8370-artige
  `swconfig`-ABI bereit und haelt unsichere Prozessaktionen eingegrenzt.
- `webportal.compose.yaml` stellt die Setup-UI auf
  `https://127.0.0.1:9443/` und eine HTTP-Vorschau auf
  `http://127.0.0.1:9080/` bereit.
- Die modulare CommonJS-Network-Fassade unter `network-app/` meldet Network als
  installiert/laufend, veroeffentlicht das paketierte UI-Manifest und liefert
  deterministische Setup-Payloads.
- Die modulare CommonJS-systemd-DBus-Fassade unter `systemd-dbus/` laesst
  UniFi Core bekannte Service-Units lesen, ohne systemd als PID 1 zu starten.
- `udapi-lab-shim.cjs` bildet Docker `eth0` als UDM-artige WAN-Sicht ab und
  liefert deterministische DNS-/ISP-Metadaten fuer Internet-Readiness-Checks.
- Der letzte Docker-Full-Test hat den ARM64-C-Shim neu gebaut, das Image
  `udm-pro-se-fw-sim:5.0.16` neu gebaut, den networkless Firmware-Pfad und den
  Webportal-Pfad gestartet und `unifi-core`, `ulp-go`, nginx, PostgreSQL, die
  DBus-Fassade, die Network-Fassade, UDAPI, `udapi-bridge` und `mcad`
  bestaetigt.
- Docker-HTTPS auf `https://127.0.0.1:9443/` lieferte die UniFi-OS-Setup-HTML
  mit `UNIFI_OS_MANIFEST`; `/api/system` meldete Internet-Readiness, nachdem
  sich der Core-/UDAPI-Readiness-Status gesetzt hatte.
- Die Docker-Support-Bundle-Erzeugung ueber `/api/setup/support/generate`
  lieferte ein lokales Archiv mit Lab-Systemmetadaten und `unifi-core`-Logs.

Wichtige Grenze:

- Docker laeuft weiterhin auf dem Host-Kernel und bootet keinen UDM-Kernel. Es
  mountet die gemeinsame deploybare Kernel-Ablage nur zum Vergleich und Logging.
  Fuer native Firmware-Boot-Fragen bleibt das QEMU/UTM-Profil massgeblich.
- Docker-Erfolg bedeutet daher: "die UI/API-Kompatibilitaetsschicht ist
  plausibel". Er beweist nicht Kernel-Boot, Initramfs-Uebergabe,
  systemd-Unit-Reihenfolge, VM-Netzwerk-Enumeration oder die native Network-
  Selbstsicht.

## Wo Die Lab-Arbeit Liegt

Wichtige Einstiegspunkte:

- `lab/gateway-profiles/udm-pro-se-vm/README.md`: QEMU/UTM-Runbook.
- `lab/gateway-profiles/udm-pro-se-vm/firmware.md`: Firmware-Findings und
  Boot-Status.
- `lab/gateway-profiles/udm-pro-se-vm/source-inventory.md`: Source-Grenze des
  VM-Profils.
- `lab/gateway-profiles/udm-pro-se/source-inventory.md`: Source-Grenze von
  UDM-Pro-SE-Dockerpfad und Mocks.

Projekt-eigene VM-Payloads:

- `lab/gateway-profiles/udm-pro-se-vm/initramfs/`: Initramfs-Hooks,
  Modullisten, Rootfs-Payload-Dateien, systemd-Units, Drop-ins sowie HTTP-/SSH-
  Templates fuer den VM-Pfad.
- `lab/gateway-profiles/udm-pro-se-vm/utm/`: versionierte UTM-Profileingaben
  mit Standardwerten fuer VM/Netzwerk, Kernel-Bootargs und Hinweisen zu den
  generierten UTM-Bundle-Dateien, die ausserhalb von Git bleiben.
- `lab/gateway-profiles/udm-pro-se-vm/kernel/`: Hinweise zur gemeinsamen,
  ignorierten Kernel-Deployment-Ablage unter `artifacts/deploy/kernel/`, die
  UTM und Docker nutzen.
- `lab/gateway-profiles/udm-pro-se/udapi-lab-shim.cjs`: Docker-Webportal-
  UDAPI-Lesewrapper fuer WAN-, DNS- und ISP-Metadaten.
- `lab/gateway-profiles/udm-pro-se/runtime/`: gesourcte Docker-Startup-Module,
  Wrapper-Skripte, nginx-Snippets, AWK-Filter, Templates und deterministische
  Lab-Daten fuer Firmware- und Webportal-Einstiegspunkte.
- `lab/gateway-profiles/udm-pro-se/network-app/`: modulare CommonJS-Network-
  Fassade fuer den Docker-Webportal-Pfad, getrennt nach Konfiguration, HTTP,
  Logging, Payloads, Routen und Websocket-Handling.
- `lab/gateway-profiles/udm-pro-se/systemd-dbus/`: modulare CommonJS-
  `org.freedesktop.systemd1`-Fassade fuer den Docker-Webportal-Pfad, getrennt
  nach DBus-Binding, Unit-Fixture, Interface- und Server-Modulen.
- `lab/gateway-profiles/udm-pro-se/mock/files/`: deterministische Mock-
  Dateisystemeingaben, die nach `/mock` kopiert werden.
- `lab/gateway-profiles/udm-pro-se/mock/ldpreload/`: modularer C-
  `LD_PRELOAD`-Shim:
  - `common.c`: Feature-Flags.
  - `auth.c`: enge Lab-Kompatibilitaet fuer Root-User-Pruefungen.
  - `response_patch.c`: byte-erhaltende Setup-/Readiness-Response-Patches.
  - `swconfig.c`: RTL8370-artige `libsw.so`-/OpenWrt-`swconfig`-ABI.
  - `fs_paths.c`, `fs_open.c`, `fs_io.c`, `process_control.c` und
    `socket_trace.c`: `/proc`, `/sys`, MTD, Prozess-, Socket- und Syscall-
    Interposition.

Ignorierte generierte Artefakte:

- `lab/gateway-profiles/udm-pro-se-vm/artifacts/`: kopierte Firmware,
  extrahierte Boot-Artefakte, VM-Disks, geholte Toolchains, gebaute Mock-Roots
  und generierte Initramfs-Images.

## Aktuellen VM-Pfad Neu Bauen

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
```

Direkt mit QEMU starten:

```sh
UDM_PRO_SE_FOREIGN_MODE=udm-systemd \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

Ein geklontes UTM-Profil konfigurieren:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/install-utm-profile.sh
```

Firmware-Images, extrahierte Firmware-Dateien, Captures, Keys, Tokens,
Zertifikate, Controller-URLs und private Lab-Daten bleiben aus Git draussen.
