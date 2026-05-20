# Testleitfaden

Diese Seite beschreibt, welche Teststufe fuer welche Aenderung sinnvoll ist.

## Lokaler Standard-Gate

Vor einem Commit ausfuehren:

```sh
make check
git diff --check
```

`make check` prueft Lint-Konfiguration, fuehrt Lint aus, erzwingt Repo-Policy,
validiert paketierte Config-/Profil-YAMLs und startet `go test ./...`.

## Docker-Controller-Gate

Ausfuehren, wenn Adoption, Inform, Payload-Shape, Profile oder Controller-
Kompatibilitaet geaendert wurden:

```sh
make integration-docker
```

Das Docker-Lab ist der Referenz-Controller-Pfad. Es validiert den gepinnten
UniFi-Network-Application-Container, Pending Adoption, controller-getriggerte
Adoption, persistierten lokalen State und ausgewaehlte Dry-Run-Payloads.

## Package-Gate

Ausfuehren, wenn paketierte Configs, Service-Dateien, Filesystem-Pfade oder
Release-Metadaten geaendert wurden:

```sh
make package
make package-freebsd-tgz
```

Paketinhalte nur inspizieren, ausser der Test verlangt explizit eine temporaere
Installation. Permanente Service-Aktivierung gehoert nicht in einen Package-
Smoke-Test.

## Reales Linux-Bridge-Gate

Nutzen fuer `bridge-observe`-Verhalten, das Docker nicht beweisen kann:
Topologie-Richtung, physische Uplinks, SFP-/SFP+-Platzierung und Interaktion
mit echten Upstream-UniFi-Switches.

Read-only Preflight:

```sh
ip link show
bridge fdb show br <bridge>
cat /sys/class/net/<iface>/speed
cat /proc/net/dev
```

Danach temporaeren Dry-Run oder einen kontrollierten Controller-Test mit
disposable MAC/State ausfuehren.

## Reales FreeBSD-/OPNsense-Gate

Nutzen fuer FreeBSD-Parsing und Runtime-Smoke-Tests. Nicht permanent
installieren.

Read-only Preflight:

```sh
ifconfig
ifconfig <bridge> addr
tail /var/log/messages
```

Temporaere Extraktion und temporaere State-Pfade nutzen. Fehlende Tools werden
als skipped dokumentiert und nicht waehrend dieses Gates installiert.

## Controller-Lab-Gate

Nur mit disposable MACs und explizitem Cleanup nutzen.

Checkliste:

- zuerst dry-run;
- keine echten MAC-/IP-Kollisionen;
- ein disposable Device pro Profiltest;
- Controller Forget/Remove nach dem Test;
- lokaler State-Cleanup nach Stop;
- keine Controller-Tokens, privaten URLs oder echten MACs committen.

## Was welcher Test beweist

| Test | Beweist | Beweist nicht |
| --- | --- | --- |
| `go test ./...` | Unit- und Fixture-Verhalten | Live-Controller-Kompatibilitaet |
| `make check` | Lint, Policy, Validierung, Tests | Package-Install-Verhalten |
| `make integration-docker` | gepinnter Controller-Adoption-Pfad | physische Topologie-Richtung |
| Package-Builds | Artefakte sind baubar | Runtime-Korrektheit auf Zielhost |
| reale Linux-Bridge | physische Bridge-Observation | FreeBSD-Verhalten |
| realer FreeBSD-Host | FreeBSD-Runtime-Basics | Linux-Bridge-Verhalten |
| realer Controller | Adoption gegen deployed Controller | Sicherheit bei MAC-/IP-Kollisionen |

