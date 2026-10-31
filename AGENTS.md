# Agent Instructions

## Project

`unifi-stubd` is a Go lab tool that makes a Linux host, VM, or bridge appear as
a minimal UniFi switch to a UniFi Network Controller. It is experimental,
unofficial, and intended only for isolated lab or management networks.

The safety boundary is important: the controller must not blindly provision the
host. Adoption commands can be accepted and persisted, but controller-triggered
restart, upgrade, shell, or host-networking changes must stay explicit,
lab-scoped, and reviewable.

## First Places To Read

- `README.md`: user-facing overview, quick start, services, packaging.
- `packaging/linux/etc/unifi-stubd/config.yaml`: packaged Linux service config.
- `lab/stub/`: project-owned controller integration fixtures.
- `docs/en/architecture.md`: architecture and component boundaries.
- `docs/en/protocol-notes.md`: discovery, inform, and adoption protocol notes.
- `CREDITS.md`: research sources, license boundaries, and attribution rules.
- `SECURITY.md`: lab-data and vulnerability reporting policy.

## Known Agent Support

`AGENTS.md` is the canonical instruction file. Compatibility bridge files are
thin adapters for common tools and should not duplicate project rules:

- Claude Code: `CLAUDE.md`
- Gemini CLI: `GEMINI.md`
- GitHub Copilot: `.github/copilot-instructions.md` and `.github/instructions/`
- Cursor: `.cursor/rules/unifi-stubd.mdc`
- Windsurf: `.windsurf/rules/unifi-stubd.md`
- Cline: `.clinerules/unifi-stubd.md`
- Roo Code legacy setups: `.roo/rules/unifi-stubd.md`
- Aider: `.aider.conf.yml` and `CONVENTIONS.md`

## Build And Validation

Use the root `Makefile` as the canonical build interface:

```sh
make verify
make check
make lint
make test
make package
```

`make verify` is the complete quality gate invoked from the prepared consumer
on the host. Product checks use Build Runner's `go` and `node24` services.
The Make client reads CT through native management MCP and runs its complete
selected checks in the confirmed `ct-quality` Build Runner instance, with current
CT source mounted read-only. Host Go and Node are not prerequisites.
`make check`, `lint`, `test`, and `package` are Go payload targets inside the
Build Runner workspace. See `docs/en/ci-cd.md` for runtime prerequisites,
provider preflight and unresolved controller handoffs.
`GO_PACKAGES` in the Makefile defines the owned root-module scope so installed
dependencies are not treated as project Go packages.

Invoke existing Make targets and public provider commands directly. Do not
create task-local wrapper scripts or write task scripts, logs or review evidence
under `/tmp` or `/private/tmp`. Use captured tool output; persist required run
artifacts only under the project's `dist/` or the exact state directory supplied
by the responsible provider. This does not change declared runtime caches or
the separate lab-data and deployment paths.

The controller integration entry point is `make integration-nonunifi-metrics`.
Its payload owns one disposable Compose project per run and must remove its
containers, networks, named volumes, and orphans before returning. Do not add
fixed container names, persistent restart policies, or a keep-running bypass.

The repository uses:

- Go modules and a root `go.work`.
- Go `1.27` as the product's current stable minor-version floor.
  `golangci-lint.mod` keeps the linter dependencies
  separate; Build Runner supplies the compiler.
- Go tool directives for `golangci-lint` and `nFPM`.
- YAML service configuration under `/etc/unifi-stubd/config.yaml`.
- Packaged Linux file sources under `packaging/linux/`.

## Repository Layout

- `cmd/unifi-stubd/`: command entry point and CLI wiring.
- `internal/config/`: YAML config loading and CLI override model.
- `internal/device/`: fake switch profiles, link speed, and payload data.
- `internal/discovery/`: UniFi discovery TLV packet building.
- `internal/inform/`: inform packet framing, padding, and client logic.
- `internal/adoption/`: adoption state and environment persistence.
- `internal/adoptionssh/`: built-in SSH shim for advanced adoption.
- `internal/adapters/linuxbridge/`: Linux bridge/FDB helpers.
- `tests/`: all Go tests. Do not add `_test.go` files under `internal/`.
- `packaging/linux/`: Linux service units and packaged config sources.
- `packaging/nfpm.yaml`: Debian, RPM, and Arch Linux package metadata.
- `scripts/`: policy and packaging scripts.
- `docs/en/` and `docs/de/`: user-facing docs in English and German.
- `lab/stub/`: the only lab runtime kept in this core repository; firmware
  research, QEMU/VM runtimes, and other Tier 3 assets belong to the companion
  project `unifi-stubd-labs`.

## Companion Data And Tier 3 Test Machines

- On this workstation, the project-owned companion data root is
  `/Volumes/ssd/unifi-stubd-data/`. Treat it as part of the overall
  `unifi-stubd` project context even though its contents stay outside this core
  repository. The volume name is only a filesystem path; do not use "SSD" as
  project terminology.
- `/Volumes/ssd/unifi-stubd-data/tier3/machines/` contains the binding test
  machines for controller-backed product tests. Use these test hosts when a
  test requires a controller; do not substitute a core `lab/stub` fixture, a
  generic runner host, or a production controller.
- `/Volumes/ssd/unifi-stubd-data/tier3/artifacts/` and
  `/Volumes/ssd/unifi-stubd-data/tier3/migration/` contain machine artifacts
  and migrations. They are not the authoritative test-machine inventory.
- Before Tier 3 work, read the companion runner documentation and the machine
  definitions under `tier3/machines/`; do not reconstruct their roles from
  memory.

## Coding Rules

- Follow the existing Go style and package boundaries.
- Keep exported Go identifiers documented in English.
- Keep user-facing docs in English and German when the change affects users.
- Keep tests under `tests/`, grouped by package behavior.
- Keep generated artifacts under `dist/` out of commits.
- Keep firmware images, VM disks, extracted firmware trees, and Tier 3 runtime
  state out of this core repository. They belong to `unifi-stubd-labs` and its
  separately managed data storage.
- Do not commit PCAPs, adoption keys, controller API tokens, private controller
  URLs, real lab addresses, SSH host keys, MAC tables, or client data.
- Use documentation example addresses such as `192.0.2.0/24` instead of real
  site addresses.

## Protocol And Research Boundaries

`CREDITS.md` lists public docs and reverse-engineering projects used for
research. Some research repositories do not publish a license. Treat those as
idea and protocol references only; do not copy source code from them.

If a future change copies code or structured data from any external project,
update `CREDITS.md`, `NOTICE.md`, and the license decision before merging.

## Safe Behavior Expectations

- Discovery and inform traffic should remain opt-in and lab-scoped.
- Adoption SSH behavior should emulate only the minimal commands needed for
  advanced adoption.
- Do not execute arbitrary shell commands received from a controller.
- Do not mutate host networking based on controller provisioning data unless a
  future design explicitly adds a reviewed, local-only adapter.
- Prefer deterministic profile data over controller-derived guesses.


## Current external-provider integration

Use the current Build Runner and CT sources through this project's saved
`.consumer-data/connections/<provider-id>/management.json` bindings. Generate
them with their existing public `bin/provider mcp --connection` owner; native
configuration owns source, interpreter, cwd and the once-defined receiving ENV.
Preserve unrelated settings, domain data and literal domain argv. BR and CT are
independently selected direct functions; neither is the other's prerequisite.

Use this project's own function, annotations, update, Doctor and delivery skills and
native annotations connection. That terminal surface supplies read-only notices,
explicit local diagnosis and all five own skill resources with SDK subscriptions;
it creates no full domain provider or PP catalog binding. The seven retained
`ct*` skills at `.agents/skills/` and `requirements/ct/consumer-integration.md`
remain exact foreign texts. Resolve current CT and BR instructions through their
discovered MCP resources, including canonical annotations and the stable PP catalog.

The current integration outcomes, physical role map and concrete remaining
work belong to `docs/engineering/backlog.md` and the own function skill.
No commit, staging, publication or domain operation is implied. The own
annotations skill defines unavailable-root recovery versus confirmed file
invalidation; preserve and pass this same distinction to later receivers.

## Codetool provider maintenance

Use the current stable Codetool launcher or connected MCP for the exact
canonical consumer Git root. Use `PROVIDER_CONSUMER_ROOT` defined once by its
native project/session context and inherited unchanged by children; every write,
child runtime, dependency, cache, transport, log and temporary file stays inside
it. Upstream sources are read-only. Native configuration and resource declarations
do not grant filesystem authority.

For initial consumer setup, select the current CT function and delivery skills
and call `init` through this consumer's native binding. Inspect its actual
`environment` check plan with the project agent, ask missing confirmed choices
including `check_environment` (`native` or `build-runner`), and save only this
consumer's preferences. This choice remains separate from `[python]` and the MCP
runtime. Reuse saved choices; reopen them only on request. Complete the selected
`check-environment` owner work and `first_check` instruction with its literal
context through existing native setup/connection owners. Resolve the current
contract before code changes, apply actual native formatting, run real required
checks and resolve semantic work. An explicitly selected BR path requires its
current subscribed function context and a supported native projection of the same
physical consumer. Verify context, active local Root inheritance in CT children,
contained mutable paths and current read-only CT source. A validated `/workspace`
mapping is derived OS data, not another caller root; path-string difference alone
proves no defect. Reject caller overrides and physical escapes. Observe actual
tools/checks; missing proof or capabilities remain with their owner.
Add no CT-to-BR callback, automatic activation or native fallback. A saved choice,
ready plan, instruction or empty check list establishes no working first check.
Pass the same per-consumer setup and observed-completion duty onward, including
to terminals; preserve other consumers and foreign sources.

Select the own `ct-doctor` skill only for diagnosis explicitly requested by the
user within this project. Setup, change, delivery and completion are phases,
never automatic diagnosis/repeat triggers. Use actual CT observations and
existing check owners, preserving original checkpoint owners, applicability,
coverage limits and pending behavior. Inspection performs no settings, binding,
source, report or backlog writes. Discover actual callable availability, without
a guessed CLI or unimplemented tool claim. CT's current `setup` scope uses native
configuration/init inspection and explicit contract/tool-PLAN/persistence
plausibility; those procedures run no quality/application code, BR calls,
provisioning or save. Productive Doctor owners bind suitable declared local
checks through existing owners and may have contained runtime effects while
preserving settings/domain data and original process outcomes. Small generic
development smokes remain separate and fully cleaned, without limiting productive
diagnosis or becoming a test engine.
Every C/CC MUST implement its own meaningful Doctor function and selectable skill
using actual own functions, configuration and local plausibility checks; PP
results do not replace that specialization. Every owner MUST carry both this duty
and the foreign-project boundary through its applicable contract, function,
annotations, Doctor and delivery skills and work instructions to every next C/CC,
including terminals, and require that receiver to repeat the same onward
obligation. Terminals use existing native callable owners and local guidance
without extra provider-contract/publication duties. Never visit, inspect, test,
modify or leave verification artifacts in another real project from here; provide
its owner instructions only on explicit request. Authorized own Doctor diagnosis
remains permitted. Retain local reports and follow-up through existing owners.

Before affected work, retrieve the current provider MCP instruction with the
authorized operation and known `consumer_role`. Read each `reading.required`
artifact completely through `resources/read`, deduplicating exact URIs; read
deferred implementation before using it. Resolve the current own function and
annotations skills through that contract. Apply the inherited `request_gate`:
observe the provider ID and exact current artifact bytes/modes, reuse fully
read context while those contents/modes, boundary, configuration and scope match, and reconnect a
changed source before a confirmed-safe new call. Reconcile unknown original
outcomes with the existing function/process owner before retrying. Functions have
no registration, accepted/seen identity, per-call subscription gate or acceptance
acknowledgement prerequisite; optional version labels do not select execution.

A confirmed source-owner announcement of a concrete current function or artifact
change creates mandatory authorized local implementation and verification work.
Reconcile the complete current direct PP shared package, including its dependency
declaration, applicable requirements, affected own declarations and actual
implementation. Verify matching owner/path/bytes/modes; unchanged wording, empty
notices and earlier work never waive an actually changed artifact. Reading or
forwarding alone is not adoption; source-control events alone create no update duty.
For each relevant onward effect, every owner MUST maintain a currently necessary
notice in its own canonical annotations skill and return that same notice through
callable annotations. State the affected own function, receiving applicability,
practical effect, actual local implementation status or concrete gap and necessary
next-owner action. The regular chain is PP to its directly bound C, then C's own
annotations to CC, repeated at each boundary; change delivery requires no extra
direct PP connection or unsolicited cross-project delivery. Invent no notice or
recipient without onward impact, and keep no reader tracking or notice history.
Every owner MUST implement, verify and pass this same adoption, own-annotation
and onward duty through its contract, function, annotations and delivery skills
and work instructions to every next receiver, including terminals through existing
native/local equivalents, requiring that receiver to repeat it. Missing authority,
current source or support leaves concrete pending work with its owner and next
action. Notices grant no authority, automatic maintenance or version parity.
Preserve inherited bytes/owners, physical boundaries and unknown original outcomes;
provide another real owner's instructions only on explicit request and never
visit, inspect, test or change that project under this project's authority.

The existing native client/connection owner MUST use the current catalog's
`mcp.protocol_version` and verify its required SDK/client capabilities. Protocol
selection is technical responsibility, never a user preference or function input.
Do not ask the user to choose a version, invent aliases, silently downgrade or
omit required capabilities; older SDK support alone proves no PP compatibility.
Missing support stays pending with the concrete unsupported requirement, native
client owner and implementation/update action. Every onward owner MUST pass this
same responsibility through its own contract, function, annotations and delivery
skills and work instructions to each next receiver, including native/local
terminal equivalents. Authorization concerns a concrete change, never version choice.

Every receiving MCP client, including terminals, MUST use `subscriptions/listen`
after discovery for the discovered stable catalog, own canonical annotations
resource and resource-list changes. Observe the SDK acknowledgement, keep the
stream active while using the connection and renew it after every reconnection.
Report unsupported protocol/client subscription as a concrete blocker and
implement supported client integration within the receiving owner's authority.
On a change hint, reconcile unknown original outcomes, reconnect and read complete
current context before a confirmed-safe new call. Add no polling or persistent
delivery/seen/acceptance records. Every onward owner passes this same subscription
and handling duty through its own contract, skills and work instructions to every
receiving stage, including terminals.

Define canonical `PROVIDER_CONSUMER_ROOT` once in each receiving project's native
project/session environment. Children of the same operation inherit it unchanged;
each subsequent consumer defines its own context. Add no second root, alias,
per-call override or cwd/PATH fallback.

Construct only the closed logical call
`{"provider":"<declared provider ID>","name":"<declared function name>","arguments":{}}`.
Resolve provider to exactly one current native connection through its existing
configuration owner; use the discovered function name and input schema. Preserve
literal function/user data. Add no executable, root, cwd, env, transport or
caller-built command/argv fields or absolute execution selectors.

The existing native connection/function owner alone projects that call to
standard MCP `tools/call` name/arguments or its declared literal CLI argv. It
derives common root-valued inputs, executable, cwd and environment from the one
receiving context and validated source/interpreter/binding owners. Generated
absolute native values are context-bound materializations, not another root
authority; regenerate/rebind them when that context or source binding changes.
Use `provider_contract.management` verbatim: `bin/provider`; PP retains
`bin/provider-provider`. Add no registry, router, wrapper, alias, internal bypass
or guessed JSON-to-argv mapping. Missing/ambiguous native context, connection or
declared projection leaves required work pending with its owner and next action.
Verify real supported CLI/MCP procedures, root separation and unchanged inputs;
prose alone proves no migration. Pass this same ENV ownership, exact call shape,
construction/projection and verification duty through own contracts, function
and annotations skills and work instructions to every receiver, including terminals.

PP owns mandatory common interfaces and inputs. Each function owner implements
its own commands and determines the required and optional own arguments its
behavior needs. Declare them in the existing contract and own function skill,
with current schemas for callable MCP inputs. Callers supply applicable common
and function-required inputs through existing validators; native owners alone
derive root-valued common inputs.
Each onward owner enforces and passes this same ownership boundary and canonical
method unchanged through its contracts, skills and work instructions to consumers,
including terminals. Correct affected own declarations and actual dispatch in the
authorized task and verify ordinary CLI/MCP calls.

Only the current native-bound files supply execution, resources and artifact
equality. Provider identity contains `id` only; requirement provenance contains
`provider_id`, `path` and `sha256`. Verify current bytes and modes through their
existing source owners. Identical artifacts remain equal regardless of Git,
commit or tag history. Remove owned revision selectors, historical execution
copies, approval/acceptance prerequisites and release-version execution or cache
bindings; contextual requirements are current read-only data, never fallback
implementations. Schema, protocol and functional toolchain versions retain their
technical roles. Every receiving owner MUST implement, verify and pass this same
current-file model unchanged through its own contracts, function and annotations
skills and work instructions to every subsequent consumer, including terminals.

Keep owned code and generated executable content portable: use paths relative to
their owner, external native configuration and validated physical-root derivation.
Remove hard-coded machine/project absolute paths; add no cwd/PATH fallback for
provider or project roots. Preserve physical checks and exact foreign/inherited
bytes. Each receiving owner implements and passes both duties through its own
contract, skills and work instructions to every subsequent consumer, including
terminals. Correct actual violations in the same authorized task; instruction
delivery alone is not implementation. Ordinary function dispatch performs no
maintenance automatically and does not waive the announced-change duty.

Before editing code, resolve `bin/ct contract . --json` or MCP
`contract`. Here `bin/ct` denotes the supplied current bound source's stable
launcher. Apply its `work_instruction`, full exact mandatory BASE and selected
language/style, paradigm, architecture, pattern and optional-policy criteria.
Native `ct.toml` owns selection; `--use-case` or MCP `use_case` selects
the task use case. Retain the actual source and policy/configuration binding,
scope, task base and outstanding duties in existing task context and pass the
relevant content to delegates. Restore lost context before affected work.
Changed configuration or scope requires the affected selection again.

The generating or editing agent MUST finish `work_instruction.completion` in
the calling consumer project during the authorized task: run applicable required
checks, review actual changes, correct concrete failures and mandatory-rule
violations, and rerun affected checks. Reuse unaffected valid evidence. Resolve
every `semantic_work.unresolved_checks` entry; missing prerequisites, authority or
necessary decisions leave a concrete unfinished owner action. Do not stop at
fixable findings or offer already authorized corrections as future work. Tool
availability, a plan and successful result finalization prove no semantic work.
For MCP, inspect domain `result.exit_code` and `result.result` as well as protocol
errors. Selected Anti-Bloat uses its current skill and one final task-diff review;
it adds no scheduled job or separate reviewer.

Use agent-neutral own skills and standard MCP only. Do not install, register or
require agent-specific hooks or automatic work triggered by session, edit,
compaction or delegation events. During authorized integration work, remove
proven obsolete owned hooks and bindings through their existing owner after
confirmed writer termination; preserve native settings, data and foreign content.
Pass this same method and removal duty through own contracts, skills and work
instructions to every receiver, including terminals. Canonical CLI, native skill
discovery/configuration and MCP stream/process lifecycle retain their owners;
ordinary calls do not start cleanup.

For an authorized integration update or repair, identify provably obsolete owned
integration artifacts, record concrete files, commands and changes in the local
backlog, and execute cleanup and implementation in the same task. Remove old
registration, handshakes, version coupling and contradictory instructions.
Implement own function skills, an annotations skill and callable annotations MCP,
plus necessary MCP behavior where skills are insufficient through existing
function owners. Preserve domain data, native settings, foreign artifacts and
exact inherited sources. Verify actual client skill selection, representative
workflows, real required MCP calls and applicable existing checks. Report
Removed/replaced, Implemented, Checked and Open with concrete changes, results
and blockers or none; finish implementation or name a substantiated blocker.
Pass this same work and reporting duty through own skills and work instructions
to subsequent consumers, including terminals, within their authorized projects.
Ordinary function dispatch performs no maintenance automatically and does not
waive the announced-change duty.

After functional work and its applicable checks, complete the catalog's
reorganisation before finishing the same authorized task. Classify own internal
engineering instructions/code/data, product and delegated content, including each
skill's owner and purpose. Separate maintained sources and generated product
deliverables from generated task instructions/helpers/data/fixtures/evidence.
Keep engineering documentation and backlogs in `docs/engineering/`, maintained
engineering automation in `scripts/`, and generated task work under
`project_storage.evidence/<responsible-owner>/<task>/`, using `inputs/`, `scripts/`,
`results/`, `review/` and `release/` as needed. Retain native AGENTS/configuration
and skill-discovery anchors, stable product paths and foreign materialized paths;
use existing runtime, dependency, connection and verification storage owners.
Name actual role paths and own skill classifications in the implementation
supplement's Data and boundaries section, or existing terminal guidance. Preserve
useful historical evidence, native settings, domain data and exact foreign
bytes. Delete only proven obsolete owned artifacts after their process owner
resolves unknown outcomes and confirms writer termination. Record concrete
changes, verify affected procedures and report results or substantiated blockers.
Pass this same classification, layout, sequence and completion duty onward,
including terminals. Ordinary calls do not initiate reorganisation; add no
automatic migration or cleanup framework.

The current contract resolves `ct.consumer-integration`,
`ct.implementation`, own skills and applicable inherited sources. Keep
foreign requirements and the complete shared PP package unchanged. Independently
review current contents and observed function results before ordinary authorized
Git delivery. Publication retains its explicit authorization owner and adds no
separate acceptance or activation gate. Local implementation, actual checks,
delivery and receiving-consumer execution are separate observed outcomes.
<!-- Codetool provider maintenance end -->

## Build Runner receiving integration

Resolve current Build Runner guidance through the saved native
`.consumer-data/connections/build-runner/management.json` connection. Select its
runtime skill for workloads and its integration skill for an authorized local
update. Read complete current required resources, including all five supplied
PP skills, and deferred implementation before affected work. Keep acknowledged
catalog, canonical annotations and resource-list subscriptions active throughout
MCP use. The existing native connection owner regenerates derived settings with
the bound public `bin/provider mcp --connection`; Make and CI retain declared
literal domain CLI procedures. Neither BR nor CT becomes the other's prerequisite.

The own function, annotations, update, Doctor and delivery skills under
`.agents/skills/unifi-stubd*/` use the existing check owner. Current local
procedures, persistence and physical role paths are maintained in
`docs/engineering/provider-integration.md`. No provider contract, catalog registry
or publication step is required for this terminal consumer.

Use `provider-update` and the selectable own `unifi-stubd-update` skill for
explicit updates, applicable annotations, subscribed resource-change hints and
confirmed current-source changes. Each signal selects the same complete workflow:
establish affected functions, owner and authorization; resolve unknown original
outcomes; reconnect on invalidation; read complete current context with active
acknowledged subscriptions; reconcile the complete applicable shared package and
dependency declaration, requirements, own skills/guidance and actual behavior;
finish authorized local implementation, fitting checks and current-source review;
maintain relevant canonical own notices and verify their callable projection and
filters; finish the common backlog, then applicable reorganisation/cleanup and
its checks. Each notice states the affected own function, applicability, practical
effect, actual local status/gap and next-owner action promoting `provider-update`.
Reading or forwarding upstream success proves no local adoption. Keep only
necessary current notices without recipient tracking or history. A signal grants
no mutation, Doctor, publication or foreign-project authority.

Every next C/CC at every depth, including terminals, MUST implement its own
selectable update skill advertising `provider-update` and all applicable signals,
bind that same complete local workflow and repeat this own-skill, adoption,
annotation and onward duty through its contract, function, annotations, update
and delivery skills and work instructions. Terminals use native/local discovery
without a provider-contract or publication duty. Use the regular delivering
interface without a direct PP binding, new CLI/MCP entry, updater, hook, polling,
acceptance state or automatic downstream work. Reuse valid unchanged checks and
report observed no-change results; missing authority, support or proof names its
concrete pending owner/action. Preserve both recursive Doctor and foreign-project
duties below.

The native client/connection owner selects the retained catalog's
`mcp.protocol_version` through the existing SDK, verifies negotiated protocol and
required capabilities including subscriptions, and passes that technical duty
onward. Do not ask the user to select a protocol, silently downgrade or omit
required behavior. Missing support is concrete pending work with its client owner.
Keep dependencies acyclic; during authorized maintenance remove only proved unused
own peer bindings, mandatory peer checks or delegated executable copies through
existing owners after writer termination. Preserve actual independent direct
function needs: BR supplies workload runtimes; CT selects the quality checks,
which run in this consumer's saved Build Runner environment.

Managed definitions and references stay project-owned; Docker image/cache bytes
and VM payloads retain their selected native backend resource and cleanup owner.
Shared invocation runtime, dependencies, connections, sessions, control records
and evidence retain their declared physical project roles. Before a new durable
purpose or changed location, state the backend's actual proposal, purpose,
constraints and cleanup owner; offer that proposal preselected plus supported
project/custom locations and wait for the answer. Reuse explicit existing native
choices without repeatedly asking for the same purpose/location. Follow the
existing resource/configuration owner and any applicable persistence inquiry,
authorized save, fresh read and actual use. Preserve existing data in place; add
no universal data root, new store, implicit move/export or migration authority.
Missing supported placement remains a concrete owner/action gap. Pass this same
per-purpose choice and native ownership method through own function and delivery
guidance to every next receiver, including terminals; it creates no additional
`PROVIDER_CONSUMER_ROOT` or external filesystem authority.

Every receiving owner at every C/CC depth, including terminals, MUST implement
its own meaningful domain Doctor function and selectable skill using actual own
functions, configuration and safe local plausibility checks. Upstream or PP
success is not local proof. Select Doctor only for explicitly requested local
diagnosis; inspect is the default and verify must be explicit. Setup, changes,
delivery and completion are descriptive phases, never automatic triggers. Do not
visit, inspect, test, modify or leave artifacts in another real receiving project
under this project's authority; provide that owner's instructions only on explicit
request. Authorized diagnosis inside this project retains its local scope. Pass
both duties unchanged through own function, annotations, Doctor and delivery
skills and work instructions to every next receiver, requiring that receiver to
repeat the same onward obligation. Publishing receivers also use their existing
contracts; terminals use native/local equivalents without a publication duty.

For applicable persistence, use the selected function's current declared inquiry,
purpose, optional/required condition and save trigger. Preserve authoritative
native stores and verify selected result, authorized save, fresh read and actual
application. BR's optional `runner` / `operator-targets` purpose retains
`build-runner/targets.toml` and its original owner; one-off input requests no save.
Own annotations and Doctor require no application store or report writer. Select
the own delivery skill only for explicitly requested next-owner instructions,
passing the same persistence method, complete applicable current resources,
Doctor/foreign-project duties and onward obligation without operating that
receiver. Values, native bindings, reports and populated backlogs remain local.

Maintain actual local work in `docs/engineering/backlog.md` with the current PP
seven-column format, declared statuses and priorities; order active work by
dependencies and keep completed work separate. During authorized integration
work, replace superseded owned instructions and proven obsolete owned artifacts,
preserving exact foreign bytes, native settings, domain data and unknown original
outcomes. Finish actual implementation, applicable checks and the final selected
task review, then perform and verify the declared artifact reorganisation before
completion. Report Removed/replaced, Implemented, Checked and Open with concrete
results and responsible-owner actions. Ordinary calls perform no cleanup,
reorganisation, diagnosis or publication automatically. Pass this same method,
classification, layout, sequence and completion duty onward, including terminals.
