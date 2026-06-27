# OPNsense API Generator How-to

`unifi-stubd-opnsense` is a companion generator. It does not run inside the
`unifi-stubd` daemon and it does not change OPNsense interfaces, routes,
firewall rules, or VLANs. It reads OPNsense through read-only API calls and
creates a normal `unifi-stubd` YAML file for review.

The current packages install `unifi-stubd`. Until packaging for the companion
tool is added, build `unifi-stubd-opnsense` separately and copy it to the
OPNsense host when you want to run the generator there.
For field-level behavior, merge rules, endpoints, and troubleshooting, see the
[OPNsense API Generator Reference](opnsense-generator-reference.md).

## 1. Create an OPNsense API key

In the OPNsense WebGUI:

1. Open `System > Access > Users`.
2. Select the user that should own this integration, or create a dedicated
   read-only lab user.
3. In the user's API key section, create a new key.
4. Download the generated key/secret file once and keep it private.

OPNsense API keys are key/secret pairs. The key is used as the HTTP Basic Auth
username and the secret as the password.
The OPNsense documentation covers the same API key flow in
[Use the API](https://docs.opnsense.org/development/how-tos/api.html) and
[Local Users & Groups](https://docs.opnsense.org/manual/how-tos/user-local.html).

## 2. Install the companion binary on OPNsense

Build the binary on a development host from this repository:

```sh
GOOS=freebsd GOARCH=amd64 go build \
  -o dist/unifi-stubd-opnsense \
  ./cmd/unifi-stubd-opnsense
```

Use `GOARCH=arm64` for an ARM FreeBSD/OPNsense host. Then copy the binary:

```sh
scp dist/unifi-stubd-opnsense root@opnsense.example:/usr/local/bin/
ssh root@opnsense.example chmod 0755 /usr/local/bin/unifi-stubd-opnsense
```

On the OPNsense shell, verify that it starts:

```sh
/usr/local/bin/unifi-stubd-opnsense -h
```

## 3. Store API credentials on OPNsense

On the OPNsense shell, create private files for the API key and secret:

```sh
mkdir -p /usr/local/etc/unifi-stubd
chmod 700 /usr/local/etc/unifi-stubd
umask 077
ee /usr/local/etc/unifi-stubd/opnsense-api-key
ee /usr/local/etc/unifi-stubd/opnsense-api-secret
chmod 600 /usr/local/etc/unifi-stubd/opnsense-api-key
chmod 600 /usr/local/etc/unifi-stubd/opnsense-api-secret
```

Paste only the raw key into `opnsense-api-key` and only the raw secret into
`opnsense-api-secret`. Do not put the key or secret into the source YAML.

## 4. Identify the OPNsense interfaces

Use OPNsense UI `Interfaces > Overview` or the shell:

```sh
ifconfig -l
ifconfig ixl0
ifconfig vtnet0
```

For a UXG-Pro-shaped lab profile, the represented UniFi ports are fixed profile
data:

```text
port 1 -> eth0, profile role wan,  1G RJ45
port 2 -> eth1, profile role lan,  1G RJ45
port 3 -> eth2, profile role wan2, 10G SFP+
port 4 -> eth3, profile role lan2, 10G SFP+
```

If real OPNsense WAN is `ixl0` and should appear as UniFi physical port 3, map
`port: 3` to `interface: ixl0`. The controller-facing `ifname` stays `eth2`;
`ixl0` is generated as `source_interface`.

## 5. Create the OPNsense source file

Create `/usr/local/etc/unifi-stubd/opnsense-source.yaml` on OPNsense:

```sh
ee /usr/local/etc/unifi-stubd/opnsense-source.yaml
chmod 600 /usr/local/etc/unifi-stubd/opnsense-source.yaml
```

Example:

```yaml
base_url: https://127.0.0.1
api_key_file: /usr/local/etc/unifi-stubd/opnsense-api-key
api_secret_file: /usr/local/etc/unifi-stubd/opnsense-api-secret
api_key_env: ""
api_secret_env: ""
ca_file: ""
insecure_skip_verify: false
timeout_ms: 2000
uplink_port: 3
gateway_status: true
interfaces:
  - port: 3
    interface: ixl0
    name: WAN SFP+
    role: wan
    network_group: WAN
    network_name: opnsense_wan
    vlan: 3
  - port: 4
    interface: vtnet0
    name: LAN SFP+ to server-lan1
    role: lan
    network_group: LAN
    network_name: opnsense_lan
    vlan: 1
wan_health:
  source: static
  interval_seconds: 10
  timeout_ms: 1000
  targets: []
```

Use `insecure_skip_verify: true` only for an explicit lab endpoint with a
self-signed certificate that you cannot validate with `ca_file`.

## 6. Validate locally without API calls

```sh
/usr/local/bin/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml \
  -validate
```

This checks the base config, source YAML, and credential loading. It does not
contact OPNsense.

## 7. Generate and review the config

Print the generated config to a temporary file:

```sh
/usr/local/bin/unifi-stubd-opnsense \
  -config /usr/local/etc/unifi-stubd/config.yaml \
  -source /usr/local/etc/unifi-stubd/opnsense-source.yaml \
  > /tmp/unifi-stubd.generated.yaml
```

Validate the generated `unifi-stubd` config:

```sh
/usr/local/bin/unifi-stubd \
  -validate \
  -config /tmp/unifi-stubd.generated.yaml
```

Review the generated port overrides before installing them:

```sh
grep -n "port_overrides" /tmp/unifi-stubd.generated.yaml
grep -n "source_interface\\|interface: ixl0\\|role: wan" /tmp/unifi-stubd.generated.yaml
```

## 8. Install the generated config

Keep a backup, then replace the service config only after review:

```sh
cp -p /usr/local/etc/unifi-stubd/config.yaml \
  /usr/local/etc/unifi-stubd/config.yaml.before-opnsense
install -m 0600 /tmp/unifi-stubd.generated.yaml \
  /usr/local/etc/unifi-stubd/config.yaml
service unifi-stubd restart
/usr/local/bin/unifi-stubd -status-json
```

The generator is not a live sync service. Re-run it when OPNsense interface
assignments change and review the generated YAML again before replacing the
daemon config.

## Remote workstation variant

You can also run the generator from a development workstation. In that case,
copy the current stub config from OPNsense, keep the API source and credential
files local to the workstation, and point `base_url` at the OPNsense management
address:

```sh
scp root@opnsense.example:/usr/local/etc/unifi-stubd/config.yaml ./config.opnsense.yaml
go run ./cmd/unifi-stubd-opnsense \
  -config ./config.opnsense.yaml \
  -source ./opnsense-source.yaml \
  > generated.yaml
scp generated.yaml root@opnsense.example:/tmp/unifi-stubd.generated.yaml
```

Then validate and install `/tmp/unifi-stubd.generated.yaml` on OPNsense as
shown above.
