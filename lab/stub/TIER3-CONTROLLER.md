# Candidate qualification and controller contract

Test Runner owns the registered Linux ARM64 testhost's VM lifecycle through
public `start`, `wait`, `stop` and `destroy`. Core's candidate qualifier invokes
those operations through the current native Build Runner binding and owns its
bounded guest fixture, exact candidate process, SSH tunnel and machine evidence.
The companion supplies authoritative machine definitions, controller preparation
and fresh per-lane provenance/authentication. With schema 1, Core owns one disposable
controller Compose project and its API/browser helpers. With schema 2, Labs
owns the external UniFi OS controller, data and endpoint tunnels; Core owns
only a disposable Compose project containing API/browser helpers.
This boundary requires a local isolated Docker daemon whose published loopback
port is reachable on the host opening the guest SSH connection.

Test Runner has no public detached-execution operation for an already running
guest without resetting it. Public `run` can execute a guest job when it owns
the complete generate/reset/start/execute/cleanup lifecycle. It cannot attach
as post-start execution to the guest already participating in this controller
handshake while preserving that live lifecycle. This specific interface gap
limits consumer SSH here to the bounded guest fixture/candidate process,
live executable/status evidence and reverse inform tunnel. The tunnel itself
has no Test Runner interface. The qualifier executes these consumer steps;
Test Runner owns the VM boundary rather than executing the qualification job.
Its available public lifecycle operations remain in use.

Run the product qualifier from Core's native project/session context:

```sh
make qualify-candidate
make qualify-candidate-matrix
```

Both require the already selected `LAB_DATA_ROOT` domain-data locator; this is
not another provider root or a placement change. Its existing
`tier3/machines/unifi-stubd-lab/` role owns private evidence and control data.
The registered Test Runner owns VM copies and resource cleanup in its existing
native location. No default path or VM move is implied. Inherit the single
`PROVIDER_CONSUMER_ROOT` and `PROVIDER_CONSUMER_CONTEXT` unchanged from Core's
native owner. `UNIFI_STUBD_CORE_ROOT` is obsolete and rejected.
The saved Build Runner source, working directory and `PYTHON` interpreter must
match that same native binding; setting only a second root is insufficient.

Supply these literal domain inputs for both targets:

- `TIER3_CANDIDATE_BINARY`, the retained
  `dist/candidates/<candidate>/bin/unifi-stubd_linux_arm64` absolute path;
  `TIER3_CANDIDATE_SHA256`, `TIER3_CANDIDATE_VERSION` and
  `TIER3_CANDIDATE_MAC` (local `02:` address) and `TIER3_CANDIDATE_IP`, the
  selected guest's actual IPv4 address (never a committed lab address);
- `TIER3_SSH_IDENTITY_FILE` and `TIER3_SSH_KNOWN_HOSTS_FILE`, existing absolute
  private inputs matching the native transport, with its host/port endpoint
  already trusted by normal strict OpenSSH known-host checking;
- `TIER3_TEST_RUNNER_MANIFEST`, a Core-owned testhost template, and its existing
  `TIER3_TEST_RUNNER_HOST_SERVICE` and `TIER3_TEST_RUNNER_SSH_PORT` binding.

The read-only `tools/qualification_binding.py` validates these literal SSH inputs
against the existing Core native Test Runner TOML record, product source and seed
role. It reads no private-key bytes and emits no private inventory fields.
Guest internal IP and SSH transport trust are separate. The qualifier verifies
the requested IP on the actual guest before changing its fixture or service.

Single-lane qualification additionally requires a unique `TIER3_RUN_ID` and
all applicable controller fields listed below with the `TIER3_` prefix.
The matrix instead requires `TIER3_MATRIX_RUN_ID`; it freshly resolves current
official Stable and RC releases. `TIER3_MATRIX_SCOPE=stable-only` explicitly
runs only Stable and retains that scope in its result; it proves no RC or
combined acceptance. Schema 2 requires distinct freshly provisioned controller
instances and, for each selected lane, `TIER3_STABLE_CONTROLLER_` or
`TIER3_RC_CONTROLLER_` plus `API_ORIGIN`, `INFORM_PORT`, `AUTH_FILE` and
`PROVENANCE_FILE`. A single schema-2 lane uses the same suffixes prefixed
`TIER3_CONTROLLER_`. Labs must supply actual preparation/reset evidence;
asserting `fresh_controller=true` alone proves no reset.

Maintained orchestration, strict data parsing and native context validation live
in `scripts/qualify-candidate*.sh` and `scripts/lib/` relative to this directory.
`configs/qualification.env` contains only the active product fixture parameters;
`configs/qualification-contract-files.txt` lists the Core source inputs whose
current bytes and modes are bound during a run. Candidate source archives and
hashes are provenance, never execution selectors. Generated handshakes remain
under `dist/`; private machine evidence remains in the assigned data role.
No firmware, disk, provider implementation copy or Labs executable is imported.

From Core, use:

```sh
UNIFI_STUB_TIER3_GATE_MANIFEST=dist/tier3-controller-input/<run-id>.env \
  make integration-nonunifi-metrics
```

This canonical Make entry forwards the explicit manifest into the prepared
Build Runner. The matrix entry immediately delegates a nonempty manifest to
this bounded Tier 3 gate; the gate validates its workspace and input before
runtime operations. Without a manifest, the existing standalone Stable/RC
matrix remains the default. `integration-tier3-candidate-controller-cleanup`
retains its exact-run recovery purpose. A bound
Stable run qualifies only that lane; it does not imply RC acceptance.

The input is a regular, non-executable `KEY=value` file, one value per key, no
whitespace or duplicate keys. Schema 1 requires exactly:

- `SCHEMA_VERSION=1`, `RUN_ID` matching the filename;
- `CONTROLLER_CHANNEL` (`stable` or `release-candidate`),
  `CONTROLLER_SELECTOR=latest`, `CONTROLLER_EXPECTED_VERSION`,
  `CONTROLLER_RELEASE_STAGE` (`GA` or `RC`), `CONTROLLER_RELEASE_ID`,
  `CONTROLLER_RELEASE_URL`, `CONTROLLER_RELEASE_CREATED_AT`,
  `CONTROLLER_RESOLVED_AT`;
- `CONTROLLER_ZIP_URL`, `CONTROLLER_ZIP_SHA256`,
  `CONTROLLER_RELEASE_QUERY_ENDPOINT`, `CONTROLLER_RELEASE_QUERY_SHA256`,
  `CONTROLLER_RELEASE_RESPONSE_SHA256`,
  `CONTROLLER_RUNTIME_BASE_SELECTOR=lscr.io/linuxserver/unifi-network-application:latest`;
- `CANDIDATE_MAC` (local `02:` address), `CANDIDATE_SHA256`,
  `CANDIDATE_VERSION`, `CANDIDATE_IP`, `EXPECTED_UPLINK_SPEED_MBPS` (observed
  on the guest fixture).

The caller freshly resolves the official published release. Core verifies the
version/official ZIP binding and its SHA-256, pulls the current native runtime
base, records its immutable digest and checks the running controller version.
An unavailable current release is an open acceptance prerequisite; an older
release must not be substituted.

Schema 2 uses the same 18 common fields and replaces the three schema-1-only
fields (`CONTROLLER_ZIP_URL`, `CONTROLLER_ZIP_SHA256`,
`CONTROLLER_RUNTIME_BASE_SELECTOR`) with these four fields, for exactly 22:

- `SCHEMA_VERSION=2` and `CONTROLLER_SELECTOR=latest`;
- `CONTROLLER_API_ORIGIN=https://host.docker.internal:<port>`, an isolated
  Labs endpoint reachable from the Docker helper network;
- `CONTROLLER_INFORM_PORT`, the Labs-owned host-loopback inform endpoint;
- `CONTROLLER_AUTH_SHA256` and `CONTROLLER_PROVENANCE_SHA256`.

The authoritative field list is `configs/tier3-controller-fields.txt` relative
to this directory. The qualifier and gate read it as data; neither sources the
input manifest. Schema 1 and its cancellation fixture remain supported.

The schema-2 caller provides private, regular, non-symlink files next to the
input manifest: `<run-id>.auth.json` and `<run-id>.controller.json`, both mode
0600, in a real input directory with mode 0700. The Core qualifier sets that mode
on its exact Core input namespace even when schema 1 created it earlier.
Authentication is exactly a JSON object containing nonempty `username`
and `password` strings, at most 65536 bytes. Core verifies both sidecar hashes
and streams credentials on stdin with `compose run -T --rm`. The maintained
helper entrypoint bounds stdin, writes mode 0600
`/run/unifi-stubd-controller/auth.json` inside the disposable container and
passes its path to the API/browser client. Credentials never become Compose
environment values, command arguments or image layers. The qualifier removes
its matching derived Core authentication copy after confirmed helper/guest
cleanup; failed cleanup retains it for owner recovery. The external original
and provenance/bindings remain unchanged.

The browser grants local-network access only to the selected controller origin
so Chromium can open that controller's local WebSockets. Controller CSRF
headers are added only to requests whose origin has been checked. The single
public GET to `https://static.ui.com/fingerprint/ui/public.json` uses an independent
Node fetch without controller cookies, credentials or headers, with normal
TLS validation, no redirects and a streaming 4 MiB JSON limit before retaining
the next chunk. Other external browser
requests remain blocked. Failed UI checks retain bounded, redacted diagnostics
in the run's private `controller-ui.log`.
The UniFi OS check opens the native Port Manager with the target-device filter
and reads observed speed, RX/TX packets and RX/TX bytes from the same unique,
visible port row.

Controller provenance is at most 4 MiB. Core wraps the Labs owner's
JSON as `{"schema_version":1,"run_id":"<run-id>","source":{...}}` without
inventing a runtime identity. `source` requires `schema_version=1`,
`owner=unifi-stubd-labs`, `controller_channel`, `controller_version`,
`fresh_controller=true`, a nonempty `controller_instance`, the authoritative
inventory `machine_id`, and lowercase 64-digit SHA-256 values for
`testhost_manifest_sha256`, `controller_image_sha256` and
`controller_package_sha256`. Additional source facts may retain exact original
artifact provenance. The source channel/version must match the freshly resolved
release. A freshness assertion is owner evidence, not verification that a reset
or provisioning operation ran. Stable and RC require separate fresh instances;
the Core matrix rejects equal instance identifiers. This handoff does not
create, reset or clean up those external controllers. An unready controller,
missing per-lane provisioning, actual reset proof or controller cleanup remains
an open Labs owner prerequisite.

For an isolated Runner cancellation test only, the same 21-field input may use
`CONTROLLER_SELECTOR=cancellation-fixture`. This explicitly pins a historical
controller release and its original official query/response and ZIP digests;
keep the current channel query as separate evidence. Candidate identity and
positive uplink speed are synthetic test inputs, not observed guest evidence.
Use the same Make entry, a fresh run ID, a local synthetic MAC and a documentation
IP address. Start no guest or candidate. Cancel the exact owning Runner session
after its matching `ready.env` appears. Even if that MAC becomes visible, this
mode must fail before adoption, connection, UI checks or `accepted.env`; it can
never return `RESULT=passed`. The Core qualifier accepts only `latest` for product
qualification. A successful cancellation test requires the public cancel command
to succeed, the owning invocation to terminate, `RESULT=failed`,
`ADOPTED=false`, `CONNECTED=false`, `TELEMETRY=false`, `CLEANUP=passed` and
positive confirmation that the exact owned backend resources are absent.

Files under `dist/tier3-controller/<run-id>/` form a bounded handshake:

1. `ready.env` repeats the input and adds `MANIFEST_SHA256`,
   `CONTROLLER_ACTUAL_VERSION`, `CONTROLLER_RUNTIME_BASE_DIGEST`,
   `INFORM_HOST=127.0.0.1`, `INFORM_PORT` and `GUEST_INFORM_PORT=8080`.
   Core retains and reads back the controller's loopback inform override.
   Schema 2 repeats all 22 fields, including both sidecar hashes and endpoints,
   and omits `CONTROLLER_RUNTIME_BASE_DIGEST`; the external image binding stays
   in its hash-verified provenance. It checks the authenticated actual Network
   version and applies the same loopback inform readback, without bootstrapping
   or starting the external controller.
2. The Core qualifier opens strict authenticated SSH forwarding from guest
   `127.0.0.1:8080` to the host endpoint, then starts its verified candidate.
3. `accepted.env` repeats the binding and records `ADOPTED=true`,
   `CONNECTED=true`, `TELEMETRY=true` after the API connection and browser
   speed/traffic checks. The Core qualifier reads its live process/status evidence
   and atomically writes `guest-complete.sha256` containing the input digest.
4. Core waits at most 60 seconds for that acknowledgement, removes its
   containers, networks, named volumes, orphans and owned image aliases, then
   atomically writes `result.env`. It repeats the binding and adds `RESULT`,
   `ADOPTED`, `CONNECTED`, `TELEMETRY` and `CLEANUP`.

Only a zero command exit with matching `RESULT=passed` and `CLEANUP=passed`
passes. Signals cancel the owned Docker client before cleanup. The caller must
wait for the command to exit. After an interrupted caller, use the same input
with `make integration-tier3-candidate-controller-cleanup`; it requires the
matching run ownership digest. Never reuse a run ID or operate another run's
Compose project. There is no keep-running option.
With schema 2, Core's `CLEANUP=passed` covers only its helper project and owned
image aliases; `result.env` also contains
`CONTROLLER_CLEANUP=external-owner-required`. Labs must separately verify its controller runtime, data and
endpoint-tunnel cleanup before reporting complete controller acceptance.
