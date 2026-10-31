# OPNsense-API-Generator How-to

`unifi-stubd-opnsense` liest OPNsense ueber read-only API-Aufrufe und erzeugt
ein normales `unifi-stubd`-YAML-Dokument zur Betreiberpruefung. Das Companion-
Kommando ist kein Daemon-Hook und aendert keine Interfaces, Routen, Firewall-
Regeln oder VLANs. Feldverhalten und Merge-Regeln stehen verbindlich in der
[Generator-Referenz](opnsense-generator-reference.md); als Vorlage dient die
[bereinigte Source-Fixture](../../lab/stub/configs/hosts/opnsense-api-source.example.yaml).

## 1. Private API-Credentials erstellen

Unter `System > Access > Users` einen dedizierten read-only Lab-User waehlen,
einen API-Key erzeugen und das Key-/Secret-Paar einmal herunterladen. OPNsense
dokumentiert dies unter
[Use the API](https://docs.opnsense.org/development/how-tos/api.html) und
[Local Users & Groups](https://docs.opnsense.org/manual/how-tos/user-local.html).

Nur die rohen Werte in privaten Dateien auf dem OPNsense-Host speichern:

```sh
install -d -m 0700 /usr/local/etc/unifi-stubd
umask 077
ee /usr/local/etc/unifi-stubd/opnsense-api-key
ee /usr/local/etc/unifi-stubd/opnsense-api-secret
chmod 600 /usr/local/etc/unifi-stubd/opnsense-api-key \
  /usr/local/etc/unifi-stubd/opnsense-api-secret
```

Credentials niemals in Source- oder generiertes YAML schreiben.

## 2. Companion-Kommando installieren

Den einmal definierten [nativen Kontext und aktuellen Einstieg](ci-cd.md) laden,
dann den Build ueber den konfigurierten Build Runner dieses Consumers ausfuehren:

```sh
"$BUILD_RUNNER_ENTRY" --project-name unifi-stubd \
  run --tier 1 --service go -- \
  env GOOS=freebsd GOARCH=amd64 go build \
    -o dist/unifi-stubd-opnsense ./cmd/unifi-stubd-opnsense
```

Fuer einen ARM-Host `GOARCH=arm64` verwenden, dann Binary und bereinigte
Fixture kopieren:

```sh
scp dist/unifi-stubd-opnsense root@opnsense.example:/usr/local/bin/
scp lab/stub/configs/hosts/opnsense-api-source.example.yaml \
  root@opnsense.example:/usr/local/etc/unifi-stubd/opnsense-source.yaml
ssh root@opnsense.example chmod 0755 /usr/local/bin/unifi-stubd-opnsense
ssh root@opnsense.example chmod 0600 \
  /usr/local/etc/unifi-stubd/opnsense-source.yaml
```

Die private Source-Datei fuer das vorgesehene Lab bearbeiten. Vorhandene lokale
Interfaces ueber `Interfaces > Overview` oder `ifconfig -l` bestimmen; reale
Source-Dateien oder generierte Configs nie in dieses Repository kopieren.

## 3. Generieren, pruefen und installieren

Generiertes Material im privaten Review-Verzeichnis ausserhalb des Repositories
halten. Zuerst Inputs ohne API-Aufruf validieren, dann genau einen Kandidaten
erzeugen, das vollstaendige YAML validieren und pruefen und es erst danach
explizit installieren:

```sh
review_config=/var/db/unifi-stubd/review/unifi-stubd.generated.yaml
install -d -m 0700 /var/db/unifi-stubd/review

/usr/local/bin/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml \
  -validate

umask 077
/usr/local/bin/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml \
  -out "$review_config"
/usr/local/bin/unifi-stubd -validate -config "$review_config"
ee "$review_config"

cp -p /usr/local/etc/unifi-stubd/config.yaml \
  /usr/local/etc/unifi-stubd/config.yaml.before-opnsense
install -m 0600 "$review_config" /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
/usr/local/bin/unifi-stubd -status-json
```

Der Generator ist kein Live-Sync-Dienst. Bei geaenderten lokalen Interface-
Zuweisungen denselben Review wiederholen.

## Variante vom Entwicklungsrechner

Kopierte Basisconfig, Source-Datei und Credentials in privatem Deployment-
Speicher ausserhalb des Checkouts halten. Den Kandidaten auf dem
Entwicklungsrechner erzeugen und direkt in denselben Review-Pfad des Zielhosts
streamen:

```sh
ssh root@opnsense.example \
  'install -d -m 0700 /var/db/unifi-stubd/review'
./dist/unifi-stubd-opnsense \
  -config /path/to/private/config.yaml \
  -source /path/to/private/opnsense-source.yaml |
ssh root@opnsense.example \
  'umask 077; cat > /var/db/unifi-stubd/review/unifi-stubd.generated.yaml'
```

Danach Validierung, vollstaendiges Review und explizite Installation auf dem
Zielhost aus Schritt 3 ausfuehren. Repository-Pfade sind kein privater
Deployment-Speicher.
