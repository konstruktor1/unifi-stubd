# Native provider integration

This is maintained local engineering guidance for the `unifi-stubd` terminal
consumer. `AGENTS.md` owns project authority and lab safety, the root `Makefile`
owns product verification, and the current source providers own their functional
contracts. This project creates no provider contract, registry or publication
requirement merely to consume those functions.

## Current context and invocation

The native project/session owner defines one canonical physical
`PROVIDER_CONSUMER_ROOT` and `PROVIDER_CONSUMER_CONTEXT`. Children inherit them
unchanged. The latter contains `consumer_id`, `persistence` and `resources`, with
resource fields `id`, `owner`, `path` and `providers`. These data describe the
receiving context and resource owners; they confer no filesystem authority.

Saved connections are `.consumer-data/connections/build-runner/management.json`
and `.consumer-data/connections/ct/management.json`. The existing native
configuration owner validates their physical current command, interpreter,
consumer cwd and ENV against the active context, then obtains current derived
settings through the bound public `bin/provider mcp --connection`. A saved root
does not override the active receiving root. Rebind changed derived settings
through that owner; retain unrelated native settings and foreign entries.

Construct only `provider`, `name` and `arguments` in a logical call. The exact
provider selects its one saved native connection; discovered current schemas
own function names and inputs. The native owner alone projects standard MCP
`tools/call` or the provider's declared literal CLI procedure. BR's `runner`
requires `runner` and literal `argv` within `arguments`; the selected runner's
parser owns its further domain inputs. Add no caller root, command string,
execution selector, alias, registry, router, wrapper or guessed JSON-to-argv
mapping. Preserve the relative public management entry `bin/provider` and PP's
distinct `bin/provider-provider` entry. Derived absolute native paths retain
physical validation and are not another root authority.

BR and CT are independently selected functions. BR supplies declared runtimes;
CT supplies selected code-quality criteria and checks. CT's saved execution
preference belongs to this consumer, not a CT-to-BR callback or provider
prerequisite. Native Make/CI
domain CLI dispatch retains its existing owner and exact literal arguments.
`scripts/provider_context.py` supplies shared current-context reading and native
binding validation for the existing local check owner, without becoming another
domain executor or public startup interface.
Keep these actual direct dependencies acyclic. During authorized integration
maintenance, remove proved unused own peer bindings, mandatory peer checks and
delegated executable copies through their existing owners after confirming writer
termination. Preserve original foreign content, unrelated settings and required
direct function needs. BR's retirement of its own CT dependency does not remove
this project's independent CT quality function.
The native context owner propagates initialization failure and stops before
evaluating exports or starting dependent work. The existing CI exporter assignment
under `set -euo pipefail` and native binding reader's checked subprocess retain
that boundary; a failed initializer is never usable context. Pass the same duty
through own skills and onward instructions.

## GitHub CI roles and source preparation

General quality and Linux packaging jobs use GitHub-hosted Ubuntu workers with
BR's ephemeral containers. The existing self-hosted runner serves only manual
FreeBSD packaging and retains its Docker-daemon restriction. A GitHub-hosted
controller orchestration job still needs the assigned Tier 3 machine and declared
lab access; scheduling that job proves no controller availability.

The maintained `.github/actions/provider-setup/action.yml` prepares independent
current BR and CT source bindings under the existing project dependency owner,
then the existing preflight exports each native management connection. Native
job context defines the exact physical `PROVIDER_CONSUMER_ROOT` and
`PROVIDER_CONSUMER_CONTEXT` once; children inherit them. Sources, MCP client
support and generated connections keep their declared private roles. Current
source/default-branch acquisition is transport, never a release/commit execution
selector or a reason to reuse a historical tree. Neither BR nor CT becomes the
other's source-delivery prerequisite.

On disposable GitHub-hosted Ubuntu workers, provider setup and host orchestration
use the same root UID as the BR controller. Setup first validates the native
physical root, then transfers ownership of that job checkout before creating
private provider state. The private `0700` checks remain in force. This setup
rejects self-hosted workers and does not apply to the FreeBSD runner.

Private Forgejo reads use `PROVIDER_SOURCE_TOKEN` from GitHub Actions secrets
and `vars.PROVIDER_SOURCE_USERNAME`; `BUILD_RUNNER_SOURCE_URL` and `CT_SOURCE_URL`
are configured source URLs in native GitHub repository variables.
Credentials stay out of saved source URLs, connections, evidence and committed
files. Both repositories reject anonymous reads; the selected GitHub secret and
username variable are now configured with repository-restricted read access.
Actual fresh clone/SDK/public-export/preflight checks passed in a disposable
native consumer. The source owner has published the verified CT-004/CT-006
corrections to its current default branch. Source publication and repository
delivery retain their explicit owners; workflow preparation alone implies neither.
The native context reader compares retained CT instruction bytes/modes against
the subscribed current source, so mismatches stop with their owner action rather
than selecting old instructions or modifying foreign content. A fresh setup
derives the optional CT_SOURCE host reference from the saved CT connection
through the existing preflight before public BR export. BR-only jobs do not
require CT. `make workflow-lint` is the Go payload for the workflow/action checks
in the complete gate.
Track those concrete actions in CI-001 and claim remote success only after an
actual current-source CI run. Local verification remains a separate result.

## Reading, subscriptions and recovery

The native SDK client gives the initial typed `server/discover` request five
minutes for a cold public-launcher bootstrap. The SDK stamps the already selected
catalog protocol and client capabilities; the observed server capabilities and
acknowledged subscriptions remain required. This changes no workload timeout,
retry policy or protocol selection.

Select the source's runtime skill for BR workloads and integration skill for
authorized local integration changes. Through the current saved connection,
discover resources and functions and retrieve `get_instructions` for the
authorized operation and known role `terminal`. Read every `reading.required`
resource completely through `resources/read`, deduplicating exact URIs and
checking provider ID, byte length, SHA-256 and mode. All five supplied PP skills
remain required. The local client also reads the complete declared shared PP
package, including its dependency declaration, and the current implementation
supplement through the same integrity reader. It installs no executable source
copy. Read other deferred implementation before affected source use.
Pointers locate content and permit no fragment-only reading.

The native client/connection owner selects the retained catalog's
`mcp.protocol_version` through the existing SDK and verifies negotiated protocol,
required client/server capabilities and subscription support. This is technical
implementation work; ask no user protocol-choice question, silently downgrade or
infer compatibility from older SDK support. Missing support names the responsible
client owner and concrete implementation/update action. Recheck the discovered
catalog against the selected current protocol before dependent calls and pass
this same duty through every onward instruction surface, including terminals.

After discovery, subscribe with SDK `subscriptions/listen` to the discovered
stable catalog bound by `provider-provider.catalog`, the provider's canonical
annotations resource and resource-list changes. Observe the SDK acknowledgement
and keep the stream active during connection use. Renew after reconnection.
Read complete current guidance and use declared schemas; file presence,
resource retrieval and a connection export alone prove no function behavior.

Reuse completely read context only while relevant current bytes/modes, boundary,
configuration and scope remain unchanged. A hint or lost stream stops dependent
calls. Preserve and resolve unknown original outcomes with the existing
function/process owner before reconnecting and reading complete affected context
for a confirmed-safe new call. Never replay automatically, poll for acceptance,
hot-swap source or retain a seen/accepted ledger. Missing protocol or subscription
support is a concrete receiving-client blocker.

For the own surface, an unavailable native root reports unknown availability
without hints or a stale latch; restoring the root permits the same session.
An unavailable or unsafe required file under an accessible root invalidates the
session, publishes subscribed resource/list hints and stays stale after restore.
That confirmed file invalidation requires reconnect and complete affected
reading. Preserve this distinction through all onward guidance.

Only current native-bound files supply execution and resources. Provider identity
contains `id` only; requirement provenance contains `provider_id`, `path` and
`sha256`. Current bytes, modes and actual build inputs establish equality/cache
integrity; Git events, release labels and history select no execution or update
duty. A confirmed concrete announced artifact/function change requires authorized
receiving implementation and verification of the complete applicable current
package, dependency declaration, requirements and own affected declarations.
Verify already equal artifacts; empty notices and earlier work waive no actual
change. Read-only source access and announcements grant no foreign writes.

## CT choices and first check

The seven direct foreign CT skills are `ct`, `ct-annotations`, `ct-anti-bloat`,
`ct-i18n`, `ct-doctor`, `ct-delivery` and `ct-update`. They and applicable
`requirements/ct/` resources retain their exact current descriptors, bytes and
modes; do not rewrite them as project-owned guidance.

For authorized initial CT integration, select its current function/delivery
guidance and call public `init . --json` through this project's binding. Inspect
its actual code/configuration evidence and `environment` plan, then ask only the
returned missing choices. Recommendations and defaults are not answers. Existing
saved choices reopen only on explicit request. The local confirmed choices are
the existing formatter/linter conventions, disabled optional i18n policy,
`check_environment = "build-runner"` and enabled seven-day source review under
`maintenance.sources`; languages, procedural/modular-monolith/facade and selected
Anti-Bloat retain their existing choices. These values are this consumer's
settings, not defaults for a next receiver.

Before a save, query CT `persistence-instructions` for `function: init`,
`purpose: consumer-preferences`. Submit the fully confirmed answers through
`init`'s existing native writer; partial answers remain unsaved. Preserve unrelated
`ct.toml` keys/comments, `[python]`, native settings and formatter files. A fresh
contract/init read must show actual selection/application. The preference remains
separate from CT's Python interpreter and fixed management MCP runtime.

`preferences-saved` completes the dialog, not setup. Finish returned `owner_work`
and `first_check.instruction` with its literal `context.calls`. Obtain the current
contract before code edits, inspect effective native formatter configuration,
apply the selected real formatter to authorized code, and execute plan/required
checks in the chosen environment. Resolve actual failures and all applicable
`semantic_work.unresolved_checks`, then complete selected semantic/source work
and one final Anti-Bloat review. Preserve actual process outcomes and reuse valid
unchanged evidence. A saved choice, ready plan or empty check list proves no
working first check.

For the selected BR path, this consumer's native connection/runtime owner uses
current subscribed BR guidance and its supported literal runner procedure. CT
adds no callback, activation or host-native fallback. Before provisioning or
claiming success, verify that the actual namespace projection binds the same
physical consumer, preserves native context and keeps the active local
`PROVIDER_CONSUMER_ROOT` unchanged in CT children. A validated `/workspace`
mapping is derived runtime data; different host/container path strings alone
prove no violation. CT source stays current and read-only; runtime, dependencies,
HOME, caches and temporary files stay in the receiving boundary. Reject caller
root overrides and escapes. Missing execution capability or proof remains with
its concrete runtime/connection owner in the backlog.

Here `make quality-ct` invokes the existing `scripts/check-ct.py` owner. That
client reads CT's current contract through its native management connection and
retains successful diagnostics. Only the saved BR choice prepares the maintained
native-architecture image through `make quality-ct-image`, then projects execution
to the discovered BR `runner`. The supported native choice uses CT's native check
owner without preparing that image. The BR choice runs the current public
`/opt/codetool/bin/ct check .` with
`--use-case cli --require-tools --json` through
`build-runner/instances/ct-quality.yaml`. The existing BR execution owner maps
this same physical consumer to `/workspace`; its active Root remains unchanged
in CT children and the native consumer context is passed unchanged.

The maintained image input is
`build-runner/project-setup/ct-quality/Dockerfile`, with pinned toolchain/image
inputs for Go 1.27, Node 24, Python 3.15, Ruff, ShellCheck, shfmt and CT's native
dependencies. BR has no consumer Dockerfile build operation, so only this image
build uses the native Docker interface; execution remains with its public runner.
On cancellation, bounded process-group shutdown closes the owned Make/Docker
clients. It does not prove that the Docker daemon cancelled its build; reconcile
the original backend image/build outcome through its native owner before retrying.
The instance validates physical Root/cwd, required tools, the current CT source's
read-only mount and contained HOME/Go/npm/pip/cache/temp paths under
`.provider-provider/runtime/unifi-stubd/ct-quality/`. Images/layers/build cache
retain their selected native Docker storage owner.

The native BR connection carries `BUILD_RUNNER_HOST_PATH_CT_SOURCE`, derived from
the current saved CT public entry by the existing connection/export owner. Its
`CT_SOURCE` host reference supplies the manifest's read-only `/opt/codetool`
mount; never hard-code a workstation source path in the manifest or image. The
client rejects a mismatching source input instead of selecting a copy or native
fallback. A remote CI worker still needs its own correctly prepared connections,
derived source input and image build capability; local observations certify no
remote setup.

## CT rule-source review

Enabled source review becomes due during CT function invocation, without a timer,
scheduled job, event hook or automatic browsing. Preserve CLI `ct-source-review:`
stderr and successful MCP `result.diagnostics` and give due work to the calling
project agent; successful domain results do not discharge those diagnostics.
Missing initial choice asks once; disabled preference starts no review.

When due, read current subscribed CT context, resolve a fresh contract and call
`source-review` for the selected use case. Execute its `agent_task.instruction`
and literal context: browse actual applicable primary publications, assess their
derived guidance and complete authorized consumer corrections and fitting checks.
Foreign provider corrections remain with their source owner. A work order,
returned prompt, success flag or invocation proves no completed review.

Only after actual review, query `persistence-instructions` for
`source-review` / `source-review-result`. Copy the fresh nine-field
`result_template`: `schema_version`, `kind`, `project`, `rules_digest`,
`reviewed_at`, `sources`, `findings`, `checks` and
`checks_not_applicable_reason`. The current template is schema 2. Preserve its
project/rules binding; use the actual timezone-aware completion time and exactly
one `reviewed` or `not-applicable`
disposition/observation for each returned source URL. Do not hide inaccessible required knowledge as
inapplicable. Record concrete findings, actual check references and a reason if
checks are inapplicable. Save only that object through the declared bounded
exclusive native 0600 writer at `review_result_path`; reconcile unknown writes
before retrying. A fresh `source-review` call and later ordinary calls must show
the completed review's actual interval effect. An accepted record alone proves
no semantic review.

Every next receiver repeats this same initial-once preference, native save/read/
application, actual selected-environment first-check and invocation-based review
method through its own contract or native equivalent, function, annotations,
update and delivery skills and work instructions. Keep values and populated
reviews local. Provide another real receiver's instructions only on explicit
request; perform none of its setup, review or tests from here.

### Review binding across validated namespaces

Schema 2 uses the existing native `PROVIDER_CONSUMER_CONTEXT.consumer_id` label
in the result's `project` field solely to distinguish the review scope. The work
order's `project` remains the validated physical project path. The label grants
no Root, source, provider or filesystem authority. The native context owner uses
distinct labels for independent review scopes and independently validates each
same-consumer namespace projection; a copied record with a duplicate label proves
no physical identity. The private record owner still reads only inside the active
validated Root.

The current rule digest binds complete selected criterion content, style and use
case, excluding derived full-text navigation and launcher locators. The host Root
and `/workspace`, and their current public launcher paths, may differ after the
native owner validates their same-consumer mapping. Different review labels,
criteria, style or use case invalidate a saved result. Missing or invalid labels
cannot complete an enabled review; missing or disabled preferences gain no gate.

Schema 1 is invalid and has no compatibility reader or automatic migration. For
a proven binding-only correction, obtain a fresh schema-2 template and the current
inquiry. Verify unchanged criteria, style, use case and reviewed primary-source
URLs against the actual earlier review. Preserve the earlier observations and
actual completion timestamp, archive the old bytes as task evidence, and safely
replace the single authoritative record through its native owner. Changed
guidance requires the applicable actual source review. Never advance the clock,
create a second store or override the Root to make a result valid. Verify fresh
native and real BR readback of the same record and due interval before claiming
completion. The current adoption and observed checks are tracked as CT-004 in
the local backlog and under
`.consumer-data/evidence/unifi-stubd/ct-precision-cleanup-20261030/`.

Carry this project's known `cli` selection into CT auxiliary calls through their
current declared `use_case` input, including persistence inquiries and all three
Anti-Bloat commands. CLI uses `--use-case=cli` at the actual subcommand; MCP uses
the literal argument in its current schema. Annotations and requested CT Doctor
propagate that selection to their existing inquiries. Omission keeps default
`library` and remembers no prior call; another actual selection still invalidates
the saved review. Add no preference, Root/ENV selector, extra record or clock
rewrite. Every next owner passes the same known task context through its own
current native projection and verifies the actual diagnostic behavior.

## Own functions and diagnosis

`scripts/check-ct.py` remains the local check and own MCP owner. Its saved
`.consumer-data/connections/unifi-stubd/annotations.json` connection supplies
`annotations`, `doctor` and five exact own skill resources. The canonical annotations
skill is the sole current notice source, with an optional nonempty function
filter. `scripts/integration_doctor.py` owns the local Doctor specialization.
Discover actual input schemas through `tools/list` and use native CLI help for
its supported local projection; neither name creates another provider.

Select Doctor only for explicitly requested diagnosis in this project. Common
optional inputs are `mode` (default `inspect`, explicit `verify`), `scope`
(`setup`), `consumer_role` and `phase`; the function owner applies omitted
defaults once after validation. Null, empty, unknown and extra inputs remain
invalid. Setup, changes, delivery and completion describe phases, never automatic
diagnosis or repair triggers.

Inspection safely observes current native bindings, own integration and packaged
YAML anchors, CT TOML syntax and product JSON schema format. Explicit verification
runs the existing repository policy check, native BR identity and CT contract
procedures through their owners. It preserves settings, domain data and lab
safety. It starts no controller, firmware/deployment action,
network mutation, regression suite or automatic repair. Contained support runtime
effects retain their existing provider/process owner. Own configuration and
plausibility observations are required; forwarded PP or BR success cannot
replace them. Packaged YAML semantic validation, Go/JavaScript checks and
controller acceptance remain required unexecuted coverage, so this setup report
remains incomplete. Full product acceptance retains its actual owner.

Return the common scope/findings/coverage/outcome meanings. Keep original
checkpoint owners, meaningful current evidence, supported failures and required
unexecuted work. `not_applicable` requires positive current exclusion; missing
inputs remain `not_checked`, unresolved applicability remains explicit. Required
unexecuted/unknown observations make coverage incomplete; any supported failure
keeps outcome failed independently of coverage. A pass requires complete
mandatory coverage and a meaningful applicable positive observation. Metadata
inspection certifies no unexecuted workload. Valid negative findings are ordinary
results; source invalidation, cancellation and unknown original outcomes retain
their existing error channels. Diagnosis writes no report or backlog.

Every C/CC at every depth, including terminals, MUST implement its own meaningful
domain Doctor function and selectable skill through actual own functions,
configuration and safe local checks. Do not visit, inspect, test, modify or leave
artifacts in another real receiving project under this project's authority;
provide that owner's instructions only on explicit request. Authorized own
diagnosis remains permitted. Pass both duties unchanged through own function,
annotations, Doctor and delivery skills and work instructions to every next
receiver, requiring the same onward obligation there. Publishing receivers also
use their contracts; terminals use native/local equivalents without gaining a
provider-contract or publication duty. Small disposable development smokes remain
separate and fully cleaned; they do not limit meaningful productive diagnosis.

## Provider update

`provider-update` selects `.agents/skills/unifi-stubd-update/SKILL.md` through
native skill discovery. Explicit updates, applicable annotations, subscribed
resource-change hints and confirmed current-source changes all select that same
complete local workflow, never a read-only or notice-only shortcut. A signal
grants no mutation, Doctor, foreign-project or publication authority. This
terminal needs no provider contract, extra direct PP connection, new CLI/MCP
entry, updater, hook, polling or acceptance state.

Establish affected functions, owner and authorization; resolve unknown original
outcomes, reconnect on invalidation and read complete applicable current context
under the request gate with active acknowledged subscriptions. Reconcile the
complete current package and dependency declaration, applicable requirements,
own skills/guidance and actual implementation through existing owners. Preserve
exact inherited artifacts, native settings, domain data and literal inputs.
Record concrete obsolete owned paths/commands and intended changes in the local
backlog, then finish authorized implementation and applicable existing checks.
Use the current CT contract for affected code, actual native skill selection and
required CLI/MCP behavior, including the maintained integration verifier where
affected. Independently review current changes and observed results; reuse valid
unchanged evidence and report observed no-change results without fabricated edits.
Doctor remains explicit diagnosis only. Missing authority, support or proof names
its concrete pending owner/action.

Assess each own exposed function and delegated duty. Maintain only necessary
current notices in the canonical own annotations skill and verify the same
callable `annotations` projection and filters. Each notice states the affected
own function, receiving applicability, practical effect, actual local status/gap
and next-owner action promoting `provider-update`. Upstream success and forwarding
prove no local adoption. Remove irrelevant notices without reader tracking or
announcement history. Follow the regular provider-to-receiver chain; no automatic
downstream work or unsolicited foreign-project access follows.

Every next C/CC at every depth, including terminals, MUST implement its own
selectable update skill advertising `provider-update` and the applicable signals,
bind the actual complete local process, and repeat this same duty through its
contract, function, annotations, update and delivery skills and work instructions.
Terminals use native/local equivalents without publication. Preserve both
recursive meaningful own Doctor and foreign-project duties. Finish the common
backlog, then applicable reorganisation/cleanup and its checks after functional
work/checks. Report local adoption, checks, own notices, pending actions,
separately authorized publication and actual downstream execution distinctly.
The update guidance itself creates no application store or settings. Affected
functions retain their declared persistence inquiries, save triggers, native
owners and fresh-read/application duties, including CT preferences and actual
source-review results. Ordinary calls acquire no empty-store prerequisite.

## Runtime data placement

Project-owned manifests, definitions, build inputs and integration references do
not impose their location on Docker image/layer/cache bytes or VM payloads. Those
retain the selected daemon, builder, VM or hypervisor's native resource and cleanup
owner. Existing shared integration runtime, dependencies, connections, sessions,
control records and task evidence retain their physical project roles. Backend
placement creates no additional `PROVIDER_CONSUMER_ROOT` or general external
filesystem authority.

Before a new durable purpose or changed location, state its purpose, executing
backend, actual current storage proposal, retention/cleanup owner and constraints.
Offer that proposal preselected alongside supported project/custom locations and
wait for the answer; preselection is not authorization. Reuse an explicit existing
native choice without asking again for the same purpose/location. Missing or
ambiguous native configuration is not a prior choice. If no concrete proposal or
supported chosen placement exists, name its backend owner and pending capability.

Apply the answer through the existing native configuration/resource owner. Use
the declared persistence inquiry when applicable, save only on its authorized
trigger, then verify a fresh read and actual application. A one-off choice
requests no new settings store. Preserve existing data in place; add no universal
data-root variable, move, export or rewrite of backend storage without separately
authorized work through that owner's interface. Docker's actual selected
daemon/builder store remains its proposal; VM durable roots and hypervisor
datastores retain their existing selectors. Optional exports are separately
requested copies. Pass this same per-purpose choice, preservation and fresh-read/
application method through own function and delivery guidance to every next
receiver, including terminals.

## Persistence and onward work

Own verification, annotations and Doctor need no persistent application store.
Native connection settings retain their existing configuration owner. No report
writer, empty-store, backlog or acceptance record is a prerequisite to calls.

BR's optional `runner` / `operator-targets` purpose is queried through its current
`persistence-instructions` function with required `function` and `purpose` inputs.
Its concrete response owns useful selected values, native
`build-runner/targets.toml` schema, authorized save/update/reset, current reader,
precedence and conflict handling. A one-off input requests no save. Reading
creates no default file. Preserve unrelated keys, secret ownership and invalid
data for the responsible owner's correction. Explicit inputs override saved
values, then declared defaults apply. Verify an authorized selected result,
save, fresh read and actual application through existing owners; successful
storage alone proves no domain effect. Never repeat a completed domain action
solely to retry persistence; reconcile unknown writes before safe retry.

Select the own delivery skill only for explicitly requested next-owner
instructions. Carry complete applicable current guidance/resources and exact
shared artifacts through their existing owners, preserving foreign IDs, paths,
bytes, modes and propagation. Do not copy this terminal's native values, reports
or populated backlog into delegated product guidance. Each receiver implements
its own function, annotations, Doctor and delivery skills and repeats the same
Doctor/foreign-project, invocation, subscription, persistence, verification and
onward duties. Delivery performs no automatic maintenance or publication and
proves no receiving execution.

The current PP catalog owns the common seven backlog columns: ID; Work and
intended use; Status; Responsible owner; Completion criterion / observed outcome;
Remarks and dependencies; Use cases. Statuses are Open (decision unresolved),
Not started, In progress, Blocked (concrete missing owner/input/authority and next
action) and Complete (observed work and applicable checks). P0 covers minimum
prerequisite decisions/declarations; P1 covers complete implementation, onward
duties, checks and review. Maintain actual work in
`docs/engineering/backlog.md`, with stable local IDs, active work ordered by
priority/dependencies and completed work separate. Pass this format and meaning
onward; populated records stay internal.

## Data and boundaries

| Role | Actual owner/path | Classification and boundary |
| --- | --- | --- |
| Product source and deliverables | `cmd/`, `internal/`, `tests/`, packaging/lab sources; generated product output in `dist/` | Stable product paths and existing Make owners; no provider migration of product behavior. |
| Native anchors | `AGENTS.md`, Makefile, native settings and `.agents/skills/` | Preserved native authority/discovery paths. |
| Own function skill | `.agents/skills/unifi-stubd/SKILL.md` | Project-owned functional guidance for development, verification and authorized integration. |
| Own annotations skill | `.agents/skills/unifi-stubd-annotations/SKILL.md` | Project-owned canonical functional notices and callable projection source. |
| Own update skill | `.agents/skills/unifi-stubd-update/SKILL.md` | Project-owned complete local provider-update workflow and recursive selectable-skill duty; native discovery anchor and functional guidance. |
| Own Doctor skill | `.agents/skills/unifi-stubd-doctor/SKILL.md` | Project-owned functional selection, local diagnosis and honest coverage guidance. |
| Own delivery skill | `.agents/skills/unifi-stubd-delivery/SKILL.md` | Project-owned requested onward guidance, persistence/backlog method and recursive duties. |
| Foreign functional guidance | Seven `.agents/skills/ct*/` entrypoints and applicable `requirements/ct/`; current BR/PP resources from native sources | Exact foreign bytes/modes remain with their source owner; task captures are read-only evidence, never fallback execution. |
| Maintained engineering | `docs/engineering/` and `scripts/` | Internal guidance/backlog and reviewed automation, including `check-ct.py`, `provider_context.py`, `integration_doctor.py` and `check-provider-integration.py`; no generated task wrappers. |
| Runtime and MCP SDK | `.provider-provider/runtime/`, specifically `.provider-provider/runtime/mcp-venv` for the shared SDK and `.provider-provider/runtime/unifi-stubd/ct-quality/` for contained CT quality runtime data | Current provider/runtime owners; no historical source selection or reassignment to generic dependency storage. |
| Own client dependency | `.consumer-data/dependencies/unifi-stubd/mcp/`, declared in `requirements/unifi-stubd/mcp.txt` | Existing contained own client environment. Other dependencies retain their existing owners. |
| Native CI setup | `.github/workflows/`, `.github/actions/provider-setup/`, `.github/actions/provider-preflight/` and `scripts/prepare-ci-providers.sh` | Maintained native job/source configuration and setup automation; only FreeBSD selects the existing self-hosted worker. GitHub environment/output channels retain their platform protocol owner; they introduce no provider Root authority. |
| Ephemeral CI provider sources | `.consumer-data/dependencies/ci-providers/<provider-id>/` | Current read-only source checkouts for one hosted job; Git preserves source modes inside private ancestors, public exporters own connections, and native job teardown owns removal. No historical source cache or release selector. |
| CI setup cache and temporary data | `.consumer-data/dependencies/unifi-stubd/pip-cache/` and `.provider-provider/runtime/unifi-stubd/ci-setup/tmp/` | Existing dependency/runtime owners, physically inside the receiving job; no new durable backend or cross-project root. |
| Preserved historical dependency data | `.consumer-data/dependencies/code-checker/service-mcp/`, retained `home` child | Retained offline Go modules and profile/configuration data, classified by the earlier current-file cleanup; no current executable integration, source selector or fallback. Preserve these useful data in place rather than deleting by the retired directory name. |
| Native connections | `.consumer-data/connections/<provider-id>/` | Generated context-bound configuration through public native exporters; no second root authority. |
| Domain settings | Consumer manifests under `build-runner/` and optional `build-runner/targets.toml` | Domain configuration/data retained by actual readers/writers; no automatic creation or migration. |
| CT quality runtime definition | `build-runner/instances/ct-quality.yaml` and `build-runner/project-setup/ct-quality/Dockerfile` | Maintained project-owned execution/build inputs at stable domain paths; current CT source remains a native-derived read-only mount, generated image bytes retain the Docker backend owner. |
| CT consumer choices and rule review | `ct.toml`; `.consumer-data/evidence/ct/source-review/` under the current inquiry/work-order owner | Native preferences and actual review result remain consumer-local; no scheduler or redundant settings store. |
| Generated task work | `.consumer-data/evidence/unifi-stubd/<task>/{inputs,results,review,release}/` | Private task inputs, observations and review, separate from maintained source and product output. |
| Disposable verification | `tests/.work/` | Bounded generic fixtures/data/runtime removed after confirmed child termination on success and failure; never another real receiver. |

Private storage directories use mode 0700 and authored evidence mode 0600.
Shared invocation writes, runtime, dependencies, caches, children and temporary
files remain physically inside the one receiving root in their declared roles.
Backend image/cache bytes and VM payloads retain the separately selected native
resource owner under Runtime data placement; project definitions never relocate
them. Existing physical owners reject escapes, unsafe links, mounts and aliases
before effects. Provider sources remain read-only; keep owned
executable paths relative to their owner or derive validated native locations.
Do not add cwd/PATH root fallbacks or author machine-specific absolute paths.

## Completion and reorganisation

Record concrete obsolete owned artifacts and intended changes in the local
backlog, then complete authorized implementation and applicable checks. Remove
old owned registration, handshakes, version coupling and contradictory guidance;
preserve native settings, domain data, exact foreign artifacts and unresolved
original outcomes. Delete only proven obsolete owned files after their process
owner confirms writer termination.

Obtain the current CT contract before affected code edits and apply all selected
mandatory rules. Use real format and semantic checks, resolve concrete failures
and all applicable `semantic_work.unresolved_checks`, inspect actual authorized
changes and rerun affected checks. Reuse unchanged valid evidence. Selected
Anti-Bloat uses one final task-diff review through its current work-order owner;
dirty unrelated work requires an explicit task-path selection. Do not mask a
missing required tool or proof with exclusions or a successful plan.

After functional work and its applicable checks, perform the declared artifact
reorganisation and fitting checks before completing that same task. Keep
engineering docs/backlog in `docs/engineering/`, maintained automation in
`scripts/` and generated task work in the responsible evidence task roles.
Preserve native anchors, stable product paths, exact foreign materialized paths
and useful historical evidence. Record actual moves/cleanup and results in the
backlog. Add no migration hook, automatic cleanup framework or persistent
delivery/acceptance ledger. Pass this classification, layout, ordering,
verification and completion duty unchanged through every later receiving stage.

Report Removed/replaced, Implemented, Checked and Open with concrete changes,
observed results and a responsible owner's substantiated next action, or none.
Implementation, actual checks, independent review, explicitly authorized
publication and receiving execution are separate outcomes. Ordinary function
calls start none of the integration maintenance or publication automatically.
