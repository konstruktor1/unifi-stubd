# OPNsense API Generator How-to

`unifi-stubd-opnsense` reads OPNsense through read-only API calls and produces
a normal `unifi-stubd` YAML document for operator review. It is a companion
command, not a daemon hook, and does not change interfaces, routes, firewall
rules, or VLANs. Field behavior and merge rules are canonical in the
[generator reference](opnsense-generator-reference.md); use the
[sanitized source fixture](../../lab/stub/configs/hosts/opnsense-api-source.example.yaml)
as the template.

## 1. Create private API credentials

In `System > Access > Users`, select a dedicated read-only lab user, create an
API key, and download its key/secret pair once. OPNsense documents this under
[Use the API](https://docs.opnsense.org/development/how-tos/api.html) and
[Local Users & Groups](https://docs.opnsense.org/manual/how-tos/user-local.html).

Store only the raw values in private files on the OPNsense host:

```sh
install -d -m 0700 /usr/local/etc/unifi-stubd
umask 077
ee /usr/local/etc/unifi-stubd/opnsense-api-key
ee /usr/local/etc/unifi-stubd/opnsense-api-secret
chmod 600 /usr/local/etc/unifi-stubd/opnsense-api-key \
  /usr/local/etc/unifi-stubd/opnsense-api-secret
```

Never place credentials in source or generated YAML.

## 2. Install the companion command

Build through this consumer's configured Build Runner:

```sh
./build-runner/runtime/entry build-runner --project-name unifi-stubd \
  run --tier 1 --service go -- \
  env GOOS=freebsd GOARCH=amd64 go build \
    -o dist/unifi-stubd-opnsense ./cmd/unifi-stubd-opnsense
```

Use `GOARCH=arm64` for an ARM host, then copy the binary and the sanitized
fixture:

```sh
scp dist/unifi-stubd-opnsense root@opnsense.example:/usr/local/bin/
scp lab/stub/configs/hosts/opnsense-api-source.example.yaml \
  root@opnsense.example:/usr/local/etc/unifi-stubd/opnsense-source.yaml
ssh root@opnsense.example chmod 0755 /usr/local/bin/unifi-stubd-opnsense
ssh root@opnsense.example chmod 0600 \
  /usr/local/etc/unifi-stubd/opnsense-source.yaml
```

Edit the private source file for the intended lab. Use `Interfaces > Overview`
or `ifconfig -l` to identify existing local interfaces; do not copy real source
files or generated configurations back into this repository.

## 3. Generate, review, and install

Keep generated material in the private review directory outside the repository.
Validate inputs without contacting the API, generate one candidate, validate
and review the complete YAML, then install it explicitly:

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

The generator is not a live-sync service. Repeat the same review when local
interface assignments change.

## Remote workstation variant

Keep the copied base config, source file, and credentials in private deployment
storage outside the checkout. Generate on the workstation and stream the
candidate directly into the same target-host review path:

```sh
ssh root@opnsense.example \
  'install -d -m 0700 /var/db/unifi-stubd/review'
./dist/unifi-stubd-opnsense \
  -config /path/to/private/config.yaml \
  -source /path/to/private/opnsense-source.yaml |
ssh root@opnsense.example \
  'umask 077; cat > /var/db/unifi-stubd/review/unifi-stubd.generated.yaml'
```

Then perform the target-host validation, full review, and explicit installation
from step 3. Do not use a repository path as private deployment storage.
