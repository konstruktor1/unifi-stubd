# Build, Checks, and Delivery

The root `Makefile` is the canonical build and quality interface. This page
maps its entry points, runtime prerequisites and provider handoffs.

Load the once-defined context through the existing native project/session owner
before host orchestration or direct provider calls. Make validates and inherits
it for those goals:

```sh
: "${PROVIDER_CONSUMER_ROOT:?Load the native project context}"
: "${PROVIDER_CONSUMER_CONTEXT:?Load the native consumer context}"
BUILD_RUNNER_ENTRY="$(jq -er '.command | select(endswith("/bin/provider")) | sub("/provider$"; "/build-runner")' .consumer-data/connections/build-runner/management.json)" || exit "$?"
export BUILD_RUNNER_ENTRY
```

The CI owner supplies this context once and passes it unchanged to later steps. Current
sources come from native bindings; no installed snapshot, provider.json, commit
or tag selects them. CT's CLI is its declared current `bin/ct`.
Its service MCP uses the
reviewed external management launcher saved in
`.consumer-data/connections/ct/management.json`.

Go payload targets run inside Build Runner's `/workspace` with `GOCACHE` and the
existing workspace guard. They need no host provider connection or receiving
context injected into the container. The native setup owner must prepare
`tests/` and disposable `tests/.work/` as private mode 0700 directories for the
shared verification-storage owner. This is a local filesystem prerequisite;
Git does not record directory modes.

## Consumer entry points

All agent-run Go work uses the configured Build Runner as described in
[`AGENTS.md`](../../AGENTS.md). Do not invoke host Go or a Go-based Make target
as a fallback.

| Outcome | Entry point |
| --- | --- |
| Complete quality gate from the natively configured consumer on the host | `make verify` |
| Read and check the current native BR integration context | `make provider-check` |
| Run CT in its saved check environment | `make quality-ct` |
| Inspect local setup and report unexecuted product coverage | `make doctor` |
| Runner-local product subgate (not a complete pre-merge gate) | `make check` |
| Controller-facing gate (bound Tier 3 run or Stable/RC matrix) | `make integration-nonunifi-metrics` |
| Linux packages | `make package` |
| FreeBSD package set (compatibility-only; blocked as evidence) | `make package-freebsd-pkg-repos` |
| Static package repositories | `make package-repos` |

`make verify` first runs `make provider-check`, then installs locked project
JavaScript dependencies and runs ESLint and Prettier with the declared `node24`
service. It runs `make check vulncheck`
in the `go` service, including gofmt, `go vet`, golangci-lint, repository policy,
configuration/profile validation, tests and vulnerability scanning. Then
`make quality-ct` uses `scripts/check-ct.py` to read CT's current contract through
management MCP and honor its saved `build-runner` execution choice. The discovered
BR `runner` executes CT's current public CLI in the ephemeral `ct-quality`
instance for all selected Go, JavaScript, Bash and Python checks.
The official SDK client uses the contained project environment at
`.consumer-data/dependencies/unifi-stubd/mcp/`, declared by
`requirements/unifi-stubd/mcp.txt`, and the saved current management connection.
The CT check toolchain comes from the maintained quality image, including
Go, Node, Python, Ruff, ShellCheck and shfmt; host Go or Node is not required.
Missing tools or an unavailable provider runtime fail the gate. Package and
controller acceptance remain additional checks when required by the change.

`build-runner/instances/ct-quality.yaml` binds the current CT source read-only at
`/opt/codetool` using `CT_SOURCE`. The native BR connection derives
`BUILD_RUNNER_HOST_PATH_CT_SOURCE` from the saved CT source binding; retain that
generated input when preparing another runner. The image's maintained source is
`build-runner/project-setup/ct-quality/Dockerfile`. `make quality-ct-image` uses
native Docker build because BR exposes no consumer Dockerfile build operation;
check execution remains with the public BR instance runner. The manifest
validates the physical `/workspace` Root/cwd, read-only CT source and contained
HOME, dependencies, caches and temporary paths. A missing binding, image or
validation failure stops the check without a host-native fallback.

`golangci-lint.mod` and `golangci-lint.sum` isolate the linter dependency graph
from the product module. Both require Go 1.27 or newer. The Makefile invokes
the tool module inside the declared Build Runner Go runtime.
Its Go commands use the owned package scope in `GO_PACKAGES`; installed
dependencies under `node_modules` are outside that scope.

Current BR runtime selection requires fresh originating host CPU and actual
effective image-platform observations for each Docker invocation. Docker daemon
architecture is a separate observation. Preserve those identities and the
declared runtime owner's validation; do not substitute a generic runner or infer
the host from the daemon or image.

## Open provider handoffs

Earlier gate and compiler observations remain in the
[verification record](backlog.md#current-verification). The ongoing current-file
integration, observed checks and pending owners are tracked in the
[engineering backlog](../engineering/backlog.md).
Each language's required tools run in its declared environment.
The native MCP transport, dependencies and operation data remain inside the
consumer root.

`make provider-check` validates the saved BR binding against its public fresh
connection export, reads all current required MCP resources under an acknowledged
active catalog/annotations/list subscription, calls annotations and queries the
optional `runner` / `operator-targets` persistence guidance. The inquiry creates
no target store or saved selection. The existing `scripts/check-ct.py` owner uses
`scripts/provider_context.py` for shared native binding and SDK context handling
for BR and CT. Its CT gate checks provider ID, consumer root and configuration
against the current contract.
Closed logical calls project the discovered direct name/arguments schemas and forward
failed checks to Make. The client closes the connection on completion or interruption,
waiting for the provider to confirm shutdown. CI uses this same `make verify`
entry. Its [provider setup](../../.github/actions/provider-setup/action.yml)
prepares current independent BR/CT sources, the native client and management
connections for the exact checkout; the selected CT image is prepared through
the existing Make owner. Local success does not establish those prerequisites
or a successful run on a GitHub-hosted worker.

The same check owner serves `annotations`, `doctor` and five exact own skill
resources through `.consumer-data/connections/unifi-stubd/annotations.json`.
The saved native `--mcp-annotations` entry remains valid. Native function,
annotations, update, Doctor and delivery skills live at `.agents/skills/unifi-stubd*/`;
their resources and list changes support SDK subscriptions and reconnect after
actual byte/mode changes. This terminal surface declares no extra PP catalog or
provider contract. AGENTS.md and the
[local integration guidance](../engineering/provider-integration.md) own the
implementation, storage and onward duties.

For `provider-update`, explicit BR or CT updates, applicable annotations, subscribed
resource-change hints and confirmed current-source changes, select the own
`unifi-stubd-update` skill. Each signal selects the same complete authorized local
adoption, verification, own-notice and onward workflow described in the local
guidance. This is a skill instruction, with no new updater command or automatic
Doctor, downstream-project or publication action.

The seven current CT skills are retained as exact foreign resources under
`.agents/skills/ct*/`. During authorized CT integration, its public `init` asks
only missing initial preferences and saves complete confirmed answers in this
consumer's `ct.toml`; saved choices reopen only on request. The current choices
preserve existing formatter/linter conventions and disabled optional i18n,
select `build-runner` for CT checks, and enable a seven-day rule-source review.
The selected check environment is separate from Python/MCP runtime selection.
Its setup owner must finish `owner_work` and `first_check`, including actual
tool discovery, formatter checks and namespace/containment proof. A saved BR
choice is not a passed check and permits no native fallback; observed setup
results remain in the engineering backlog.

Rule-source review is due work reported during CT calls, with no background job.
Preserve CLI `ct-source-review:` stderr and successful MCP diagnostics for the
calling agent. When due, use the current `source-review` work order, read actual
primary sources and complete authorized corrections/checks before saving its
declared result and verifying a fresh due-state read. The
[local integration guidance](../engineering/provider-integration.md) owns the
concrete persistence and onward method. Native context initialization must
propagate failure and stop before exports or dependent work.

Select Doctor only for requested local diagnosis. `make doctor` uses inspection.
Its diagnostic command intentionally exits 1 with incomplete coverage, so Make
reports a failed target: required product-runtime acceptance has not run.
The configured own Python also supports
`scripts/check-ct.py --doctor`, with optional `--mode inspect|verify`,
`--scope setup`, `--consumer-role` and `--phase` values from its help/schema.
Explicit verification runs the existing repository policy check, BR identity
and CT contract procedures. It runs no workload or controller and preserves
packaged YAML semantic validation, Go/JavaScript checks and controller acceptance
as unexecuted coverage. This diagnostic result is separate from `make verify`.

`make integration-nonunifi-metrics` uses the consumer-owned
`build-runner/instances/controller-gate.yaml` through Build Runner. Its Tier 2
base is an immutable CircleCI `cimg/base` image with Bash, curl, jq, Make,
Docker CLI, Compose and Buildx. Validation checks those tools and Docker API
access before running the matrix; no tools are installed per job. Go builds
belong to the product and helper Dockerfiles, so this orchestration image does
not need Go.

The manifest explicitly grants Docker API access and the root identity needed
by the socket inside the orchestration container. Use it only with an isolated
lab daemon. The Make entry passes the physical
lab path for daemon-visible bind mounts; the default route requires that path
to be visible to the selected Docker endpoint. It does not infer a remote path.

The retained matrix payload resolves the
latest published Stable and Release Candidate at run start and
runs the same candidate in both disposable lanes. Details and acceptance
conditions live in the [Docker controller lab](docker-lab.md). Full product
acceptance additionally requires the focused host-safety and no-fabrication
evidence defined by the [requirements](requirements.md).
This describes retained fixture behavior. Controller-backed acceptance must use
the assigned Tier 3 machine; a missing binding leaves owner work unfinished and
does not authorize a core fixture or GitHub worker as a substitute.

## Automation ownership

- [CI](../../.github/workflows/ci.yml) owns `make verify`, coverage and the
  main-branch Linux package build on GitHub-hosted Ubuntu workers. Build Runner
  supplies the ephemeral toolchain containers; installation smoke uses a
  disposable GitHub-hosted worker. Checks run on repository-branch pushes and
  manual dispatch. Fork pull requests receive a blocking notice without executing
  their code; after review, a maintainer transfers the contribution to a
  repository branch.
- [Controller Integration](../../.github/workflows/tier2-docker.yml) owns the
  manually dispatched controller gate. GitHub hosts its orchestration job and
  prepares BR through the same provider setup. A hosted job does not establish
  access to a binding Tier 3 test machine, controller readiness or host-safety
  acceptance. Controller-backed work still requires the assigned machine,
  current companion inventory and the declared lab access path.
- [Package Repositories](../../.github/workflows/package-pages.yml) uses
  GitHub-hosted Ubuntu workers for Linux release artifacts and the package site.
- [FreeBSD Packaging](../../.github/workflows/freebsd-packaging.yml) alone uses
  the existing self-hosted `docker-dmz-unifi-stubd` runner for its manual
  compatibility build and temporary run artifacts. That role keeps its Docker
  daemon restriction and cannot publish.

### Provider sources on GitHub-hosted workers

[Provider setup](../../.github/actions/provider-setup/action.yml) obtains BR and
CT independently from their configured private Forgejo source locations into the
consumer's existing `.consumer-data/dependencies/` role. It reads the current
default branch rather than selecting a provider release, tag or historical
execution tree. Source transport does not replace the public launchers' current
artifact checks. No source is inferred from the checkout's cwd or a Mac path.

Private read access uses the GitHub Actions secret `PROVIDER_SOURCE_TOKEN` and
repository variables `PROVIDER_SOURCE_USERNAME`, `BUILD_RUNNER_SOURCE_URL` and
`CT_SOURCE_URL`. Source locations stay in native GitHub settings. Supply read-only access to both
provider repositories; keep credentials out of URLs, saved Git remotes,
connections, logs and committed files. The configured source URLs are native CI
inputs. The GitHub repository token grants no Forgejo access by itself.

The job defines the exact physical checkout once through
`PROVIDER_CONSUMER_ROOT` and its native `PROVIDER_CONSUMER_CONTEXT`; children
inherit them. The job supplies compatible native Python and the project MCP client,
then the existing preflight generates each saved management connection with that
source's public `bin/provider mcp --connection`, checks `help --json` and runs
`check --json`. BR retains the derived current CT source input needed by this
consumer's selected CT environment. Go, Node and the selected CT tools run in BR
containers. BR and CT remain independent functions.

Both repositories reject anonymous reads. The selected GitHub secret and username
variable are now configured. The read-only token is restricted to the two provider
repositories; a fresh disposable consumer successfully cloned both, installed the
SDK and generated native bindings through the actual preflight. The source owner
has published the verified CT-004/CT-006 corrections to its current default branch.
Repository delivery and an observed GitHub run complete CI-001; local tests and a
prepared workflow alone do not prove successful remote execution.
The native client compares retained CT instructions with current source resource
bytes and modes, so a mismatched delivery stops with its responsible owner action.
`make verify` also runs `make workflow-lint` in the Go service for workflow and
local-action metadata validation.

### Historical runner observation

The read-only GitHub observation on 2026-10-28 found no provider launcher
variables and one online FreeBSD runner without `docker-engine`. The earlier
proposal to require a Docker-enabled self-hosted worker for general CI was too
broad and is superseded by the role split above. The FreeBSD runner is not to be
relabelled or given Docker access for these jobs. Earlier green workflow runs do
not verify the new provider bootstrap.

The administration inventory's `unifi-stubd-gh-runner` deployment on the
`docker` guest describes that existing FreeBSD role; it is not a prerequisite
for GitHub-hosted general checks. The direct FreeBSD builder remains an
external-machine compatibility path without Build Runner evidence or a
publication route.

Release decisions and operator steps are documented in the
[development workflow](development.md) and [release checklist](release.md).
